// Package update handles edits and updates to the database records.
package update

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Defacto2/df2/pkg/logger"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
)

var (
	ErrColumn  = errors.New("column value must be either platform or section")
	ErrDB      = errors.New("database handle pointer cannot be nil")
	ErrPointer = errors.New("pointer value cannot be nil")
)

// Update row values based on conditions.
type Update struct {
	Query string // Query is an SQL statement.
	Args  []any  // Args are SQL statement values.
}

// Execute Query and Args to update the database and returns the total number of changes.
func (u Update) Execute(db *sql.DB) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	query, err := db.Prepare(u.Query)
	if err != nil {
		return 0, fmt.Errorf("update execute prepare: %w", err)
	}
	defer query.Close()
	res, err := query.Exec(u.Args...)
	if err != nil {
		return 0, fmt.Errorf("update execute exec: %w", err)
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
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, "UPDATE files SET record_title = \"\" WHERE files.record_title = ?", string(col))
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	str.Total(w, int(rows), fmt.Sprintf("named title fixes applied, `%s`", string(col)))
	return nil
}

// Distinct returns a unique list of values from the table column.
// Column must be either platform or section.
func Distinct(db *sql.DB, column string) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	if column != "platform" && column != "section" {
		return nil, ErrColumn
	}
	stmt := "SELECT DISTINCT platform FROM `files` WHERE platform != \"\""
	if column == "section" {
		stmt = "SELECT DISTINCT section FROM `files` WHERE section != \"\""
	}
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("distinct query %q: %w", column, err)
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	result := ""
	values := []string{}
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("distinct rows scan: %w", err)
		}
		values = append(values, result)
	}
	return values, nil
}

func Sections(db *sql.DB, w io.Writer, sections *[]string) error {
	if db == nil {
		return ErrDB
	}
	if sections == nil {
		return fmt.Errorf("sections %w", ErrPointer)
	}
	u := Update{
		Query: "UPDATE files SET section=? WHERE `section`=?",
	}
	for _, s := range *sections {
		u.Args = []any{strings.ToLower(s), s}
		c, err := u.Execute(db)
		if err != nil {
			fmt.Fprintln(w, err)
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
	u = Update{
		Query: "UPDATE files SET section=? WHERE `platform`=?",
		Args:  []any{"releaseadvert", "audio"},
	}
	c, err := u.Execute(db)
	if err != nil {
		return fmt.Errorf("execute %w", err)
	}
	if c == 0 {
		return nil
	}
	str := fmt.Sprintf("%s %s \"%s\"",
		color.Question.Sprint(c), color.Info.Sprint("platform ⟫ audio ⟫"), color.Primary.Sprint("releaseadvert"))
	printcr(w, c, &str)
	return nil
}

func Platforms(db *sql.DB, w io.Writer, platforms *[]string) error {
	if db == nil {
		return ErrDB
	}
	if platforms == nil {
		return fmt.Errorf("platforms %w", ErrPointer)
	}
	u := Update{
		Query: "UPDATE files SET platform=? WHERE `platform`=?",
	}
	for _, p := range *platforms {
		u.Args = []any{strings.ToLower(p), p}
		c, err := u.Execute(db)
		if err != nil {
			fmt.Fprintln(w, err)
		}
		if c == 0 {
			continue
		}
		s := fmt.Sprintf("%s %s \"%s\"",
			color.Question.Sprint(c), color.Info.Sprint("platform ⟫"), color.Primary.Sprint(p))
		printcr(w, c, &s)
	}
	return nil
}

func printcr(w io.Writer, i int64, s *string) {
	if i == 0 {
		logger.PrintCR(w, *s)
		return
	}
	fmt.Fprintln(w, "\n"+*s)
}
