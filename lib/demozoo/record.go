package demozoo

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/demozoo/internal/fix"
	"github.com/Defacto2/df2/lib/demozoo/internal/prods"
	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
)

const sep = ","

// Record of a file item.
type Record struct {
	Count          int
	FilePath       string // absolute path to file
	ID             string // MySQL auto increment id
	UUID           string // record unique id
	Filename       string
	Filesize       string
	FileZipContent string
	CreatedAt      string
	UpdatedAt      string
	SumMD5         string // file download MD5 hash
	Sum384         string // file download SHA384 hash
	Readme         string
	DOSeeBinary    string
	Platform       string
	GroupFor       string
	GroupBy        string
	Title          string
	Section        string
	CreditText     []string
	CreditCode     []string
	CreditArt      []string
	CreditAudio    []string
	WebIDDemozoo   uint // demozoo production id
	WebIDPouet     uint
	LastMod        time.Time // file download last modified time
}

func (r *Record) String(total int) string {
	const leadingZeros = 4
	// calculate the number of prefixed zero characters
	d := leadingZeros
	if total > 0 {
		d = len(strconv.Itoa(total))
	}
	return fmt.Sprintf("%s %0*d. %v (%v) %v",
		color.Question.Sprint("→"), d, r.Count, color.Primary.Sprint(r.ID),
		color.Info.Sprint(r.WebIDDemozoo),
		r.CreatedAt)
}

// Save the record to the database.
func (r *Record) Save() error {
	db := database.Connect()
	defer db.Close()
	query, args := r.SQL()
	update, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("save prepare: %w", err)
	}
	defer update.Close()
	_, err = update.Exec(args...)
	if err != nil {
		return fmt.Errorf("save exec: %w", err)
	}
	return nil
}

func (r *Record) SQL() (query string, args []interface{}) {
	// an range map iternation is not used due to the varied comparisons
	set := setSQL(r)
	args = setArg(r, set)
	if len(set) == 0 {
		return "", args
	}
	set = append(set, "updatedat=?")
	args = append(args, []interface{}{time.Now()}...)
	set = append(set, "updatedby=?")
	args = append(args, []interface{}{database.UpdateID}...)
	query = "UPDATE files SET " + strings.Join(set, sep) + " WHERE id=?"
	args = append(args, []interface{}{r.ID}...)
	return query, args
}

func setSQL(r *Record) []string {
	var set []string
	if r.Filename != "" {
		set = append(set, "filename=?")
	}
	if r.Filesize != "" {
		set = append(set, "filesize=?")
	}
	if r.FileZipContent != "" {
		set = append(set, "file_zip_content=?")
	}
	if r.SumMD5 != "" {
		set = append(set, "file_integrity_weak=?")
	}
	if r.Sum384 != "" {
		set = append(set, "file_integrity_strong=?")
	}
	const errYear = 0o001
	if r.LastMod.Year() != errYear {
		set = append(set, "file_last_modified=?")
	}
	if r.WebIDPouet != 0 {
		set = append(set, "web_id_pouet=?")
	}
	if r.WebIDDemozoo == 0 && len(set) > 0 {
		set = append(set, "web_id_demozoo=?")
	}
	if r.DOSeeBinary != "" {
		set = append(set, "dosee_run_program=?")
	}
	if r.Readme != "" {
		set = append(set, "retrotxt_readme=?")
	}
	if r.Title != "" {
		set = append(set, "record_title=?")
	}
	set = append(set, setCredit(r)...)
	if r.Platform != "" {
		set = append(set, "platform=?")
	}
	return set
}

func setCredit(r *Record) []string {
	var set []string
	if len(r.CreditText) > 0 {
		set = append(set, "credit_text=?")
	}
	if len(r.CreditCode) > 0 {
		set = append(set, "credit_program=?")
	}
	if len(r.CreditArt) > 0 {
		set = append(set, "credit_illustration=?")
	}
	if len(r.CreditAudio) > 0 {
		set = append(set, "credit_audio=?")
	}
	return set
}

func setArg(r *Record, set []string) (args []interface{}) {
	if r.Filename != "" {
		args = append(args, []interface{}{r.Filename}...)
	}
	if r.Filesize != "" {
		args = append(args, []interface{}{r.Filesize}...)
	}
	if r.FileZipContent != "" {
		args = append(args, []interface{}{r.FileZipContent}...)
	}
	if r.SumMD5 != "" {
		args = append(args, []interface{}{r.SumMD5}...)
	}
	if r.Sum384 != "" {
		args = append(args, []interface{}{r.Sum384}...)
	}
	const errYear = 0o001
	if r.LastMod.Year() != errYear {
		args = append(args, []interface{}{r.LastMod}...)
	}
	if r.WebIDPouet != 0 {
		args = append(args, []interface{}{r.WebIDPouet}...)
	}
	if r.WebIDDemozoo == 0 && len(set) > 0 {
		args = append(args, []interface{}{""}...)
	}
	if r.DOSeeBinary != "" {
		args = append(args, []interface{}{r.DOSeeBinary}...)
	}
	if r.Readme != "" {
		args = append(args, []interface{}{r.Readme}...)
	}
	if r.Title != "" {
		args = append(args, []interface{}{r.Title}...)
	}
	args = append(args, setCredits(r)...)
	if r.Platform != "" {
		args = append(args, []interface{}{r.Platform}...)
	}
	return args
}

func setCredits(r *Record) (args []interface{}) {
	if len(r.CreditText) > 0 {
		j := strings.Join(r.CreditText, sep)
		args = append(args, []interface{}{j}...)
	}
	if len(r.CreditCode) > 0 {
		j := strings.Join(r.CreditCode, sep)
		args = append(args, []interface{}{j}...)
	}
	if len(r.CreditArt) > 0 {
		j := strings.Join(r.CreditArt, sep)
		args = append(args, []interface{}{j}...)
	}
	if len(r.CreditAudio) > 0 {
		j := strings.Join(r.CreditAudio, sep)
		args = append(args, []interface{}{j}...)
	}
	return args
}

// ZipContent reads an archive and saves its content to the database.
func (r *Record) ZipContent() (ok bool, err error) {
	if r.FilePath == "" {
		return false, fmt.Errorf("zipcontent: %w", ErrFilePath)
	} else if r.Filename == "" {
		return false, fmt.Errorf("zipcontent: %w", ErrFilename)
	}
	a, err := archive.Read(r.FilePath, r.Filename)
	if err != nil {
		return false, fmt.Errorf("zipcontent read: %w", err)
	}
	r.FileZipContent = strings.Join(a, "\n")
	return true, nil
}

func (st *Stat) FileExist(r *Record) (missing bool) {
	if s, err := os.Stat(r.FilePath); os.IsNotExist(err) || s.IsDir() {
		st.Missing++
		return true
	}
	return false
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
			return true, fmt.Errorf("confirm: %w", err)
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

// Fix repairs imported Demozoo data conflicts.
func Fix() error {
	return fix.Configs()
}

type Records struct {
	Rows     *sql.Rows
	ScanArgs []interface{}
	Values   []sql.RawBytes
}

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
	f, err := Fetch(r.WebIDDemozoo)
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

// RefreshMeta synchronises missing file entries with Demozoo sourced metadata.
func RefreshMeta() error {
	start := time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(selectByID(""))
	if err != nil {
		return fmt.Errorf("meta query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("meta rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("meta columns: %w", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// fetch the rows
	var st Stat
	for rows.Next() {
		if err := st.NextRefresh(Records{rows, scanArgs, values}); err != nil {
			logs.Println(fmt.Errorf("meta rows: %w", err))
		}
	}
	st.summary(time.Since(start))
	return nil
}
