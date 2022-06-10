package demozoo

import (
	"crypto/md5"
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/fix"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
)

var (
	ErrFilePath = errors.New("filepath requirement cannot be empty")
	ErrFilename = errors.New("filename requirement cannot be empty")
	ErrTooFew   = errors.New("too few record values")
	ErrNA       = errors.New("this feature is not implemented")
	ErrNoRel    = errors.New("no productions exist for this releaser")
)

const (
	dos = "dos"
	win = "windows"
	sep = ","
)

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

func (r *Record) SQL() (query string, args []any) {
	// a range map iternation is not used due to the varied comparisons
	set := setSQL(r)
	args = setArgs(r, set)
	if len(set) == 0 {
		return "", args
	}
	set = append(set, "updatedat=?")
	args = append(args, []any{time.Now()}...)
	set = append(set, "updatedby=?")
	args = append(args, []any{database.UpdateID}...)
	query = "UPDATE files SET " + strings.Join(set, sep) + " WHERE id=?"
	args = append(args, []any{r.ID}...)
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

func setArgs(r *Record, set []string) (args []any) {
	if r.Filename != "" {
		args = append(args, []any{r.Filename}...)
	}
	if r.Filesize != "" {
		args = append(args, []any{r.Filesize}...)
	}
	if r.FileZipContent != "" {
		args = append(args, []any{r.FileZipContent}...)
	}
	if r.SumMD5 != "" {
		args = append(args, []any{r.SumMD5}...)
	}
	if r.Sum384 != "" {
		args = append(args, []any{r.Sum384}...)
	}
	const errYear = 0o001
	if r.LastMod.Year() != errYear {
		args = append(args, []any{r.LastMod}...)
	}
	if r.WebIDPouet != 0 {
		args = append(args, []any{r.WebIDPouet}...)
	}
	if r.WebIDDemozoo == 0 && len(set) > 0 {
		args = append(args, []any{sql.NullInt16{}}...)
	}
	if r.DOSeeBinary != "" {
		args = append(args, []any{r.DOSeeBinary}...)
	}
	if r.Readme != "" {
		args = append(args, []any{r.Readme}...)
	}
	if r.Title != "" {
		args = append(args, []any{r.Title}...)
	}
	args = append(args, setCredits(r)...)
	if r.Platform != "" {
		args = append(args, []any{r.Platform}...)
	}
	return args
}

func setCredits(r *Record) (args []any) {
	if len(r.CreditText) > 0 {
		j := strings.Join(r.CreditText, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditCode) > 0 {
		j := strings.Join(r.CreditCode, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditArt) > 0 {
		j := strings.Join(r.CreditArt, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditAudio) > 0 {
		j := strings.Join(r.CreditAudio, sep)
		args = append(args, []any{j}...)
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
	ScanArgs []any
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
	scanArgs := make([]any, len(values))
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

func RefreshPouet() error {
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
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// fetch the rows
	var st Stat
	for rows.Next() {
		if err := st.NextPouet(Records{rows, scanArgs, values}); err != nil {
			logs.Println(fmt.Errorf("meta rows: %w", err))
		}
	}
	st.summary(time.Since(start))
	return nil
}

func (r *Record) FileMeta() error {
	stat, err := os.Stat(r.FilePath)
	if err != nil {
		return fmt.Errorf("record file meta stat: %w", err)
	}
	r.Filesize = strconv.Itoa(int(stat.Size()))
	// file hashes
	f, err := os.Open(r.FilePath)
	if err != nil {
		return fmt.Errorf("record file meta open: %w", err)
	}
	defer f.Close()
	h1 := md5.New() // nolint: gosec
	if _, err := io.Copy(h1, f); err != nil {
		return fmt.Errorf("record file meta io copy for the md5 hash: %w", err)
	}
	r.SumMD5 = fmt.Sprintf("%x", h1.Sum(nil))
	h2 := sha512.New384()
	if _, err := io.Copy(h2, f); err != nil {
		return fmt.Errorf("record file meta io copy for the sha512 hash: %w", err)
	}
	r.Sum384 = fmt.Sprintf("%x", h2.Sum(nil))
	return nil
}

func (r *Record) DoseeMeta() error {
	names, err := r.variations()
	if err != nil {
		return fmt.Errorf("record dosee meta: %w", err)
	}
	d, err := archive.Demozoo(r.FilePath, r.UUID, &names)
	if err != nil {
		return fmt.Errorf("record dosee meta: %w", err)
	}
	if strings.EqualFold(r.Platform, dos) && d.DOSee != "" {
		r.DOSeeBinary = d.DOSee
	}
	if d.NFO != "" {
		r.Readme = d.NFO
	}
	return nil
}

// Category are tags for production imports.
type Category int

const (
	Text     Category = iota // Text based files.
	Code                     // Code are binary files.
	Graphics                 // Graphics are images.
	Music                    // Music is audio.
	Magazine                 // Magazine are publications.
)

func (c Category) String() string {
	return [...]string{"text", "code", "graphics", "music", "magazine"}[c]
}

// check record to see if it needs updating.
func (r *Record) check() (update bool) {
	switch {
	case
		r.Filename == "",
		r.Platform == "",
		r.Filesize == "",
		r.Sum384 == "",
		r.SumMD5 == "",
		r.FileZipContent == "":
		return true
	default:
		logs.Printf("skipped, no changes needed %v", str.Y())
		return false
	}
}

// Stat are the remote query statistics.
type Stat struct {
	Count   int
	Fetched int
	Missing int
	Total   int
	ByID    string
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
			fmt.Printf("id %q does not have an associated Demozoo link\n", st.ByID)
			return
		}
		fmt.Printf("id %q does not have any empty cells that can be replaced with Demozoo data, "+
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

func (r *Record) Download(overwrite bool, api *prods.ProductionsAPIv1, st Stat) (skip bool) {
	if st.FileExist(r) || overwrite {
		if r.UUID == "" {
			fmt.Print(color.Error.Sprint("UUID is empty, cannot continue"))
			return true
		}
		name, link := api.DownloadLink()
		if link == "" {
			logs.Print(color.Note.Sprint("no suitable downloads found"))
			return true
		}
		const OK = 200
		logs.Printcrf("%s%s %s", r.String(st.Total), color.Primary.Sprint(link), download.StatusColor(OK, "200 OK"))
		head, err := download.Get(r.FilePath, link)
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

// last modified time passed via HTTP.
func (r *Record) lastMod(head http.Header) {
	lm := head.Get("Last-Modified")
	if len(lm) < 1 {
		return
	}
	t, err := time.Parse(download.RFC5322, lm)
	if err != nil {
		logs.Printf(" • last-mod value %q ?", lm)
		return
	}
	r.LastMod = t
	if time.Now().Year() == t.Year() {
		logs.Printf(" • %s", t.Format("2 Jan"))
		return
	}
	logs.Printf(" • %s", t.Format("Jan 06"))
}

// parseAPI confirms and parses the API request.
func (r *Record) parseAPI(st Stat, overwrite bool, storage string) (skip bool, err error) {
	if database.CheckUUID(r.Filename) == nil {
		// handle anomaly where the Filename was incorrectly given UUID
		fmt.Println("Clearing filename which is incorrectly set as", r.Filename)
		r.Filename = ""
	}
	var f Product
	err = f.Get(r.WebIDDemozoo)
	if err != nil {
		return true, fmt.Errorf("parse api fetch: %w", err)
	}
	code, status, api := f.Code, f.Status, f.API
	if ok, err := r.confirm(code, status); err != nil {
		return true, parseAPIErr(err)
	} else if !ok {
		return true, nil
	}
	if err := r.pingPouet(&api); err != nil {
		return true, parseAPIErr(err)
	}
	r.FilePath = filepath.Join(storage, r.UUID)
	if skip := r.Download(overwrite, &api, st); skip {
		return true, nil
	}
	if update := r.check(); !update {
		return true, nil
	}
	if r.Platform == "" {
		r.platform(&api)
	}
	return r.parse(&api)
}

func parseAPIErr(err error) error {
	return fmt.Errorf("%s%w", "parse api: ", err)
}

func (r *Record) parse(api *prods.ProductionsAPIv1) (bool, error) {
	switch {
	case r.Filename == "":
		// handle an unusual case where filename is missing but all other metadata exists
		if n, _ := api.DownloadLink(); n != "" {
			fmt.Print(n)
			r.Filename = n
			r.save()
		} else {
			fmt.Println("could not find a suitable value for the required filename column")
			return true, nil
		}
		fallthrough
	case
		r.Filesize == "",
		r.SumMD5 == "",
		r.Sum384 == "":
		if err := r.FileMeta(); err != nil {
			return true, parseAPIErr(err)
		}
		r.save()
		fallthrough
	case r.FileZipContent == "":
		if zip, err := r.ZipContent(); err != nil {
			return true, parseAPIErr(err)
		} else if zip {
			if err := r.DoseeMeta(); err != nil {
				return true, parseAPIErr(err)
			}
		}
		r.save()
	}
	return false, nil
}

func (r *Record) pingPouet(api *prods.ProductionsAPIv1) error {
	const success = 299
	if id, code, err := api.PouetID(true); err != nil {
		return fmt.Errorf("ping pouet: %w", err)
	} else if id > 0 && code <= success {
		r.WebIDPouet = uint(id)
	}
	return nil
}

func (r *Record) platform(api *prods.ProductionsAPIv1) {
	const msdos, windows = 4, 1
	for _, p := range api.Platforms {
		switch p.ID {
		case msdos:
			r.Platform = dos
		case windows:
			r.Platform = win
		default:
			continue
		}
	}
}

func (r *Record) save() {
	if err := r.Save(); err != nil {
		logs.Printf(" %v \n", str.X())
		logs.Log(err)
		return
	}
	logs.Printf(" 💾%v", str.Y())
}

func (r *Record) variations() ([]string, error) {
	names := []string{}
	if r.GroupBy != "" {
		v, err := groups.Variations(r.GroupBy)
		if err != nil {
			return nil, fmt.Errorf("record group by variations: %w", err)
		}
		names = append(names, v...)
	}
	if r.GroupFor != "" {
		v, err := groups.Variations(r.GroupFor)
		if err != nil {
			return nil, fmt.Errorf("record group for variations: %w", err)
		}
		names = append(names, v...)
	}
	return names, nil
}
