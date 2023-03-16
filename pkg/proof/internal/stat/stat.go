package stat

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logger"
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

func Init(cfg configger.Config) (Proof, error) {
	dir, err := directories.Init(cfg, false)
	if err != nil {
		return Proof{}, err
	}
	return Proof{
		Base:     logger.SprintPath(dir.UUID),
		BasePath: dir.UUID,
		start:    time.Now(),
	}, nil
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
