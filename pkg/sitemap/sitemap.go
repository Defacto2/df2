// Package sitemap generates an list of pages of the website.
// The generated XML file is stored on the site's root and can be parsed by search engine bots.
package sitemap

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/sitemap/internal/urlset"
)

var (
	ErrPointer = errors.New("pointer value cannot be nil")
)

const (
	// Location is the URL of the website.
	Location = "https://defacto2.net"

	// DockerLoc is the URL for the developer hosted on a Docker container.
	DockerLoc = "http://localhost:8560"

	// Namespace is the XML name space.
	Namespace = "http://www.sitemaps.org/schemas/sitemap/0.9"

	// limit the number of urls as permitted by Bing and Google search engines.
	Limit = 50000
)

// Create generates and prints the sitemap.
func Create(db *sql.DB, w io.Writer, dir string) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	id := ""
	tmpl := &urlset.Set{XMLNS: Namespace}
	var createdat, updatedat sql.NullString
	count, err := nullsDeleteAt(db)
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT `id`,`createdat`,`updatedat` FROM `files` WHERE `deletedat` IS NULL")
	if err != nil {
		return fmt.Errorf("create db query: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("create db rows: %w", rows.Err())
	}
	defer rows.Close()
	// handle static urls
	const paths = 29
	tmpl.URLs = make([]urlset.Tag, paths+count)
	c, i := tmpl.StaticURLs(dir)
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
		loc, err := url.JoinPath(Location, "f")
		if err != nil {
			return err
		}
		tmpl.URLs[i] = urlset.Tag{
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
	if err := createOutput(w, tmpl); err != nil {
		return err
	}
	return nil
}

func createOutput(w io.Writer, tmpl *urlset.Set) error {
	if tmpl == nil {
		return fmt.Errorf("urlset set %w", ErrPointer)
	}
	if w == nil {
		w = io.Discard
	}
	empty := urlset.Tag{}
	trimmed := []urlset.Tag{}
	for i, x := range tmpl.URLs {
		if x == empty {
			trimmed = tmpl.URLs[0:i]
			break
		}
	}
	tmpl.URLs = trimmed
	b, err := xml.MarshalIndent(tmpl, "", "")
	if err != nil {
		return fmt.Errorf("create xml marshal indent: %w", err)
	}
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("writer xml header: %w", err)
	}
	b = append(b, []byte("\n")...)
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("writer xml: %w", err)
	}
	return nil
}

func nullsDeleteAt(db *sql.DB) (int, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	count := 0
	if err := db.QueryRow("SELECT COUNT(*) FROM `files` WHERE `deletedat` IS NULL").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// lastmodValue parse createdat and updatedat to use in the <lastmod> tag.
func lastmodValue(createdat, updatedat sql.NullString) string {
	lm := ""
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
