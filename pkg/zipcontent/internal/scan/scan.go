package scan

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/directories"
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
	if count == 0 {
		fmt.Fprint(w, "nothing to do")
	}
	elapsed := time.Since(s.start).Seconds()
	t := fmt.Sprintf("Total archives scanned: %v, time elapsed %.1f seconds", count, elapsed)
	fmt.Fprintf(w, "\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}
