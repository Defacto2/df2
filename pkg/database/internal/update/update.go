package update

import (
	"context"
	"fmt"
	"strings"

	"github.com/Defacto2/df2/pkg/database/internal/connect"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/gookit/color"
)

// Update row values based on conditions.
type Update struct {
	Query string
	Args  []any
}

// Execute Query and Args to update the database and returns the total number of changes.
func (u Update) Execute() (int64, error) {
	db := connect.Connect()
	defer db.Close()
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
	return count, db.Close()
}

type Column string

const (
	Filename Column = "files.filename"
	GroupBy  Column = "files.group_brand_by"
	GroupFor Column = "files.group_brand_for"
)

// NamedTitles remove record titles that match the filename.
func (col Column) NamedTitles() error {
	ctx := context.Background()
	db := connect.Connect()
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
		logs.Printcrf("no named title fixes needed")
		return nil
	}
	logs.Printcrf("%d named title fixes applied", rows)
	return nil
}

// Distinct returns a unique list of values from the table column.
func Distinct(column string) ([]string, error) {
	var result string
	db := connect.Connect()
	defer db.Close()
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
	return values, db.Close()
}

func Sections(sections *[]string) {
	var u Update
	u.Query = "UPDATE files SET section=? WHERE `section`=?"
	for _, s := range *sections {
		u.Args = []any{strings.ToLower(s), s}
		c, err := u.Execute()
		if err != nil {
			logs.Log(err)
		}
		if c == 0 {
			continue
		}
		str := fmt.Sprintf("%s %s \"%s\"",
			color.Question.Sprint(c), color.Info.Sprint("section ⟫"), color.Primary.Sprint(s))
		printcr(c, &str)
	}
	// set all audio platform files to use intro section
	// releaseadvert
	u.Query = "UPDATE files SET section=? WHERE `platform`=?"
	u.Args = []any{"releaseadvert", "audio"}
	c, err := u.Execute()
	if err != nil {
		logs.Log(err)
	}
	if c == 0 {
		return
	}
	str := fmt.Sprintf("%s %s \"%s\"",
		color.Question.Sprint(c), color.Info.Sprint("platform ⟫ audio ⟫"), color.Primary.Sprint("releaseadvert"))
	printcr(c, &str)
}

func Platforms(platforms *[]string) {
	var u Update
	u.Query = "UPDATE files SET platform=? WHERE `platform`=?"
	for _, p := range *platforms {
		u.Args = []any{strings.ToLower(p), p}
		c, err := u.Execute()
		if err != nil {
			logs.Log(err)
		}
		if c == 0 {
			continue
		}
		s := fmt.Sprintf("%s %s \"%s\"",
			color.Question.Sprint(c), color.Info.Sprint("platform ⟫"), color.Primary.Sprint(p))
		printcr(c, &s)
	}
}

func printcr(i int64, s *string) {
	if i == 0 {
		logs.Printcr(*s)
		return
	}
	logs.Println("\n" + *s)
}
