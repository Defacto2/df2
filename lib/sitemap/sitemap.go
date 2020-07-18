package sitemap

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/viper"
)

const (
	resource string = "https://defacto2.net/f/"
	static   string = "https://defacto2.net/"
)

// limit the number of urls as permitted by Bing and Google search engines
const limit int = 50000

// url composes the <url> tag in the sitemap
type url struct {
	Location string `xml:"loc"`
	// optional attributes
	LastModified string `xml:"lastmod,omitempty"`
	ChangeFreq   string `xml:"changefreq,omitempty"`
	Priority     string `xml:"priority,omitempty"`
}

// Urlset is a sitemap XML template
type Urlset struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	Svs     []url    `xml:"url"`
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
func Create() {
	// query
	var id string
	var createdat, updatedat sql.NullString
	db := database.Connect()
	rows, err := db.Query("SELECT `id`,`createdat`,`updatedat` FROM `files` WHERE `deletedat` IS NULL")
	logs.Check(err)
	logs.Check(rows.Err())
	defer db.Close()

	v := &Urlset{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	c := 0

	// handle static urls
	d := viper.GetString("directory.views")
	for _, u := range urls {
		i := filepath.Join(d, u, "index.cfm")
		if s, err := os.Stat(i); !os.IsNotExist(err) {
			v.Svs = append(v.Svs, url{fmt.Sprintf("%v", static+u), lastmod(s), "", "0.9"})
			c++
			continue
		}
		j := filepath.Join(d, u) + ".cfm"
		if s, err := os.Stat(j); !os.IsNotExist(err) {
			v.Svs = append(v.Svs, url{fmt.Sprintf("%v", static+u), lastmod(s), "", "0.8"})
			c++
			continue
		}
		k := filepath.Join(d, strings.ReplaceAll(u, "-", "")+".cfm")
		if s, err := os.Stat(k); !os.IsNotExist(err) {
			v.Svs = append(v.Svs, url{fmt.Sprintf("%v", static+u), lastmod(s), "", "0.7"})
			c++
			continue
		}
		v.Svs = append(v.Svs, url{fmt.Sprintf("%v", static+u), "", "", "1"})
	}

	// handle query results
	for rows.Next() {
		err = rows.Scan(&id, &createdat, &updatedat)
		logs.Check(err)
		// check for valid createdat and updatedat entries
		_, errU := updatedat.Value()
		_, errC := createdat.Value()
		if errU != nil || errC != nil {
			continue // skip record (log in future?)
		}
		// parse createdat and updatedat to use in the <lastmod> tag
		var lastmod string
		if udValid := updatedat.Valid; udValid {
			lastmod = updatedat.String
		} else if cdValid := createdat.Valid; cdValid {
			lastmod = createdat.String
		}
		lastmodFields := strings.Fields(lastmod)
		// NOTE: most search engines do not bother with the lastmod value so it could be removed to improve size.
		var lastmodValue string // blank by default; <lastmod> tag has `omitempty` set, so it won't display if no value is given
		if len(lastmodFields) > 0 {
			t := strings.Split(lastmodFields[0], "T") // example value: 2020-04-06T20:51:36Z
			lastmodValue = t[0]
		}
		v.Svs = append(v.Svs, url{fmt.Sprintf("%v%v", resource, database.ObfuscateParam(id)), lastmodValue, "", ""})
		c++
		if c >= limit {
			break
		}
	}
	output, err := xml.MarshalIndent(v, "", "")
	logs.Check(err)
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
}

func lastmod(s os.FileInfo) string {
	return s.ModTime().UTC().Format("2006-01-02")
}
