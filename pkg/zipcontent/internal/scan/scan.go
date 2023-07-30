// Package scan initializes and stores statistics on a file archive.
package scan

import (
	"database/sql"
	"io"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/str"
)

// Stats contain the statistics of the archive scan.
type Stats struct {
	BasePath string          // BasePath is the path to the file downloads directory.
	Count    int             // Count the database table row index.
	Missing  int             // Missing UUID as files count.
	Total    int             // Total rows in the database table.
	Columns  []string        // Columns are column names from the database table.
	Values   *[]sql.RawBytes // Values of the database table.
	start    time.Time       // Processing duration.
}

// Init initializes the archive scan statistics.
func Init(cfg conf.Config) (Stats, error) {
	dir, err := directories.Init(cfg, false)
	if err != nil {
		return Stats{}, err
	}
	return Stats{BasePath: dir.UUID, start: time.Now()}, nil
}

// Summary prints the number of archive scanned.
func (s *Stats) Summary(w io.Writer) {
	if w == nil {
		w = io.Discard
	}
	count := s.Count - s.Missing
	str.Total(w, count, "archives")
	str.TimeTaken(w, time.Since(s.start).Seconds())
}
