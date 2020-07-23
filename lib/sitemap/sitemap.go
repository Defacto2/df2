package sitemap

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/spf13/viper"
)

const (
	resource = "https://defacto2.net/f/"
	static   = "https://defacto2.net/"
	index    = "index.cfm"
	cfm      = ".cfm"
)

// limit the number of urls as permitted by Bing and Google search engines.
const limit = 50000

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
	Svs     []url    `xml:"url,omitempty"`
}

var urls = [28]string{
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

// Create generates and prints the sitemap.
func Create() error {
	// query
	id, v, count := "", &URLset{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}, 0
	var createdat, updatedat sql.NullString
	dbc := database.Connect()
	defer dbc.Close()
	rowCnt, err := dbc.Query("SELECT COUNT(*) FROM `files` WHERE `deletedat` IS NULL")
	if err != nil {
		return err
	} else if rowCnt.Err() != nil {
		return rowCnt.Err()
	}
	defer rowCnt.Close()
	for rowCnt.Next() {
		if err := rowCnt.Scan(&count); err != nil {
			return err
		}
	}
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query("SELECT `id`,`createdat`,`updatedat` FROM `files` WHERE `deletedat` IS NULL")
	if err != nil {
		return err
	} else if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	c, i := 0, 0
	// handle static urls
	d := viper.GetString("directory.views")
	v.Svs = make([]url, len(urls)+count)
	for i, u := range urls {
		file := filepath.Join(d, u, index)
		if s, err := os.Stat(file); !os.IsNotExist(err) {
			v.Svs[i] = url{fmt.Sprintf("%v", static+u), lastmod(s), "", "0.9"}
			c++
			continue
		}
		j := filepath.Join(d, u) + cfm
		if s, err := os.Stat(j); !os.IsNotExist(err) {
			v.Svs[i] = url{fmt.Sprintf("%v", static+u), lastmod(s), "", "0.8"}
			c++
			continue
		}
		k := filepath.Join(d, strings.ReplaceAll(u, "-", "")+cfm)
		if s, err := os.Stat(k); !os.IsNotExist(err) {
			v.Svs[i] = url{fmt.Sprintf("%v", static+u), lastmod(s), "", "0.7"}
			c++
			continue
		}
		v.Svs[i] = url{fmt.Sprintf("%v", static+u), "", "", "1"}
	}
	// handle query results.
	for rows.Next() {
		i++
		if err = rows.Scan(&id, &createdat, &updatedat); err != nil {
			return err
		}
		// check for valid createdat and updatedat entries.
		if _, err := updatedat.Value(); err != nil {
			continue
		}
		if _, err := createdat.Value(); err != nil {
			continue
		}
		// parse createdat and updatedat to use in the <lastmod> tag.
		var lastmod string
		if ok := updatedat.Valid; ok {
			lastmod = updatedat.String
		} else if ok := createdat.Valid; ok {
			lastmod = createdat.String
		}
		lastmodFields := strings.Fields(lastmod)
		// NOTE: most search engines do not bother with the lastmod value so it could be removed to improve size.
		// blank by default; <lastmod> tag has `omitempty` set, so it won't display if no value is given.
		var lastmodValue string
		if len(lastmodFields) > 0 {
			t := strings.Split(lastmodFields[0], "T") // example value: 2020-04-06T20:51:36Z
			lastmodValue = t[0]
		}
		v.Svs[i] = url{fmt.Sprintf("%v%v", resource, database.ObfuscateParam(id)), lastmodValue, "", ""}
		c++
		if c >= limit {
			break
		}
	}
	// trim empty urls so they're not included in the xml.
	empty := url{}
	var trimmed []url
	for i, x := range v.Svs {
		if x == empty {
			trimmed = v.Svs[0:i]
			break
		}
	}
	v.Svs = trimmed
	output, err := xml.MarshalIndent(v, "", "")
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write([]byte(xml.Header)); err != nil {
		return err
	}
	if _, err := os.Stdout.Write(output); err != nil {
		return err
	}
	return db.Close()
}

func lastmod(s os.FileInfo) string {
	return s.ModTime().UTC().Format("2006-01-02")
}
