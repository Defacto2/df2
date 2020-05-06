package database

import (
	"fmt"
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
func (u Update) Execute() (count int64, err error) {
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
	count, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func distinct(column string) (values []string, err error) {
	var result string
	db := Connect()
	defer db.Close()
	rows, err := db.Query("SELECT DISTINCT `" + column + "` AS `result` FROM `files` WHERE `" + column + "` != \"\"")
	if err != nil {
		return values, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&result)
		logs.Log(err)
		values = append(values, result)
	}
	return values, nil
}

func print(res int64, str *string) {
	if res == 0 {
		logs.Printcr(*str)
	} else {
		logs.Println("\n" + *str)
	}
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
		str := fmt.Sprintf("%s %s \"%s\"", color.Question.Sprint(res), color.Info.Sprint("platform ⟫"), color.Primary.Sprint(s))
		print(res, &str)
	}
}

func updateSections(sections *[]string) {
	var sect Update
	// apply lowercase to all section values
	sect.Query = "UPDATE files SET section=? WHERE `section`=?"
	for _, s := range *sections {
		sect.Args = []interface{}{strings.ToLower(s), s}
		res, err := sect.Execute()
		if err != nil {
			logs.Log(err)
		}
		str := fmt.Sprintf("%s %s \"%s\"", color.Question.Sprint(res), color.Info.Sprint("section ⟫"), color.Primary.Sprint(s))
		print(res, &str)
	}
	// set all audio platform files to use intro section
	// releaseadvert
	sect.Query = "UPDATE files SET section=? WHERE `platform`=?"
	sect.Args = []interface{}{"releaseadvert", "audio"}
	res, err := sect.Execute()
	if err != nil {
		logs.Log(err)
	}
	str := fmt.Sprintf("%s %s \"%s\"", color.Question.Sprint(res), color.Info.Sprint("platform ⟫ audio ⟫"), color.Primary.Sprint("releaseadvert"))
	print(res, &str)
}
