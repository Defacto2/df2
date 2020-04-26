package demozoo

import (
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

type records struct {
	rows     *sql.Rows
	scanArgs []interface{}
	values   []sql.RawBytes
}

// RefreshQueries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (req Request) RefreshQueries() error {
	start := time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(selectByID())
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
	// fetch the rows
	st := stat{count: 0, missing: 0}
	for rows.Next() {
		if skip := st.nextRefresh(records{rows, scanArgs, values}); skip {
			continue
		}
	}
	logs.Check(rows.Err())
	st.summary(time.Since(start))
	return nil
}

func (st *stat) nextRefresh(rec records) (skip bool) {
	err := rec.rows.Scan(rec.scanArgs...)
	logs.Check(err)
	st.count++
	r := newRecord(st.count, rec.values)
	logs.Printfcr(r.String(0))
	code, status, api := Fetch(r.WebIDDemozoo)
	if ok := r.confirm(code, status); !ok {
		return true
	}
	r.pouet(api)
	r.title(api)
	r.authors(api.Authors())
	if reflect.DeepEqual(newRecord(st.count, rec.values), r) {
		logs.Printf("• skipped %v", logs.Y())
		return true
	}
	if err = r.Save(); err != nil {
		logs.Printf("• saved %v ", logs.X())
		logs.Log(err)
	} else {
		logs.Printf("• saved %v", logs.Y())
	}
	return false
}

func (r *Record) confirm(code int, status string) (ok bool) {
	if code == 404 {
		r.WebIDDemozoo = 0
		if err := r.Save(); err == nil {
			logs.Printf("(%s)\n", download.StatusColor(code, status))
		}
		return false
	}
	if code < 200 || code > 299 {
		logs.Printf("(%s)\n", download.StatusColor(code, status))
		return false
	}
	return true
}

func (r *Record) groups(api ProductionsAPIv1) {
	// this is removed due to different methodologies between Defacto2 and DZ
	// DZ puts groups into the Title, DF2 needs the groups categories
	// example: https://demozoo.org/productions/158376/
}

func (r *Record) pouet(api ProductionsAPIv1) {
	pid, _ := api.PouetID(false)
	if r.WebIDPouet != uint(pid) {
		r.WebIDPouet = uint(pid)
		logs.Printf("PN:%s ", color.Note.Sprint(pid))
	}
}

func (r *Record) title(api ProductionsAPIv1) {
	if r.Section != "magazine" && !strings.EqualFold(r.Title, api.Title) {
		logs.Printf("i:%s ", color.Secondary.Sprint(api.Title))
		r.Title = api.Title
	}
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
