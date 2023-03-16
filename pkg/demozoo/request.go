package demozoo

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logger"
	"go.uber.org/zap"
)

// Request Demozoo entries.
type Request struct {
	All       bool   // Parse all demozoo entries.
	Overwrite bool   // Overwrite any existing files.
	Refresh   bool   // Refresh all demozoo entries.
	ByID      string // Filter by ID.
}

// Query parses a single Demozoo entry.
func (r *Request) Query(db *sql.DB, w io.Writer, log *zap.SugaredLogger, cfg configger.Config, id string) error {
	if err := database.CheckID(id); err != nil {
		return fmt.Errorf("query id %s: %w", id, err)
	}
	r.ByID = id
	if err := r.Queries(db, w, log, cfg); err != nil {
		return fmt.Errorf("query queries: %w", err)
	}
	return nil
}

// Queries parses all new proofs.
// Overwrite will replace existing assets such as images.
// All parses every Demozoo entry, not just records waiting for approval.
func (r Request) Queries(db *sql.DB, w io.Writer, log *zap.SugaredLogger, cfg configger.Config) error { //nolint:cyclop,funlen
	var st Stat
	stmt, start := selectByID(r.ByID), time.Now()
	values, scanArgs, rows, err := values(db, stmt)
	if err != nil {
		return err
	}
	dir, err := directories.Init(cfg, false)
	if err != nil {
		return err
	}
	storage := dir.UUID
	if err = st.sumTotal(Records{rows, scanArgs, values}, r); err != nil {
		return fmt.Errorf("queries sumtotal: %w", err)
	}
	queriesTotal(w, st.Total)
	rows, err = db.Query(stmt)
	if err != nil {
		return fmt.Errorf("queries query 2: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("queries rows 2: %w", rows.Err())
	}
	defer rows.Close()
	for rows.Next() {
		st.Fetched++
		if skip, err := st.nextResult(Records{rows, scanArgs, values}, r); err != nil {
			log.Errorf("queries nextrow: %w", err)
			continue
		} else if skip {
			continue
		}
		rec, err := NewRecord(log, st.Count, values)
		if err != nil {
			log.Errorf("queries new: %w", err)
			continue
		}
		logger.Printcrf(w, rec.String(st.Total))
		if update := rec.check(w); !update {
			continue
		}
		if skip, err := rec.parseAPI(db, w, log, cfg, st, r.Overwrite, storage); err != nil {
			log.Errorf("queries parseapi: %w", err)
			continue
		} else if skip {
			continue
		}
		if st.Total == 0 {
			break
		}
		rec.save(db, w, log)
	}
	if r.ByID != "" {
		st.ByID = r.ByID
		st.printer(w)
		return nil
	}
	if st.Total > 0 {
		fmt.Fprintln(w)
	}
	st.summary(w, time.Since(start))
	return nil
}

func values(db *sql.DB, stmt string) ([]sql.RawBytes, []any, *sql.Rows, error) {
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("queries query 1: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("queries rows 1: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("queries columns: %w", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	return values, scanArgs, rows, nil
}

func queriesTotal(w io.Writer, total int) {
	if total == 0 {
		fmt.Fprintln(w, "nothing to do")
		return
	}
	fmt.Fprintln(w, "Total records", total)
}

// Skip the Request?
func (r Request) skip() bool {
	if !r.All && !r.Refresh && !r.Overwrite {
		return true
	}
	return false
}
