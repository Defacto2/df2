package demozoo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
)

// Stat are the remote query statistics.
type Stat struct {
	Count   int
	Fetched int
	Missing int
	Total   int
	ByID    string
}

// FileExist returns false if the FilePath of the record points to a missing file.
func (st *Stat) FileExist(r *Record) bool {
	if s, err := os.Stat(r.FilePath); os.IsNotExist(err) || s.IsDir() {
		st.Missing++
		return true
	}
	return false
}

// Records contain more than one Record.
type Records struct {
	Rows     *sql.Rows
	ScanArgs []any
	Values   []sql.RawBytes
}

// NextRefresh iterates over the Records to update sync their Demozoo data to the database.
func (st *Stat) NextRefresh(rec Records) error {
	if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
		return fmt.Errorf("next scan: %w", err)
	}
	st.Count++
	r, err := NewRecord(st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 1: %w", err)
	}
	logs.Printcrf(r.String(0))
	var f Product
	err = f.Get(r.WebIDDemozoo)
	if err != nil {
		return fmt.Errorf("next fetch: %w", err)
	}
	var ok bool
	code, status, api := f.Code, f.Status, f.API
	if ok, err = r.confirm(code, status); err != nil {
		return fmt.Errorf("next confirm: %w", err)
	} else if !ok {
		return nil
	}
	if err = r.pouet(&api); err != nil {
		return fmt.Errorf("next refresh: %w", err)
	}
	r.title(&api)
	a := api.Authors()
	r.authors(&a)
	var nr Record
	nr, err = NewRecord(st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 2: %w", err)
	}
	if reflect.DeepEqual(nr, r) {
		logs.Printf("• skipped %v", str.Y())
		return nil
	}
	if err = r.Save(); err != nil {
		logs.Printf("• saved %v ", str.X())
		return fmt.Errorf("next save: %w", err)
	}
	logs.Printf("• saved %v", str.Y())
	return nil
}

// NextPouet iterates over the linked Demozoo records and sync any linked Pouet data to the local files table.
func (st *Stat) NextPouet(rec Records) error {
	if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
		return fmt.Errorf("next scan: %w", err)
	}
	st.Count++
	r, err := NewRecord(st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 1: %w", err)
	}
	if r.WebIDPouet > 0 {
		return nil
	}
	logs.Printcrf(r.String(0))
	var f Product
	err = f.Get(r.WebIDDemozoo)
	if err != nil {
		return fmt.Errorf("next fetch: %w", err)
	}
	var ok bool
	code, status, api := f.Code, f.Status, f.API
	if ok, err = r.confirm(code, status); err != nil {
		return fmt.Errorf("next confirm: %w", err)
	} else if !ok {
		return nil
	}
	if err = r.pouet(&api); err != nil {
		return fmt.Errorf("next refresh: %w", err)
	}
	var nr Record
	nr, err = NewRecord(st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 2: %w", err)
	}
	if reflect.DeepEqual(nr, r) {
		logs.Printf("• skipped %v", str.Y())
		return nil
	}
	if err = r.Save(); err != nil {
		logs.Printf("• saved %v ", str.X())
		return fmt.Errorf("next save: %w", err)
	}
	logs.Printf("• saved %v", str.Y())
	return nil
}

// nextResult checks for the next new record.
func (st *Stat) nextResult(rec Records, req Request) (skip bool, err error) {
	if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
		return false, fmt.Errorf("next result rows scan: %w", err)
	}
	if n := database.IsDemozoo(rec.Values); !n && req.skip() {
		return true, nil
	}
	st.Count++
	return false, nil
}

func (st Stat) print() {
	if st.Count == 0 {
		if st.Fetched == 0 {
			log.Printf("id %q does not have an associated Demozoo link\n", st.ByID)
			return
		}
		log.Printf("id %q does not have any empty cells that can be replaced with Demozoo data, "+
			"use --id=%v --overwrite to refetch the linked download and reapply data\n", st.ByID, st.ByID)
		return
	}
	logs.Println()
}

func (st Stat) summary(elapsed time.Duration) {
	t := fmt.Sprintf("Total Demozoo items handled: %v, time elapsed %.1f seconds", st.Count, elapsed.Seconds())
	logs.Println(strings.Repeat("─", len(t)))
	logs.Println(t)
	if st.Missing > 0 {
		logs.Println("UUID files not found:", st.Missing)
	}
}

// sumTotal calculates the total number of conditional rows.
func (st *Stat) sumTotal(rec Records, req Request) error {
	for rec.Rows.Next() {
		if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
			return fmt.Errorf("sum total rows scan: %w", err)
		}
		if n := database.IsDemozoo(rec.Values); !n && req.skip() {
			continue
		}
		st.Total++
	}
	return nil
}

// Download the first available remote file linked in the Demozoo production record.
func (r *Record) Download(overwrite bool, api *prods.ProductionsAPIv1, st Stat) (skip bool) {
	if st.FileExist(r) || overwrite {
		if r.UUID == "" {
			log.Print(color.Error.Sprint("UUID is empty, cannot continue"))
			return true
		}
		name, link := api.DownloadLink()
		if link == "" {
			logs.Print(color.Note.Sprint("no suitable downloads found"))
			return true
		}
		const OK = 200
		logs.Printcrf("%s%s %s", r.String(st.Total), color.Primary.Sprint(link), download.StatusColor(OK, "200 OK"))
		head, err := download.GetSave(r.FilePath, link)
		if err != nil {
			logs.Log(err)
			return true
		}
		logs.Printcrf(r.String(st.Total))
		logs.Printf("• %s", name)
		r.downloadReset(name)
		r.lastMod(head)
	}
	return false
}

func (r *Record) downloadReset(name string) {
	r.Filename = name
	r.Filesize = ""
	r.SumMD5 = ""
	r.Sum384 = ""
	r.FileZipContent = ""
}
