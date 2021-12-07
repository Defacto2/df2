package stat

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
)

type Stat struct {
	Base      string          // Base is the relative path to file downloads which use UUID as filenames.
	BasePath  string          // BasePath to file downloads which use UUID as filenames.
	Columns   []string        // Column names.
	Count     int             // Count row index.
	Missing   int             // Missing UUID files count.
	Overwrite bool            // Overwrite flag (--overwrite) value.
	Total     int             // Total rows.
	Values    *[]sql.RawBytes // Values of the rows.
	start     time.Time       // processing time
}

func Init() Stat {
	dir := directories.Init(false)
	return Stat{Base: logs.Path(dir.UUID), BasePath: dir.UUID, start: time.Now()}
}

func (s *Stat) Summary(id string) {
	if id != "" && s.Total < 1 {
		return
	}
	total := s.Count - s.Missing
	if total == 0 {
		fmt.Print("nothing to do")
	}
	elapsed := time.Since(s.start).Seconds()
	t := fmt.Sprintf("Total proofs handled: %v, time elapsed %.1f seconds", total, elapsed)
	logs.Printf("\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}
