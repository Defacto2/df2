package demozoo

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
)

var (
	ErrDir       = errors.New("filepath points to a directory")
	ErrOverwrite = errors.New("overwrite is false, but an existing download exists and cannot be overwritten")
	ErrDownload  = errors.New("no suitable downloads found")
	ErrProdAPI   = errors.New("productions api pointer cannot be nil")
	ErrRecord    = errors.New("pointer to the record cannot be nil")
	ErrRecords   = errors.New("pointer to the rows of records cannot be nil")
	ErrUUID      = errors.New("uuid is empty and cannot be used")
)

// Stat are statistics for the remote query.
type Stat struct {
	Count   int    // Count records.
	Fetched int    // Fetched records.
	Missing int    // Missing files.
	Total   int    // Total files.
	ByID    string // ByID filtering.
}

// FileExist returns false when the record FilePath points to a non-existant file.
func (st *Stat) FileExist(r *Record) (bool, error) {
	if r == nil {
		return false, ErrRecord
	}
	s, err := os.Stat(r.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			st.Missing++
			return false, nil
		}
		return false, err
	}
	if s.IsDir() {
		return false, ErrDir
	}
	return true, nil
}

// Records contain more than one Record.
type Records struct {
	Rows   *sql.Rows      // Rows is the result of a query.
	Args   []any          // Args are data types.
	Values []sql.RawBytes // Values are byte encoded values.
}

// NextRefresh iterates over the Records to update sync their Demozoo data to the database.
func (st *Stat) NextRefresh(db *sql.DB, w io.Writer, rec Records) error {
	if db == nil {
		return database.ErrDB
	}
	if rec.Rows == nil {
		return ErrRecords
	}
	if w == nil {
		w = io.Discard
	}
	if err := rec.Rows.Scan(rec.Args...); err != nil {
		return fmt.Errorf("next scan: %w", err)
	}
	st.Count++
	r, err := NewRecord(st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 1: %w", err)
	}
	logger.Printcrf(w, r.String(0))
	var f Product
	err = f.Get(r.WebIDDemozoo)
	if err != nil {
		return fmt.Errorf("next fetch: %w", err)
	}
	ok := false
	code, status, api := f.Code, f.Status, f.API
	if ok, err = r.confirm(db, w, code, status); err != nil {
		return fmt.Errorf("next confirm: %w", err)
	} else if !ok {
		return nil
	}
	if err = r.pouet(w, &api); err != nil {
		return fmt.Errorf("next pouet: %w", err)
	}
	if err := r.title(w, &api); err != nil {
		return fmt.Errorf("next title: %w", err)
	}
	a := api.Authors()
	if err := r.authors(w, &a); err != nil {
		return err
	}
	nr, err := NewRecord(st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 2: %w", err)
	}
	if reflect.DeepEqual(nr, r) {
		fmt.Fprintf(w, "• skipped %v", str.Y())
		return nil
	}
	if err = r.Save(db); err != nil {
		fmt.Fprintf(w, "• saved %v ", str.X())
		return fmt.Errorf("next save: %w", err)
	}
	fmt.Fprintf(w, "• saved %v", str.Y())
	return nil
}

// NextPouet iterates over the linked Demozoo records and sync any linked Pouet data to the local files table.
func (st *Stat) NextPouet(db *sql.DB, w io.Writer, rec Records) error {
	if db == nil {
		return database.ErrDB
	}
	if rec.Rows == nil {
		return ErrRecords
	}
	if w == nil {
		w = io.Discard
	}
	if err := rec.Rows.Scan(rec.Args...); err != nil {
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
	logger.Printcrf(w, r.String(0))
	f := Product{}
	err = f.Get(r.WebIDDemozoo)
	if err != nil {
		return fmt.Errorf("next fetch: %w", err)
	}
	ok := false
	code, status, api := f.Code, f.Status, f.API
	if ok, err = r.confirm(db, w, code, status); err != nil {
		return fmt.Errorf("next confirm: %w", err)
	} else if !ok {
		return nil
	}
	if err = r.pouet(w, &api); err != nil {
		return fmt.Errorf("next refresh: %w", err)
	}
	nr, err := NewRecord(st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 2: %w", err)
	}
	if reflect.DeepEqual(nr, r) {
		fmt.Fprintf(w, "• skipped %v", str.Y())
		return nil
	}
	if err = r.Save(db); err != nil {
		fmt.Fprintf(w, "• saved %v ", str.X())
		return fmt.Errorf("next save: %w", err)
	}
	fmt.Fprintf(w, "• saved %v", str.Y())
	return nil
}

// nextResult checks for the next new record.
func (st *Stat) nextResult(rec Records, req Request) (bool, error) {
	if rec.Rows == nil {
		return false, ErrRecords
	}
	if err := rec.Rows.Scan(rec.Args...); err != nil {
		return false, fmt.Errorf("next result rows scan: %w", err)
	}
	n, err := database.IsDemozoo(rec.Values)
	if err != nil {
		return false, err
	}
	if !n && req.skip() {
		return true, nil
	}
	st.Count++
	return false, nil
}

func (st Stat) printer(w io.Writer) {
	if w == nil {
		w = io.Discard
	}
	if st.Count == 0 {
		if st.Fetched == 0 {
			fmt.Fprintf(w, "id %q does not have an associated Demozoo link\n", st.ByID)
			return
		}
		fmt.Fprintf(w, "id %q does not have any empty cells that can be replaced with Demozoo data, "+
			"use --id=%v --overwrite to refetch the linked download and reapply data\n", st.ByID, st.ByID)
		return
	}
	fmt.Fprintln(w)
}

func (st Stat) summary(w io.Writer, elapsed time.Duration) {
	if w == nil {
		w = io.Discard
	}
	t := fmt.Sprintf("Total Demozoo items handled: %v, time elapsed %.1f seconds", st.Count, elapsed.Seconds())
	fmt.Fprintln(w, strings.Repeat("─", len(t)))
	fmt.Fprintln(w, t)
	if st.Missing > 0 {
		fmt.Fprintln(w, "UUID files not found:", st.Missing)
	}
}

// sumTotal calculates the total number of conditional rows.
func (st *Stat) sumTotal(rec Records, req Request) error {
	if rec.Rows == nil {
		return ErrRecords
	}
	for rec.Rows.Next() {
		if err := rec.Rows.Scan(rec.Args...); err != nil {
			return fmt.Errorf("sum total rows scan: %w", err)
		}
		n, err := database.IsDemozoo(rec.Values)
		if err != nil {
			return err
		}
		if !n && req.skip() {
			continue
		}
		st.Total++
	}
	return nil
}

// Download the first available remote file linked in the Demozoo production record.
func (r *Record) Download(w io.Writer, api *prods.ProductionsAPIv1, st Stat, overwrite bool) error {
	if api == nil {
		return ErrProdAPI
	}
	if w == nil {
		w = io.Discard
	}
	exist, err := st.FileExist(r)
	if err != nil {
		return err
	}
	if exist && !overwrite {
		return ErrOverwrite
	}
	if r.UUID == "" {
		return ErrUUID
	}
	dw := io.Discard
	if flag.Lookup("test.v") != nil {
		dw = os.Stdout
	}
	name, link := api.DownloadLink(dw)
	if link == "" {
		return ErrDownload
	}
	const OK = 200
	logger.Printcrf(w, "%s%s %s", r.String(st.Total), color.Primary.Sprint(link),
		download.StatusColor(OK, "200 OK"))
	head, err := download.GetSave(w, r.FilePath, link)
	if err != nil {
		return err
	}
	logger.Printcrf(w, r.String(st.Total))
	fmt.Fprintf(w, "• %s", name)
	r.downloadReset(name)
	return r.lastMod(w, head)
}

func (r *Record) downloadReset(name string) {
	r.Filename = name
	r.Filesize = ""
	r.SumMD5 = ""
	r.Sum384 = ""
	r.FileZipContent = ""
}
