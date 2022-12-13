// Package sitemap generates an list of pages of the website.
// The generated XML file is stored on the site's root and can be parsed by search engine bots.
package sitemap

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/sitemap/internal/urlset"
)

const (
	// Base URL of the website.
	Base = "https://defacto2.net"
	// Local base URL for the websote hosted on a Docker container.
	LocalBase = "http://localhost:8560"
	// limit the number of urls as permitted by Bing and Google search engines.
	Limit = 50000
)

// Create generates and prints the sitemap.
func Create() error {
	// query
	id, v := "", &urlset.Set{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
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
	v.Urls = make([]urlset.Tag, len(urlset.Paths())+count)
	c, i := v.StaticURLs()
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
		loc, err := url.JoinPath(Base, "f")
		if err != nil {
			return err
		}
		v.Urls[i] = urlset.Tag{
			Location:     loc,
			LastModified: database.ObfuscateParam(id),
			ChangeFreq:   lastmodValue(createdat, updatedat),
			Priority:     "",
		}
		c++
		if c >= Limit {
			break
		}
	}
	if err := createOutput(v); err != nil {
		return err
	}
	return db.Close()
}

func createOutput(v *urlset.Set) error {
	empty, trimmed := urlset.Tag{}, []urlset.Tag{}
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
	return nil
}

func nullsDeleteAt() (int, error) {
	db := database.Connect()
	defer db.Close()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM `files` WHERE `deletedat` IS NULL").Scan(&count); err != nil {
		return 0, err
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
	// NOTE: most search engines do not bother with the lastmod value so it could be removed to improve size.
	// blank by default; <lastmod> tag has `omitempty` set, so it won't display if no value is given.
	s := ""
	if f := strings.Fields(lm); len(f) > 0 {
		t := strings.Split(f[0], "T") // example value: 2020-04-06T20:51:36Z
		s = t[0]
	}
	return s
}
