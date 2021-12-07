package stat

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
)

type Stats struct {
	start    time.Time       // processing time
	BasePath string          // path to file downloads with UUID as filenames
	Count    int             // row index
	Missing  int             // missing UUID files count
	Total    int             // total rows
	Columns  []string        // column names
	Values   *[]sql.RawBytes // row values
}

func Init() Stats {
	dir := directories.Init(false)
	return Stats{BasePath: dir.UUID, start: time.Now()}
}

func (s *Stats) Summary() {
	total := s.Count - s.Missing
	if total == 0 {
		fmt.Print("nothing to do")
	}
	elapsed := time.Since(s.start).Seconds()
	t := fmt.Sprintf("Total archives scanned: %v, time elapsed %.1f seconds", total, elapsed)
	logs.Printf("\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}
