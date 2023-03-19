package update

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/Defacto2/df2/pkg/logger"
	"github.com/gookit/color"
	"go.uber.org/zap"
)

// Update row values based on conditions.
type Update struct {
	Query string
	Args  []any
}

// Execute Query and Args to update the database and returns the total number of changes.
func (u Update) Execute(db *sql.DB) (int64, error) {
	update, err := db.Prepare(u.Query)
	if err != nil {
		return 0, fmt.Errorf("update execute db prepare: %w", err)
	}
	defer update.Close()
	res, err := update.Exec(u.Args...)
	if err != nil {
		return 0, fmt.Errorf("update execute db exec: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("update execute rows affected: %w", err)
	}
	return count, nil
}

type Column string

const (
	Filename Column = "files.filename"
	GroupBy  Column = "files.group_brand_by"
	GroupFor Column = "files.group_brand_for"
)

// NamedTitles remove record titles that match the filename.
func (col Column) NamedTitles(db *sql.DB, w io.Writer) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, "UPDATE files SET record_title = \"\" WHERE files.record_title = ?", string(col))
	if err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			return err1
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		logger.Printcrf(w, "no named title fixes needed")
		return nil
	}
	logger.Printcrf(w, "%d named title fixes applied", rows)
	return nil
}

// Distinct returns a unique list of values from the table column.
func Distinct(db *sql.DB, column string) ([]string, error) {
	var result string
	rows, err := db.Query("SELECT DISTINCT ? AS `result` FROM `files` WHERE ? != \"\"", column, column)
	if err != nil {
		return nil, fmt.Errorf("distinct query %q: %w", column, err)
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	values := []string{}
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("distinct rows scan: %w", err)
		}
		values = append(values, result)
	}
	return values, nil
}

func Sections(db *sql.DB, w io.Writer, l *zap.SugaredLogger, sections *[]string) {
	var u Update
	u.Query = "UPDATE files SET section=? WHERE `section`=?"
	for _, s := range *sections {
		u.Args = []any{strings.ToLower(s), s}
		c, err := u.Execute(db)
		if err != nil {
			l.Errorln(err)
		}
		if c == 0 {
			continue
		}
		str := fmt.Sprintf("%s %s \"%s\"",
			color.Question.Sprint(c), color.Info.Sprint("section ⟫"), color.Primary.Sprint(s))
		printcr(w, c, &str)
	}
	// set all audio platform files to use intro section
	// releaseadvert
	u.Query = "UPDATE files SET section=? WHERE `platform`=?"
	u.Args = []any{"releaseadvert", "audio"}
	c, err := u.Execute(db)
	if err != nil {
		l.Errorln(err)
	}
	if c == 0 {
		return
	}
	str := fmt.Sprintf("%s %s \"%s\"",
		color.Question.Sprint(c), color.Info.Sprint("platform ⟫ audio ⟫"), color.Primary.Sprint("releaseadvert"))
	printcr(w, c, &str)
}

func Platforms(db *sql.DB, w io.Writer, l *zap.SugaredLogger, platforms *[]string) {
	var u Update
	u.Query = "UPDATE files SET platform=? WHERE `platform`=?"
	for _, p := range *platforms {
		u.Args = []any{strings.ToLower(p), p}
		c, err := u.Execute(db)
		if err != nil {
			l.Errorln(err)
		}
		if c == 0 {
			continue
		}
		s := fmt.Sprintf("%s %s \"%s\"",
			color.Question.Sprint(c), color.Info.Sprint("platform ⟫"), color.Primary.Sprint(p))
		printcr(w, c, &s)
	}
}

func printcr(w io.Writer, i int64, s *string) {
	if i == 0 {
		logger.Printcr(w, *s)
		return
	}
	fmt.Fprintln(w, "\n"+*s)
}
