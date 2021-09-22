// Package sitemap generates an list of pages of the website.
// The generated XML file is stored on the site's root and can be parsed by search engine bots.
package sitemap

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/Defacto2/df2/lib/database"
)

const (
	resource = "https://defacto2.net/f/"
	static   = "https://defacto2.net/"
	index    = "index.cfm"
	cfm      = ".cfm"
	urlCount = 28
	// limit the number of urls as permitted by Bing and Google search engines.
	limit = 50000
)

// url composes the <url> tag in the sitemap.
type url struct {
	Location string `xml:"loc,omitempty"`
	// optional attributes
	LastModified string `xml:"lastmod,omitempty"`
	ChangeFreq   string `xml:"changefreq,omitempty"`
	Priority     string `xml:"priority,omitempty"`
}

// URLset is a sitemap XML template.
type URLset struct {
	XMLName xml.Name `xml:"urlset,omitempty"`
	XMLNS   string   `xml:"xmlns,attr,omitempty"`
	Urls    []url    `xml:"url,omitempty"`
}

// Create generates and prints the sitemap.
func Create() error {
	// query
	id, v := "", &URLset{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	var createdat, updatedat sql.NullString
	count, err := nullsDeleteAt()
	if err != nil {
		return err
	}
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query("SELECT `id`,`createdat`,`updatedat` FROM `files` WHERE `deletedat` IS NULL")
	if err != nil {
		return fmt.Errorf("create db query: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("create db rows: %w", rows.Err())
	}
	defer rows.Close()
	// handle static urls
	v.Urls = make([]url, len(paths())+count)
	c, i := v.staticURLs()
	// handle query results.
	for rows.Next() {
		i++
		if err = rows.Scan(&id, &createdat, &updatedat); err != nil {
			return fmt.Errorf("create rows next: %w", err)
		}
		// check for valid createdat and updatedat entries.
		if _, err = updatedat.Value(); err != nil {
			continue
		}
		if _, err = createdat.Value(); err != nil {
			continue
		}
		lmv := lastmodValue(createdat, updatedat)
		v.Urls[i] = url{fmt.Sprintf("%v%v", resource, database.ObfuscateParam(id)), lmv, "", ""}
		c++
		if c >= limit {
			break
		}
	}
	// trim empty urls so they're not included in the xml.
	empty, trimmed := url{}, []url{}
	for i, x := range v.Urls {
		if x == empty {
			trimmed = v.Urls[0:i]
			break
		}
	}
	v.Urls = trimmed
	output, err := xml.MarshalIndent(v, "", "")
	if err != nil {
		return fmt.Errorf("create xml marshal indent: %w", err)
	}
	if _, err := os.Stdout.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("create stdout xml header: %w", err)
	}
	output = append(output, []byte("\n")...)
	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("create stdout: %w", err)
	}
	return db.Close()
}

func nullsDeleteAt() (count int, err error) {
	dbc := database.Connect()
	defer dbc.Close()
	rowCnt, err := dbc.Query("SELECT COUNT(*) FROM `files` WHERE `deletedat` IS NULL")
	if err != nil {
		return count, fmt.Errorf("create count query: %w", err)
	}
	if rowCnt.Err() != nil {
		return count, fmt.Errorf("create count rows: %w", rowCnt.Err())
	}
	defer rowCnt.Close()
	for rowCnt.Next() {
		if err = rowCnt.Scan(&count); err != nil {
			return count, fmt.Errorf("create count scan: %w", err)
		}
	}
	return count, nil
}

// lastmodValue parse createdat and updatedat to use in the <lastmod> tag.
func lastmodValue(createdat, updatedat sql.NullString) string {
	var lm string
	if ok := updatedat.Valid; ok {
		lm = updatedat.String
	} else if ok := createdat.Valid; ok {
		lm = createdat.String
	}
	f := strings.Fields(lm)
	// NOTE: most search engines do not bother with the lastmod value so it could be removed to improve size.
	// blank by default; <lastmod> tag has `omitempty` set, so it won't display if no value is given.
	s := ""
	if len(f) > 0 {
		t := strings.Split(f[0], "T") // example value: 2020-04-06T20:51:36Z
		s = t[0]
	}
	return s
}

func (set *URLset) staticURLs() (c, i int) {
	// sitemap priorities
	const top, veryHigh, high, standard = "1", "0.9", "0.8", "0.7"
	uri := func(path string) string {
		return static + path
	}
	c, i, view := 0, 0, viper.GetString("directory.views")
	for i, path := range paths() {
		file := filepath.Join(view, path, index)
		if s, err := os.Stat(file); !os.IsNotExist(err) {
			set.Urls[i] = url{uri(path), lastmod(s), "", veryHigh}
			c++
			continue
		}
		j := filepath.Join(view, path) + cfm
		if s, err := os.Stat(j); !os.IsNotExist(err) {
			set.Urls[i] = url{uri(path), lastmod(s), "", high}
			c++
			continue
		}
		k := filepath.Join(view, strings.ReplaceAll(path, "-", "")+cfm)
		if s, err := os.Stat(k); !os.IsNotExist(err) {
			set.Urls[i] = url{uri(path), lastmod(s), "", standard}
			c++
			continue
		}
		set.Urls[i] = url{uri(path), "", "", top}
	}
	return c, i
}

func lastmod(s os.FileInfo) string {
	return s.ModTime().UTC().Format("2006-01-02")
}

func paths() [urlCount]string {
	return [urlCount]string{
		"welcome",
		"file",
		"file/list/new",
		"organisation/list/group",
		"organisation/list/bbs",
		"organisation/list/ftp",
		"organisation/list/magazine",
		"person/list/artists",
		"person/list/coders",
		"person/list/musicians",
		"person/list/writers",
		"search/result",
		"defacto2",
		"defacto2/donate",
		"defacto2/history",
		"defacto2/subculture",
		"contact",
		"commercial",
		"code",
		"help",
		"help/creative-commons",
		"help/privacy",
		"help/browser-support",
		"help/keyboard",
		"help/viruses",
		"help/allowed-uploads",
		"help/categories",
		"link/list",
	}
}
