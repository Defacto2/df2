package demozoo

import (
	"context"
	"crypto/md5" //nolint:gosec
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
	"github.com/Defacto2/df2/pkg/demozoo/internal/insert"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
	"go.uber.org/zap"
)

var (
	ErrFilePath = errors.New("filepath requirement cannot be empty")
	ErrFilename = errors.New("filename requirement cannot be empty")
	ErrTooFew   = errors.New("too few record values")
	ErrNA       = errors.New("this feature is not implemented")
	ErrNoRel    = errors.New("no productions exist for this releaser")
)

func apiErr(err error) error {
	return fmt.Errorf("%s%w", "parse api: ", err)
}

const (
	dos     = "dos"
	win     = "windows"
	sep     = ","
	timeout = 5 * time.Second
)

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

// Record update for an item in the "file" table of the database.
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
	const leadZeros = 4
	// calculate the number of prefixed zero characters
	d := leadZeros
	if total > 0 {
		d = len(strconv.Itoa(total))
	}
	return fmt.Sprintf("%s %0*d. %v (%v) %v",
		color.Question.Sprint("→"), d, r.Count, color.Primary.Sprint(r.ID),
		color.Info.Sprint(r.WebIDDemozoo),
		r.CreatedAt)
}

// DoseeMeta generates DOSee related metadata from the file archive.
func (r *Record) DoseeMeta(db *sql.DB, w io.Writer) error {
	names, err := r.variations()
	if err != nil {
		return fmt.Errorf("record dosee meta: %w", err)
	}
	d, err := archive.Demozoo(db, w, r.FilePath, r.UUID, &names)
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

// FileMeta generates metadata from the file archive.
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
	h1 := md5.New() //nolint: gosec
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

// Save the record to the database.
func (r *Record) Save(db *sql.DB) error {
	query, args := r.Stmt()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("save prepare: %w", err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return fmt.Errorf("save exec: %w", err)
	}
	return nil
}

// Stmt creates the SQL prepare statement and values to update a Demozoo production.
func (r *Record) Stmt() (string, []any) {
	// a range map iternation is not used due to the varied comparisons
	set, args := updates(r)
	if len(set) == 0 {
		return "", args
	}
	set = append(set, "updatedat=?")
	args = append(args, []any{time.Now()}...)
	set = append(set, "updatedby=?")
	args = append(args, []any{database.UpdateID}...)
	query := "UPDATE files SET " + strings.Join(set, sep) + " WHERE id=?"
	args = append(args, []any{r.ID}...)
	return query, args
}

// ZipContent reads an archive and saves its content to the database.
func (r *Record) ZipContent(w io.Writer) (bool, error) {
	if r.FilePath == "" {
		return false, fmt.Errorf("zipcontent: %w", ErrFilePath)
	} else if r.Filename == "" {
		return false, fmt.Errorf("zipcontent: %w", ErrFilename)
	}
	a, fn, err := archive.Read(w, r.FilePath, r.Filename)
	if err != nil {
		return false, fmt.Errorf("zipcontent read: %w", err)
	}
	r.FileZipContent = strings.Join(a, "\n")
	r.Filename = fn
	return true, nil
}

// InsertProds adds the collection of Demozoo productions to the file database.
func InsertProds(db *sql.DB, w io.Writer, p *releases.Productions) error {
	return insert.Prods(db, w, p)
}

func updates(r *Record) ([]string, []any) {
	set, args := file(r)
	if r.WebIDPouet != 0 {
		set = append(set, "web_id_pouet=?")
		args = append(args, []any{r.WebIDPouet}...)
	}
	if r.WebIDDemozoo == 0 && len(set) > 0 {
		set = append(set, "web_id_demozoo=?")
		args = append(args, []any{sql.NullInt16{}}...)
	}
	if r.DOSeeBinary != "" {
		set = append(set, "dosee_run_program=?")
		args = append(args, []any{r.DOSeeBinary}...)
	}
	if r.Readme != "" {
		set = append(set, "retrotxt_readme=?")
		args = append(args, []any{r.Readme}...)
	}
	if r.Title != "" {
		set = append(set, "record_title=?")
		args = append(args, []any{r.Title}...)
	}
	if r.Platform != "" {
		set = append(set, "platform=?")
		args = append(args, []any{r.Platform}...)
	}
	s, a := credits(r)
	set = append(set, s...)
	args = append(args, a...)
	return set, args
}

func file(r *Record) ([]string, []any) {
	var args []any
	set := []string{}
	if r.Filename != "" {
		set = append(set, "filename=?")
		args = append(args, []any{r.Filename}...)
	}
	if r.Filesize != "" {
		set = append(set, "filesize=?")
		args = append(args, []any{r.Filesize}...)
	}
	if r.FileZipContent != "" {
		set = append(set, "file_zip_content=?")
		args = append(args, []any{r.FileZipContent}...)
	}
	if r.SumMD5 != "" {
		set = append(set, "file_integrity_weak=?")
		args = append(args, []any{r.SumMD5}...)
	}
	if r.Sum384 != "" {
		set = append(set, "file_integrity_strong=?")
		args = append(args, []any{r.Sum384}...)
	}
	const errYear = 0o001
	if r.LastMod.Year() != errYear {
		set = append(set, "file_last_modified=?")
		args = append(args, []any{r.LastMod}...)
	}
	return set, args
}

func credits(r *Record) ([]string, []any) {
	var args []any
	set := []string{}
	if len(r.CreditText) > 0 {
		set = append(set, "credit_text=?")
		j := strings.Join(r.CreditText, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditCode) > 0 {
		set = append(set, "credit_program=?")
		j := strings.Join(r.CreditCode, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditArt) > 0 {
		set = append(set, "credit_illustration=?")
		j := strings.Join(r.CreditArt, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditAudio) > 0 {
		set = append(set, "credit_audio=?")
		j := strings.Join(r.CreditAudio, sep)
		args = append(args, []any{j}...)
	}
	return set, args
}

func (r *Record) authors(w io.Writer, a *prods.Authors) {
	compare := func(n, o []string, i string) bool {
		if !reflect.DeepEqual(n, o) {
			fmt.Fprintf(w, "c%s:%s ", i, color.Secondary.Sprint(n))
			if len(o) > 1 {
				fmt.Fprintf(w, "%s ", color.Danger.Sprint(o))
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

// check record to see if it needs updating.
func (r *Record) check(w io.Writer) bool {
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
		fmt.Fprintf(w, "skipped, no changes needed %v", str.Y())
		return false
	}
}

func (r *Record) confirm(db *sql.DB, w io.Writer, code int, status string) (bool, error) {
	const nofound, found, problems = 404, 200, 300
	if code == nofound {
		r.WebIDDemozoo = 0
		if err := r.Save(db); err != nil {
			return true, fmt.Errorf("confirm: %w", err)
		}
		fmt.Fprintf(w, "(%s)\n", download.StatusColor(code, status))
		return false, nil
	}
	if code < found || code >= problems {
		fmt.Fprintf(w, "(%s)\n", download.StatusColor(code, status))
		return false, nil
	}
	return true, nil
}

// last modified time passed via HTTP.
func (r *Record) lastMod(w io.Writer, head http.Header) {
	lm := head.Get("Last-Modified")
	if len(lm) < 1 {
		return
	}
	t, err := time.Parse(download.RFC5322, lm)
	if err != nil {
		fmt.Fprintf(w, " • last-mod value %q ?", lm)
		return
	}
	r.LastMod = t
	if time.Now().Year() == t.Year() {
		fmt.Fprintf(w, " • %s", t.Format("2 Jan"))
		return
	}
	fmt.Fprintf(w, " • %s", t.Format("Jan 06"))
}

func (r *Record) parse(db *sql.DB, w io.Writer, l *zap.SugaredLogger, api *prods.ProductionsAPIv1) (bool, error) {
	switch {
	case r.Filename == "":
		// handle an unusual case where filename is missing but all other metadata exists
		if n, _ := api.DownloadLink(); n != "" {
			fmt.Fprint(w, n)
			r.Filename = n
			r.save(db, w, l)
		} else {
			fmt.Fprintln(w, "could not find a suitable value for the required filename column")
			return true, nil
		}
		fallthrough
	case
		r.Filesize == "",
		r.SumMD5 == "",
		r.Sum384 == "":
		if err := r.FileMeta(); err != nil {
			return true, apiErr(err)
		}
		r.save(db, w, l)
		fallthrough
	case r.FileZipContent == "":
		if zip, err := r.ZipContent(w); err != nil {
			return true, apiErr(err)
		} else if zip {
			if err := r.DoseeMeta(db, w); err != nil {
				return true, apiErr(err)
			}
		}
		r.save(db, w, l)
	}
	return false, nil
}

// parseAPI confirms and parses the API request.
func (r *Record) parseAPI(db *sql.DB, w io.Writer, l *zap.SugaredLogger, st Stat, overwrite bool, storage string) (bool, error) {
	if database.CheckUUID(r.Filename) == nil {
		// handle anomaly where the Filename was incorrectly given UUID
		fmt.Fprintln(w, "Clearing filename which is incorrectly set as", r.Filename)
		r.Filename = ""
	}
	var f Product
	if err := f.Get(r.WebIDDemozoo); err != nil {
		return true, fmt.Errorf("parse api fetch: %w", err)
	}
	code, status, api := f.Code, f.Status, f.API
	if ok, err := r.confirm(db, w, code, status); err != nil {
		return true, apiErr(err)
	} else if !ok {
		return true, nil
	}
	if err := r.pingPouet(&api); err != nil {
		return true, apiErr(err)
	}
	r.FilePath = filepath.Join(storage, r.UUID)
	if skip := r.Download(w, l, overwrite, &api, st); skip {
		return true, nil
	}
	if update := r.check(w); !update {
		return true, nil
	}
	if r.Platform == "" {
		r.platform(&api)
	}
	return r.parse(db, w, l, &api)
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

func (r *Record) pouet(w io.Writer, api *prods.ProductionsAPIv1) error {
	pid, _, err := api.PouetID(false)
	if err != nil {
		return fmt.Errorf("pouet: %w", err)
	}
	if r.WebIDPouet != uint(pid) {
		r.WebIDPouet = uint(pid)
		fmt.Fprintf(w, "PN:%s ", color.Note.Sprint(pid))
	}
	return nil
}

func (r *Record) save(db *sql.DB, w io.Writer, l *zap.SugaredLogger) {
	if err := r.Save(db); err != nil {
		fmt.Fprintf(w, " %v \n", str.X())
		l.Errorln(err)
		return
	}
	fmt.Fprintf(w, " 💾%v", str.Y())
}

func (r *Record) title(w io.Writer, api *prods.ProductionsAPIv1) {
	if r.Section != Magazine.String() && !strings.EqualFold(r.Title, api.Title) {
		fmt.Fprintf(w, "i:%s ", color.Secondary.Sprint(api.Title))
		r.Title = api.Title
	}
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
