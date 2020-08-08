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
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

type Category int

func (c Category) String() string {
	switch c {
	case Text:
		return "text"
	case Code:
		return "code"
	case Graphics:
		return "graphics"
	case Music:
		return "music"
	case Magazine:
		return "magazine"
	}
	return ""
}

const (
	Text Category = iota
	Code
	Graphics
	Music
	Magazine
)

func category(c string) Category {
	switch strings.ToLower(c) {
	case Text.String():
		return Text
	case Code.String():
		return Code
	case Graphics.String():
		return Graphics
	case Music.String():
		return Music
	case Magazine.String():
		return Magazine
	}
	return -1
}

const (
	api = "https://demozoo.org/api/v1/productions"
	cd  = "Content-Disposition"
	cls = "PouetProduction"
	df2 = "defacto2.net"
	dos = "dos"
	win = "windows"
)

const selectSQL = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`,`web_id_demozoo`,`file_zip_content`," +
	"`updatedat`,`platform`,`file_integrity_strong`,`file_integrity_weak`,`web_id_pouet`,`group_brand_for`," +
	"`group_brand_by`,`record_title`,`section`,`credit_illustration`,`credit_audio`,`credit_program`,`credit_text`"

var (
	ErrRecordCnt  = errors.New("unexpected number of record values")
	ErrNegativeID = errors.New("demozoo production id cannot be a negative integer")
	ErrFilePath   = errors.New("filepath requirement cannot be empty")
	ErrFilename   = errors.New("filename requirement cannot be empty")
	ErrNoFile     = errors.New("file cannot be found")
)

var requestedID = ""

type Fetched struct {
	Code   int
	Status string
	API    ProductionsAPIv1
}

// Fetch a Demozoo production by its ID.
func Fetch(id uint) (Fetched, error) {
	d := Production{ID: int64(id)}
	api, err := d.data()
	if err != nil {
		return Fetched{}, fmt.Errorf("fetched %d: %w", id, err)
	}
	return Fetched{Code: d.StatusCode, Status: d.Status, API: api}, nil
}

// Request proofs.
type Request struct {
	All       bool // parse all demozoo entries
	Overwrite bool // overwrite existing files
	Refresh   bool // refresh all demozoo entries
	Simulate  bool // simulate database save
}

// Query parses a single Demozoo entry.
func (req Request) Query(id string) (err error) {
	if err = database.CheckID(id); err != nil {
		return fmt.Errorf("request query id %s: %w", id, err)
	}
	requestedID = id
	if err := req.Queries(); err != nil {
		return fmt.Errorf("request query queries: %w", err)
	}
	return nil
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (req Request) Queries() error {
	var st stat
	stmt, start := selectByID(), time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(stmt)
	if err != nil {
		return fmt.Errorf("request queries query 1: %w", err)
	} else if err = rows.Err(); err != nil {
		return fmt.Errorf("request queries rows 1: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("request queries columns: %w", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	storage := directories.Init(false).UUID
	if err = st.sumTotal(records{rows, scanArgs, values}, req); err != nil {
		return fmt.Errorf("req queries sum total: %w", err)
	}
	if st.total > 1 {
		logs.Println("Total records", st.total)
	}
	rows, err = db.Query(stmt)
	if err != nil {
		return fmt.Errorf("request queries query 2: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("request queries rows 2: %w", rows.Err())
	}
	defer rows.Close()
	for rows.Next() {
		fmt.Println()
		st.fetched++
		if skip, err := st.nextResult(records{rows, scanArgs, values}, req); err != nil {
			logs.Danger(fmt.Errorf("request queries next row: %w", err))
			continue
		} else if skip {
			continue
		}
		r, err := newRecord(st.count, values)
		if err != nil {
			logs.Danger(fmt.Errorf("request queries new record: %w", err))
			continue
		}
		logs.Printcrf(r.String(st.total))
		if update := r.check(); !update {
			continue
		}
		if skip, err := r.parseAPI(st, req.Overwrite, storage); err != nil {
			logs.Danger(fmt.Errorf("request queries parse api: %w", err))
			continue
		} else if skip {
			continue
		}
		switch {
		case st.total == 0:
			break
		case req.Simulate:
			logs.Printf(" • dry-run %v", str.Y())
		default:
			r.save()
		}
	}
	if requestedID != "" {
		st.print()
		return nil
	}
	st.summary(time.Since(start))
	return nil
}

func (req Request) flags() (skip bool) {
	if !req.All && !req.Refresh && !req.Overwrite {
		return true
	}
	return false
}

// query statistics.
type stat struct {
	count   int
	fetched int
	missing int
	total   int
}

// nextResult checks for the next, new record.
func (st *stat) nextResult(rec records, req Request) (skip bool, err error) {
	if err := rec.rows.Scan(rec.scanArgs...); err != nil {
		return false, fmt.Errorf("next result rows scan: %w", err)
	}
	if n := database.IsNew(rec.values); !n && req.flags() {
		return true, nil
	}
	st.count++
	return false, nil
}

func (st stat) print() {
	if st.count == 0 {
		var t string
		if st.fetched == 0 {
			t = fmt.Sprintf("id %q is not a Demozoo sourced file record", requestedID)
		} else {
			t = fmt.Sprintf("id %q is not a new Demozoo record, use --id=%v --overwrite to refetch the download and data", requestedID, requestedID)
		}
		logs.Println(t)
	} else {
		logs.Println()
	}
}

func (st stat) summary(elapsed time.Duration) {
	t := fmt.Sprintf("Total Demozoo items handled: %v, time elapsed %.1f seconds", st.count, elapsed.Seconds())
	logs.Println("\n" + strings.Repeat("─", len(t)))
	logs.Println(t)
	if st.missing > 0 {
		logs.Println("UUID files not found:", st.missing)
	}
}

// sumTotal calculates the total number of conditional rows.
func (st *stat) sumTotal(rec records, req Request) error {
	for rec.rows.Next() {
		if err := rec.rows.Scan(rec.scanArgs...); err != nil {
			return fmt.Errorf("sum total rows scan: %w", err)
		}
		if n := database.IsNew(rec.values); !n && req.flags() {
			continue
		}
		st.total++
	}
	return nil
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

func (r *Record) download(overwrite bool, api *ProductionsAPIv1, st stat) (skip bool) {
	if st.fileExist(r) || overwrite {
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
		logs.Printcrf("%s%s %s", r.String(st.total), color.Primary.Sprint(link), download.StatusColor(OK, "200 OK"))
		head, err := download.LinkDownload(r.FilePath, link)
		if err != nil {
			logs.Log(err)
			return true
		}
		logs.Printcrf(r.String(st.total))
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

func (r *Record) doseeMeta() error {
	names, err := r.variations()
	if err != nil {
		return fmt.Errorf("record dosee meta: %w", err)
	}
	d, err := archive.ExtractDemozoo(r.FilePath, r.UUID, &names)
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

func (r *Record) fileMeta() (err error) {
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
	h1 := md5.New()
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
	} else {
		logs.Printf(" • %s", t.Format("Jan 06"))
	}
}

// parseAPI confirms and parses the API request.
func (r *Record) parseAPI(st stat, overwrite bool, storage string) (skip bool, err error) {
	if database.CheckUUID(r.Filename) == nil {
		// handle anomaly where the Filename was incorrectly given UUID
		fmt.Println("Clearing filename which is incorrectly set as", r.Filename)
		r.Filename = ""
	}
	f, err := Fetch(r.WebIDDemozoo)
	if err != nil {
		return true, fmt.Errorf("parse api fetch: %w", err)
	}
	code, status, api := f.Code, f.Status, f.API
	if ok, err := r.confirm(code, status); err != nil {
		return true, fmt.Errorf("parse api: %w", err)
	} else if !ok {
		return true, nil
	}
	if err := r.pingPouet(&api); err != nil {
		return true, fmt.Errorf("parse api: %w", err)
	}
	r.FilePath = filepath.Join(storage, r.UUID)
	if skip := r.download(overwrite, &api, st); skip {
		return true, nil
	} else if update := r.check(); !update {
		return true, nil
	}
	if r.Platform == "" {
		r.platform(&api)
	}
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
		if err := r.fileMeta(); err != nil {
			return true, fmt.Errorf("parse api: %w", err)
		}
		r.save()
		fallthrough
	case r.FileZipContent == "":
		if zip, err := r.fileZipContent(); err != nil {
			return true, fmt.Errorf("parse api: %w", err)
		} else if zip {
			if err := r.doseeMeta(); err != nil {
				return true, fmt.Errorf("parse api: %w", err)
			}
		}
		r.save()
	}
	return false, nil
}

func (r *Record) pingPouet(api *ProductionsAPIv1) error {
	if id, code, err := api.PouetID(true); err != nil {
		return fmt.Errorf("ping pouet: %w", err)
	} else if id > 0 && code < 300 {
		r.WebIDPouet = uint(id)
	}
	return nil
}

func (r *Record) platform(api *ProductionsAPIv1) {
	const msdos, windows = 4, 1
	for _, p := range api.Platforms {
		switch p.ID {
		case msdos:
			r.Platform = dos
		case windows:
			r.Platform = win
		default:
			break
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

func (r *Record) variations() (names []string, err error) {
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

// newRecord initialises a new file record.
func newRecord(c int, values []sql.RawBytes) (r Record, err error) {
	const want = 21
	if l := len(values); l < want {
		return r, fmt.Errorf("new records %q: %w", l, err)
	}
	r = Record{
		count: c,
		ID:    string(values[0]), // id
		UUID:  string(values[1]), // uuid
		// deletedat
		CreatedAt: database.DateTime(values[3]), // createdat
		Filename:  string(values[4]),            // filename
		Filesize:  string(values[5]),            // filesize
		// web_id_demozoo
		FileZipContent: string(values[7]),            // file_zip_content
		UpdatedAt:      database.DateTime(values[8]), // updatedat
		Platform:       string(values[9]),            // platform
		Sum384:         string(values[10]),           // file_integrity_strong
		SumMD5:         string(values[11]),           // file_integrity_weak
		// web_id_pouet
		GroupFor:    string(values[13]),
		GroupBy:     string(values[14]),
		Title:       string(values[15]),
		Section:     string(values[16]),
		CreditArt:   strings.Split(string(values[17]), ","),
		CreditAudio: strings.Split(string(values[18]), ","),
		CreditCode:  strings.Split(string(values[19]), ","),
		CreditText:  strings.Split(string(values[20]), ","),
	}
	if i, err := strconv.Atoi(string(values[6])); err == nil {
		r.WebIDDemozoo = uint(i)
	}
	if i, err := strconv.Atoi(string(values[12])); err == nil {
		r.WebIDPouet = uint(i)
	}
	return r, nil
}

func selectByID() (stmt string) {
	const w = " FROM `files` WHERE `web_id_demozoo` IS NOT NULL"
	where := w
	if requestedID != "" {
		switch {
		case database.IsUUID(requestedID):
			where = fmt.Sprintf("%v AND `uuid`=%q", w, requestedID)
		case database.IsID(requestedID):
			where = fmt.Sprintf("%v AND `id`=%q", w, requestedID)
		}
	}
	return selectSQL + where
}
