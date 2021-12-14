package demozoo

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/demozoo/internal/prods"
	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
)

type Records struct {
	Rows     *sql.Rows
	ScanArgs []interface{}
	Values   []sql.RawBytes
}

// RefreshMeta synchronises missing file entries with Demozoo sourced metadata.
func RefreshMeta() error {
	start := time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(selectByID(""))
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
	var st Stat
	for rows.Next() {
		if _, err := st.NextRefresh(Records{rows, scanArgs, values}); err != nil {
			return fmt.Errorf("refresh meta next row: %w", err)
		}
	}
	st.summary(time.Since(start))
	return nil
}

func (st *Stat) NextRefresh(rec Records) (skip bool, err error) {
	if err = rec.Rows.Scan(rec.ScanArgs...); err != nil {
		return true, fmt.Errorf("next refresh rows scan: %w", err)
	}
	st.Count++
	r, err := NewRecord(st.Count, rec.Values)
	if err != nil {
		return true, fmt.Errorf("next refresh new record 1: %w", err)
	}
	logs.Printcrf(r.String(0))
	f, err := Fetch(r.WebIDDemozoo)
	if err != nil {
		return true, fmt.Errorf("next refresh fetch: %w", err)
	}
	var ok bool
	code, status, api := f.Code, f.Status, f.API
	if ok, err = r.confirm(code, status); err != nil {
		return true, fmt.Errorf("next refresh confirm: %w", err)
	} else if !ok {
		return true, nil
	}
	if err = r.pouet(&api); err != nil {
		return true, fmt.Errorf("next refresh: %w", err)
	}
	r.title(&api)
	a := api.Authors()
	r.authors(&a)
	var nr Record
	nr, err = NewRecord(st.Count, rec.Values)
	if err != nil {
		return true, fmt.Errorf("next refresh new record 2: %w", err)
	}
	if reflect.DeepEqual(nr, r) {
		logs.Printf("• skipped %v", str.Y())
		return true, nil
	}
	if err = r.Save(); err != nil {
		logs.Printf("• saved %v ", str.X())
		return true, fmt.Errorf("next refresh save: %w", err)
	}
	logs.Printf("• saved %v", str.Y())
	return false, nil
}

func (r *Record) authors(a *prods.Authors) {
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
	if len(a.Art) > 1 {
		n, old := a.Art, r.CreditArt
		if !compare(n, old, "a") {
			r.CreditArt = n
		}
	}
	if len(a.Audio) > 1 {
		n, old := a.Audio, r.CreditAudio
		if !compare(n, old, "m") {
			r.CreditAudio = n
		}
	}
	if len(a.Code) > 1 {
		n, old := a.Code, r.CreditCode
		if !compare(n, old, "c") {
			r.CreditCode = n
		}
	}
	if len(a.Text) > 1 {
		n, old := a.Text, r.CreditText
		if !compare(n, old, "t") {
			r.CreditText = n
		}
	}
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

func (r *Record) pouet(api *prods.ProductionsAPIv1) error {
	pid, _, err := api.PouetID(false)
	if err != nil {
		return fmt.Errorf("pouet: %w", err)
	}
	if r.WebIDPouet != uint(pid) {
		r.WebIDPouet = uint(pid)
		logs.Printf("PN:%s ", color.Note.Sprint(pid))
	}
	return nil
}

func (r *Record) title(api *prods.ProductionsAPIv1) {
	if r.Section != Magazine.String() && !strings.EqualFold(r.Title, api.Title) {
		logs.Printf("i:%s ", color.Secondary.Sprint(api.Title))
		r.Title = api.Title
	}
}
