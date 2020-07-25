package database

import (
	"fmt"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

// Fix any malformed section and platforms found in the database.
func Fix() error {
	dist, err := distinct("section")
	if err != nil {
		return fmt.Errorf("fix distinct section: %w", err)
	}
	updateSections(&dist)
	dist, err = distinct("platform")
	if err != nil {
		return fmt.Errorf("fix distinct platform: %w", err)
	}
	updatePlatforms(&dist)
	return nil
}

// Update row values based on conditions.
type Update struct {
	Query string
	Args  []interface{}
}

// Execute Query and Args to update the database and returns the total number of changes.
func (u Update) Execute() (count int64, err error) {
	db := Connect()
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
	count, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("update execute rows affected: %w", err)
	}
	return count, db.Close()
}

func distinct(column string) (values []string, err error) {
	var result string
	db := Connect()
	defer db.Close()
	rows, err := db.Query("SELECT DISTINCT ? AS `result` FROM `files` WHERE ? != \"\"", column, column)
	if err != nil {
		return nil, fmt.Errorf("distinct query %q: %w", column, err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("distinct rows scan: %w", err)
		}
		values = append(values, result)
	}
	return values, db.Close()
}

func print(i int64, s *string) {
	if i == 0 {
		logs.Printcr(*s)
	} else {
		logs.Println("\n" + *s)
	}
}

func updatePlatforms(platforms *[]string) {
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
		print(c, &s)
	}
}

func updateSections(sections *[]string) {
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
		print(c, &str)
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
	print(c, &str)
}
