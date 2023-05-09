// Package zipcontent processes the directory and file content of a file archive.
package zipcontent

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/record"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
	"go.uber.org/zap"
)

func stmt() string {
	const s = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`updatedat`,`retrotxt_readme`"
	const w = " WHERE file_zip_content IS NULL AND (`filename` LIKE '%.zip' OR `filename`" +
		" LIKE '%.rar' OR `filename` LIKE '%.7z')"
	return fmt.Sprintf("%s FROM `files` %s", s, w)
}

// Fix the content of zip archives within in the database.
func Fix( //nolint:cyclop,funlen
	db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg conf.Config, summary bool,
) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	s, err := scan.Init(cfg)
	if err != nil {
		return err
	}
	rows, err := db.Query(stmt())
	if err != nil {
		return err
	} else if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	args := make([]any, len(values))
	for i := range values {
		args[i] = &values[i]
	}
	for rows.Next() {
		s.Total++
	}
	rows, err = db.Query(stmt())
	if err != nil {
		return err
	} else if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return err
		}
		s.Count++
		r, err := record.New(values, s.BasePath)
		if err != nil {
			return err
		}
		s.Columns = columns
		s.Values = &values
		if err := r.Iterate(db, w, &s); err != nil {
			fmt.Fprintln(w)
			l.Errorln(err)
			continue
		}
		fmt.Fprintln(w)
	}
	if summary {
		s.Summary(w)
	}
	return nil
}
