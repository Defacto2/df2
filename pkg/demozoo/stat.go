package demozoo

import (
	"database/sql"
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
	"go.uber.org/zap"
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
func (st *Stat) NextRefresh(db *sql.DB, w io.Writer, l *zap.SugaredLogger, rec Records) error {
	if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
		return fmt.Errorf("next scan: %w", err)
	}
	st.Count++
	r, err := NewRecord(l, st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 1: %w", err)
	}
	logger.Printcrf(w, r.String(0))
	var f Product
	err = f.Get(r.WebIDDemozoo)
	if err != nil {
		return fmt.Errorf("next fetch: %w", err)
	}
	var ok bool
	code, status, api := f.Code, f.Status, f.API
	if ok, err = r.confirm(db, w, code, status); err != nil {
		return fmt.Errorf("next confirm: %w", err)
	} else if !ok {
		return nil
	}
	if err = r.pouet(w, &api); err != nil {
		return fmt.Errorf("next refresh: %w", err)
	}
	r.title(w, &api)
	a := api.Authors()
	r.authors(w, &a)
	var nr Record
	nr, err = NewRecord(l, st.Count, rec.Values)
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
func (st *Stat) NextPouet(db *sql.DB, w io.Writer, l *zap.SugaredLogger, rec Records) error {
	if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
		return fmt.Errorf("next scan: %w", err)
	}
	st.Count++
	r, err := NewRecord(l, st.Count, rec.Values)
	if err != nil {
		return fmt.Errorf("next record 1: %w", err)
	}
	if r.WebIDPouet > 0 {
		return nil
	}
	logger.Printcrf(w, r.String(0))
	var f Product
	err = f.Get(r.WebIDDemozoo)
	if err != nil {
		return fmt.Errorf("next fetch: %w", err)
	}
	var ok bool
	code, status, api := f.Code, f.Status, f.API
	if ok, err = r.confirm(db, w, code, status); err != nil {
		return fmt.Errorf("next confirm: %w", err)
	} else if !ok {
		return nil
	}
	if err = r.pouet(w, &api); err != nil {
		return fmt.Errorf("next refresh: %w", err)
	}
	var nr Record
	nr, err = NewRecord(l, st.Count, rec.Values)
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
	if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
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
	t := fmt.Sprintf("Total Demozoo items handled: %v, time elapsed %.1f seconds", st.Count, elapsed.Seconds())
	fmt.Fprintln(w, strings.Repeat("─", len(t)))
	fmt.Fprintln(w, t)
	if st.Missing > 0 {
		fmt.Fprintln(w, "UUID files not found:", st.Missing)
	}
}

// sumTotal calculates the total number of conditional rows.
func (st *Stat) sumTotal(rec Records, req Request) error {
	for rec.Rows.Next() {
		if err := rec.Rows.Scan(rec.ScanArgs...); err != nil {
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
func (r *Record) Download(w io.Writer, l *zap.SugaredLogger,
	overwrite bool, api *prods.ProductionsAPIv1, st Stat,
) bool {
	if st.FileExist(r) || overwrite {
		if r.UUID == "" {
			fmt.Fprint(w, color.Error.Sprint("UUID is empty, cannot continue"))
			return true
		}
		name, link := api.DownloadLink()
		if link == "" {
			fmt.Fprint(w, color.Note.Sprint("no suitable downloads found"))
			return true
		}
		const OK = 200
		logger.Printcrf(w, "%s%s %s", r.String(st.Total), color.Primary.Sprint(link), download.StatusColor(OK, "200 OK"))
		head, err := download.GetSave(w, r.FilePath, link)
		if err != nil {
			l.Errorln(err)
			return true
		}
		logger.Printcrf(w, r.String(st.Total))
		fmt.Fprintf(w, "• %s", name)
		r.downloadReset(name)
		r.lastMod(w, head)
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
