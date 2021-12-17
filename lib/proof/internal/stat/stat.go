package stat

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
)

// Proof data.
type Proof struct {
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

func Init() Proof {
	dir := directories.Init(false)
	return Proof{
		Base:     logs.Path(dir.UUID),
		BasePath: dir.UUID,
		start:    time.Now(),
	}
}

// Summary of the proofs.
func (s *Proof) Summary(id string) string {
	if s == nil {
		return ""
	}
	if id != "" && s.Total < 1 {
		return ""
	}
	total := s.Count - s.Missing
	if total == 0 {
		return "nothing to do\n"
	}
	elapsed := time.Since(s.start).Seconds()
	t := fmt.Sprintf("Total proofs handled: %v, time elapsed %.1f seconds", total, elapsed)
	return fmt.Sprintf("\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}
