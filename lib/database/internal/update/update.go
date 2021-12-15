package update

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Defacto2/df2/lib/database/internal/my57"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

// Update row values based on conditions.
type Update struct {
	Query string
	Args  []interface{}
}

// Execute Query and Args to update the database and returns the total number of changes.
func (u Update) Execute() (int64, error) {
	db := my57.Connect()
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

func Distinct(column string) ([]string, error) {
	var result string
	db := my57.Connect()
	defer db.Close()
	rows, err := db.Query("SELECT DISTINCT ? AS `result` FROM `files` WHERE ? != \"\"", column, column)
	if err != nil {
		return nil, fmt.Errorf("distinct query %q: %w", column, err)
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("distinct rows scan: %w", err)
		}
		values = append(values, result)
	}
	return values, db.Close()
}

func NamedTitles() {
	ctx := context.Background()
	db := my57.Connect()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	result, err := tx.ExecContext(ctx, "UPDATE files SET record_title=\"\" WHERE files.record_title=files.filename")
	if err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			log.Fatal(err1)
		}
		log.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rows == 0 {
		logs.Print("no named title fixes needed")
		return
	}
	logs.Printcrf("%d named title fixes applied", rows)
}

func Sections(sections *[]string) {
	var u Update
	u.Query = "UPDATE files SET section=? WHERE `section`=?"
	for _, s := range *sections {
		u.Args = []interface{}{strings.ToLower(s), s}
		c, err := u.Execute()
		if err != nil {
			logs.Log(err)
		}
		str := fmt.Sprintf("%s %s \"%s\"",
			color.Question.Sprint(c), color.Info.Sprint("section ⟫"), color.Primary.Sprint(s))
		printcr(c, &str)
	}
	// set all audio platform files to use intro section
	// releaseadvert
	u.Query = "UPDATE files SET section=? WHERE `platform`=?"
	u.Args = []interface{}{"releaseadvert", "audio"}
	c, err := u.Execute()
	if err != nil {
		logs.Log(err)
	}
	str := fmt.Sprintf("%s %s \"%s\"",
		color.Question.Sprint(c), color.Info.Sprint("platform ⟫ audio ⟫"), color.Primary.Sprint("releaseadvert"))
	printcr(c, &str)
}

func Platforms(platforms *[]string) {
	var u Update
	u.Query = "UPDATE files SET platform=? WHERE `platform`=?"
	for _, p := range *platforms {
		u.Args = []interface{}{strings.ToLower(p), p}
		c, err := u.Execute()
		if err != nil {
			logs.Log(err)
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
