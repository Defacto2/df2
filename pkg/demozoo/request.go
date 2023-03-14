package demozoo

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logs"
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
func (r *Request) Query(w io.Writer, log *zap.SugaredLogger, id string) error {
	if err := database.CheckID(id); err != nil {
		return fmt.Errorf("query id %s: %w", id, err)
	}
	r.ByID = id
	if err := r.Queries(w, log); err != nil {
		return fmt.Errorf("query queries: %w", err)
	}
	return nil
}

// Queries parses all new proofs.
// Overwrite will replace existing assets such as images.
// All parses every Demozoo entry, not just records waiting for approval.
func (r Request) Queries(w io.Writer, log *zap.SugaredLogger) error { //nolint:cyclop,funlen
	var st Stat
	stmt, start := selectByID(r.ByID), time.Now()
	db := database.Connect(w)
	defer db.Close()
	values, scanArgs, rows, err := values(db, stmt)
	if err != nil {
		return err
	}
	storage := directories.Init(false).UUID
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
		logs.Printcrf(w, rec.String(st.Total))
		if update := rec.check(w); !update {
			continue
		}
		if skip, err := rec.parseAPI(w, log, st, r.Overwrite, storage); err != nil {
			log.Errorf("queries parseapi: %w", err)
			continue
		} else if skip {
			continue
		}
		if st.Total == 0 {
			break
		}
		rec.save(w, log)
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
