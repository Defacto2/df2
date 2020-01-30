package demozoo

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

// RefreshQueries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (req Request) RefreshQueries() error {
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(sqlSelect())
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	//dir := directories.Init(false)
	// fetch the rows
	rw := row{count: 0, missing: 0}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		rw.count++
		r := newRecord(rw.count, values)
		logs.Print(r.String()) // counter and record intro
		// confirm API request
		code, status, api := Fetch(r.WebIDDemozoo)
		if code == 404 {
			r.WebIDDemozoo = 0
			if err = r.Save(); err == nil {
				logs.Printf("(%s)\n", download.StatusColor(code, status))
			}
			continue
		}
		if code < 200 || code > 299 {
			logs.Printf("(%s)\n", download.StatusColor(code, status))
			continue
		}
		// pouet id
		pid, _ := api.PouetID(false)
		if r.WebIDPouet != uint(pid) {
			r.WebIDPouet = uint(pid)
			logs.Printf("PN:%s ", color.Note.Sprint(pid))
		}
		// groups
		// this is disabled due to different methodologies between Defacto2 and DZ
		// DZ puts groups into the Title, DF2 needs the groups categories
		// example: https://demozoo.org/productions/158376/
		if r.Section != "magazine" && r.Section == "magazine" {
			grps := api.Groups()
			if !strings.EqualFold(r.GroupFor, grps[0]) {
				logs.Printf("for:%s ", color.Info.Sprint(grps[0]))
				r.GroupFor = grps[0]
			}
			if !strings.EqualFold(r.GroupBy, grps[1]) {
				if grps[1] == "" {
					logs.Printf("by:%s ", color.Danger.Sprint(r.GroupBy))
				} else {
					logs.Printf("by:%s ", color.Info.Sprint(grps[1]))
				}
				r.GroupBy = grps[1]
			}
		}
		// title
		if r.Section != "magazine" && !strings.EqualFold(r.Title, api.Title) {
			logs.Printf("i:%s ", color.Secondary.Sprint(api.Title))
			r.Title = api.Title
		}
		// credits
		credits := api.Authors()
		r.authors(credits)
		// save results
		if reflect.DeepEqual(newRecord(rw.count, values), r) {
			logs.Printf("• skipped %v\n", logs.Y())
			continue
		}
		if err = r.Save(); err != nil {
			logs.Printf("• saved %v ", logs.X())
			logs.Log(err)
		} else {
			logs.Printf("• saved %v\n", logs.Y())
		}
	}
	logs.Check(rows.Err())
	rw.summary()
	return nil
}

func (r *Record) authors(a Authors) {
	compare := func(n, o []string, i string) bool {
		if !reflect.DeepEqual(n, o) {
			logs.Printf("c%s:%s ", i, color.Secondary.Sprint(n))
			if len(o) > 1 {
				logs.Printf("%s ", color.Danger.Sprint(o))
			}
			return false
		}
		return true
	}
	if len(a.art) > 1 {
		new := a.art
		old := r.CreditArt
		if !compare(new, old, "a") {
			r.CreditArt = new
		}
	}
	if len(a.audio) > 1 {
		new := a.audio
		old := r.CreditAudio
		if !compare(new, old, "m") {
			r.CreditAudio = new
		}
	}
	if len(a.code) > 1 {
		new := a.code
		old := r.CreditCode
		if !compare(new, old, "c") {
			r.CreditCode = new
		}
	}
	if len(a.text) > 1 {
		new := a.text
		old := r.CreditText
		if !compare(new, old, "t") {
			r.CreditText = new
		}
	}
}
