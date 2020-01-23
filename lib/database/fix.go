package database

import (
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

// Fix any malformed section and platforms found in the database.
func Fix() {
	dist, err := distinct("section")
	logs.Check(err)
	updateSections(&dist)
	dist, err = distinct("platform")
	logs.Check(err)
	updatePlatforms(&dist)
}

// Update row values based on conditions.
type Update struct {
	Query string
	Args  []interface{}
}

// Execute Query and Args to update the database and returns the total number of changes.
func (u Update) Execute() (int64, error) {
	db := Connect()
	defer db.Close()
	update, err := db.Prepare(u.Query)
	if err != nil {
		return 0, err
	}
	res, err := update.Exec(u.Args...)
	if err != nil {
		return 0, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return ra, nil
}

func distinct(column string) ([]string, error) {
	var r []string
	var result string
	db := Connect()
	defer db.Close()
	rows, err := db.Query("SELECT DISTINCT `" + column + "` AS `result` FROM `files` WHERE `" + column + "` != \"\"")
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&result)
		logs.Log(err)
		r = append(r, result)
	}
	return r, nil
}

func updatePlatforms(platforms *[]string) {
	var plat Update
	plat.Query = "UPDATE files SET platform=? WHERE `platform`=?"
	for _, s := range *platforms {
		plat.Args = []interface{}{strings.ToLower(s), s}
		res, err := plat.Execute()
		if err != nil {
			logs.Log(err)
		}
		logs.Printf("%s %s \"%s\"\n", color.Question.Sprint(res), color.Info.Sprint("⟫ platform"), color.Primary.Sprint(s))
	}
}

func updateSections(sections *[]string) {
	var sect Update
	sect.Query = "UPDATE files SET section=? WHERE `section`=?"
	for _, s := range *sections {
		sect.Args = []interface{}{strings.ToLower(s), s}
		res, err := sect.Execute()
		if err != nil {
			logs.Log(err)
		}
		logs.Printf("%s %s \"%s\"\n", color.Question.Sprint(res), color.Info.Sprint("⟫ section"), color.Primary.Sprint(s))
	}
}
