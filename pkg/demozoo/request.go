package demozoo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logs"
)

// Request Demozoo entries.
type Request struct {
	All       bool   // Parse all demozoo entries.
	Overwrite bool   // Overwrite any existing files.
	Refresh   bool   // Refresh all demozoo entries.
	ByID      string // Filter by ID.
}

// Query parses a single Demozoo entry.
func (r *Request) Query(id string) error {
	if err := database.CheckID(id); err != nil {
		return fmt.Errorf("query id %s: %w", id, err)
	}
	r.ByID = id
	if err := r.Queries(); err != nil {
		return fmt.Errorf("query queries: %w", err)
	}
	return nil
}

// Queries parses all new proofs.
// Overwrite will replace existing assets such as images.
// All parses every Demozoo entry, not just records waiting for approval.
func (r Request) Queries() error { //nolint:cyclop,funlen
	var st Stat
	stmt, start := selectByID(r.ByID), time.Now()
	db := database.Connect()
	defer db.Close()
	values, scanArgs, rows, err := values(db, stmt)
	if err != nil {
		return err
	}
	storage := directories.Init(false).UUID
	if err = st.sumTotal(Records{rows, scanArgs, values}, r); err != nil {
		return fmt.Errorf("queries sumtotal: %w", err)
	}
	queriesTotal(st.Total)
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
			logs.Danger(fmt.Errorf("queries nextrow: %w", err))
			continue
		} else if skip {
			continue
		}
		rec, err := NewRecord(st.Count, values)
		if err != nil {
			logs.Danger(fmt.Errorf("queries new: %w", err))
			continue
		}
		logs.Printcrf(rec.String(st.Total))
		if update := rec.check(); !update {
			continue
		}
		if skip, err := rec.parseAPI(st, r.Overwrite, storage); err != nil {
			logs.Danger(fmt.Errorf("queries parseapi: %w", err))
			continue
		} else if skip {
			continue
		}
		if st.Total == 0 {
			break
		}
		rec.save()
	}
	if r.ByID != "" {
		st.ByID = r.ByID
		st.printer()
		return nil
	}
	if st.Total > 0 {
		logs.Println()
	}
	st.summary(time.Since(start))
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

func queriesTotal(total int) {
	if total == 0 {
		logs.Println("nothing to do")
		return
	}
	logs.Println("Total records", total)
}

// Skip the Request?
func (r Request) skip() bool {
	if !r.All && !r.Refresh && !r.Overwrite {
		return true
	}
	return false
}
