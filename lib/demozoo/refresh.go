package demozoo

import (
	"database/sql"
	"fmt"
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

// RefreshMeta synchronises missing file entries with Demozoo sourced metadata.
func RefreshMeta() error {
	start := time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(selectByID())
	if err != nil {
		return fmt.Errorf("refresh meta query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("refresh meta rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("refresh meta columns: %w", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// fetch the rows
	st := stat{}
	for rows.Next() {
		if _, err := st.nextRefresh(records{rows, scanArgs, values}); err != nil {
			return fmt.Errorf("refresh meta next row: %w", err)
		}
	}
	st.summary(time.Since(start))
	return nil
}

func (st *stat) nextRefresh(rec records) (skip bool, err error) {
	if err = rec.rows.Scan(rec.scanArgs...); err != nil {
		return true, fmt.Errorf("next refresh rows scan: %w", err)
	}
	st.count++
	r, err := newRecord(st.count, rec.values)
	if err != nil {
		return true, fmt.Errorf("next refresh new record 1: %w", err)
	}
	logs.Printcrf(r.String(0))
	code, status, api := Fetch(r.WebIDDemozoo)
	if ok, err := r.confirm(code, status); err != nil {
		return true, fmt.Errorf("next refresh confirm: %w", err)
	} else if !ok {
		return true, nil
	}
	r.pouet(api)
	r.title(api)
	r.authors(api.Authors())
	new, err := newRecord(st.count, rec.values)
	if err != nil {
		return true, fmt.Errorf("next refresh new record 2: %w", err)
	}
	if reflect.DeepEqual(new, r) {
		logs.Printf("• skipped %v", logs.Y())
		return true, nil
	}
	if err = r.Save(); err != nil {
		logs.Printf("• saved %v ", logs.X())
		return true, fmt.Errorf("next refresh save: %w", err)
	}
	logs.Printf("• saved %v", logs.Y())
	return false, nil
}

func (r *Record) confirm(code int, status string) (ok bool, err error) {
	const nofound, found, problems = 404, 200, 300
	if code == nofound {
		r.WebIDDemozoo = 0
		if err := r.Save(); err != nil {
			return true, fmt.Errorf("record confirm save: %w", err)
		}
		logs.Printf("(%s)\n", download.StatusColor(code, status))
		return false, nil
	}
	if code < found || code >= problems {
		logs.Printf("(%s)\n", download.StatusColor(code, status))
		return false, nil
	}
	return true, nil
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
		new, old := a.art, r.CreditArt
		if !compare(new, old, "a") {
			r.CreditArt = new
		}
	}
	if len(a.audio) > 1 {
		new, old := a.audio, r.CreditAudio
		if !compare(new, old, "m") {
			r.CreditAudio = new
		}
	}
	if len(a.code) > 1 {
		new, old := a.code, r.CreditCode
		if !compare(new, old, "c") {
			r.CreditCode = new
		}
	}
	if len(a.text) > 1 {
		new, old := a.text, r.CreditText
		if !compare(new, old, "t") {
			r.CreditText = new
		}
	}
}
