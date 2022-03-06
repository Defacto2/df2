package scan

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logs"
)

// Stats contain the statistics of the archive scan.
type Stats struct {
	BasePath string          // Path to file downloads directory with UUID as filenames.
	Count    int             // Database table row index.
	Missing  int             // Missing UUID as files count.
	Total    int             // Total rows in the database table.
	Columns  []string        // Column names of the database table.
	Values   *[]sql.RawBytes // Values of the rows in the database.
	start    time.Time       // Processing duration.
}

// Init initializes the archive scan statistics.
func Init() Stats {
	dir := directories.Init(false)
	return Stats{BasePath: dir.UUID, start: time.Now()}
}

// Summary prints the number of archive scanned.
func (s *Stats) Summary() {
	total := s.Count - s.Missing
	if total == 0 {
		fmt.Print("nothing to do")
	}
	elapsed := time.Since(s.start).Seconds()
	t := fmt.Sprintf("Total archives scanned: %v, time elapsed %.1f seconds", total, elapsed)
	logs.Printf("\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}
