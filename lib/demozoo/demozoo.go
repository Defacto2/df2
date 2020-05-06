package demozoo

import (
	"crypto/md5"
	"crypto/sha512"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

const prodAPI string = "https://demozoo.org/api/v1/productions"

const selectSQL = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`,`web_id_demozoo`,`file_zip_content`," +
	"`updatedat`,`platform`,`file_integrity_strong`,`file_integrity_weak`,`web_id_pouet`,`group_brand_for`," +
	"`group_brand_by`,`record_title`,`section`,`credit_illustration`,`credit_audio`,`credit_program`,`credit_text`"

// Request proofs.
type Request struct {
	All       bool // parse all demozoo entries
	Overwrite bool // overwrite existing files
	Refresh   bool // refresh all demozoo entries
	Simulate  bool // simulate database save
}

// query statistics
type stat struct {
	count   int
	fetched int
	missing int
	total   int
}

var prodID = ""

// Fetch a Demozoo production by its ID.
func Fetch(id uint) (code int, status string, api ProductionsAPIv1) {
	var d = Production{
		ID: int64(id),
	}
	api = *d.data()
	return d.StatusCode, d.Status, api
}

// Query parses a single Demozoo entry.
func (req Request) Query(id string) (err error) {
	if err = database.CheckID(id); err != nil {
		return err
	}
	prodID = id
	return req.Queries()
}

// newRecord initialises a new file record.
func newRecord(c int, values []sql.RawBytes) (r Record) {
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
	return r
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (req Request) Queries() error {
	st := stat{count: 0, missing: 0, total: 0}
	var stmt = selectByID()
	start := time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(stmt)
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
	storage := directories.Init(false).UUID
	st.sumTotal(records{rows, scanArgs, values}, req)
	if st.total > 1 {
		logs.Println("Total records", st.total)
	}
	rows, err = db.Query(stmt)
	if err != nil {
		return err
	}
	for rows.Next() {
		st.fetched++
		if skip := st.nextResult(records{rows, scanArgs, values}, req); skip {
			continue
		}
		r := newRecord(st.count, values)
		logs.Printfcr(r.String(st.total))
		if skip := r.parseAPI(st, req.Overwrite, storage); skip {
			continue
		}
		switch {
		case st.total <= 1:
			break
		case req.Simulate:
			logs.Printf(" • dry-run %v", logs.Y())
		default:
			r.save()
		}
	}
	logs.Check(rows.Err())
	if prodID != "" {
		if st.count == 0 {
			var t string
			if st.fetched == 0 {
				t = fmt.Sprintf("id %q is not a Demozoo sourced file record", prodID)
			} else {
				t = fmt.Sprintf("id %q is not a new Demozoo record, use --id=%v --overwrite to refetch the download and data", prodID, prodID)
			}
			logs.Println(t)
		}
		return nil
	}
	st.summary(time.Since(start))
	return nil
}

func (r *Record) save() {
	if err := r.Save(); err != nil {
		logs.Printf(" %v \n", logs.X())
		logs.Log(err)
		return
	}
	logs.Printf(" • saved %v", logs.Y())
}

func (st stat) summary(elapsed time.Duration) {
	t := fmt.Sprintf("Total Demozoo items handled: %v, time elapsed %s", st.count, elapsed)
	logs.Println("\n" + strings.Repeat("─", len(t)))
	logs.Println(t)
	if st.missing > 0 {
		logs.Println("UUID files not found:", st.missing)
	}
}

func selectByID() (sql string) {
	const w = " FROM `files` WHERE `web_id_demozoo` IS NOT NULL"
	where := w
	if prodID != "" {
		switch {
		case database.IsUUID(prodID):
			where = fmt.Sprintf("%v AND `uuid`=%q", w, prodID)
		case database.IsID(prodID):
			where = fmt.Sprintf("%v AND `id`=%q", w, prodID)
		}
	}
	return selectSQL + where
}

// sumTotal calculates the total number of conditional rows.
func (st *stat) sumTotal(rec records, req Request) {
	for rec.rows.Next() {
		err := rec.rows.Scan(rec.scanArgs...)
		logs.Check(err)
		if new := database.IsNew(rec.values); !new && req.flags() {
			continue
		}
		st.total++
	}
}

// nextResult checks for the next, new record.
func (st *stat) nextResult(rec records, req Request) (skip bool) {
	err := rec.rows.Scan(rec.scanArgs...)
	logs.Check(err)
	if new := database.IsNew(rec.values); !new && req.flags() {
		return true
	}
	st.count++
	return false
}

func (req Request) flags() (skip bool) {
	if !req.All && !req.Refresh && !req.Overwrite {
		return true
	}
	return false
}

// parseAPI confirms and parses the API request.
func (r Record) parseAPI(st stat, overwrite bool, storage string) (skip bool) {
	code, status, api := Fetch(r.WebIDDemozoo)
	if ok := r.confirm(code, status); !ok {
		return true
	}
	r.pingPouet(api)
	r.FilePath = filepath.Join(storage, r.UUID)
	if skip := r.download(overwrite, api, st); skip {
		return true
	}
	if update := r.check(); !update {
		return true
	}
	if r.Platform == "" {
		r.platform(api)
	}
	switch {
	case
		r.Filesize == "",
		r.SumMD5 == "",
		r.Sum384 == "":
		if err := r.fileMeta(); err != nil {
			logs.Log(err)
			return true
		}
		fallthrough
	case r.FileZipContent == "":
		if zip := r.fileZipContent(); zip {
			r.doseeMeta()
		}
	}
	return false
}

func (r *Record) pingPouet(api ProductionsAPIv1) {
	if id, code := api.PouetID(true); id > 0 && code < 300 {
		r.WebIDPouet = uint(id)
	}
}

func (r *Record) download(overwrite bool, api ProductionsAPIv1, st stat) (skip bool) {
	if st.fileExist(*r) || overwrite {
		name, link := api.DownloadLink()
		if len(link) == 0 {
			logs.Print(color.Note.Sprint("no suitable downloads found\n"))
			return true
		}
		logs.Printfcr("%s%s %s", r.String(st.total), color.Primary.Sprint(link), download.StatusColor(200, "200 OK"))
		head, err := download.LinkDownload(r.FilePath, link)
		if err != nil {
			logs.Log(err)
			return true
		}
		logs.Printfcr(r.String(st.total))
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

// last modified time passed via HTTP
func (r *Record) lastMod(head http.Header) {
	if lm := head.Get("Last-Modified"); len(lm) > 0 {
		if t, err := time.Parse(download.RFC5322, lm); err != nil {
			logs.Printf(" • last-mod value %q ?", lm)
		} else {
			r.LastMod = t
			if time.Now().Year() == t.Year() {
				logs.Printf(" • %s", t.Format("2 Jan"))
			} else {
				logs.Printf(" • %s", t.Format("Jan 06"))
			}
		}
	}
}

// check record to see if it needs updating.
func (r Record) check() (update bool) {
	switch {
	case
		r.Platform == "",
		r.Filesize == "",
		r.Sum384 == "",
		r.SumMD5 == "",
		r.FileZipContent == "":
		return true
	default:
		logs.Printf("skipped, no changes needed %v", logs.Y())
		return false
	}
}

func (r *Record) platform(api ProductionsAPIv1) {
	for _, p := range api.Platforms {
		switch p.ID {
		case 4:
			r.Platform = "dos"
		case 1:
			r.Platform = "windows"
		}
		break
	}
}

func (r *Record) fileMeta() (err error) {
	// file size
	stat, err := os.Stat(r.FilePath)
	if err != nil {
		return err
		//continue
	}
	r.Filesize = strconv.Itoa(int(stat.Size()))
	// file hashes
	f, err := os.Open(r.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h1 := md5.New()
	if _, err := io.Copy(h1, f); err != nil {
		return err
	}
	r.SumMD5 = fmt.Sprintf("%x", h1.Sum(nil))
	h2 := sha512.New384()
	if _, err := io.Copy(h2, f); err != nil {
		return err
	}
	r.Sum384 = fmt.Sprintf("%x", h2.Sum(nil))
	return nil
}

func (r *Record) doseeMeta() {
	var names = r.variations()
	d, err := archive.ExtractDemozoo(r.FilePath, r.UUID, &names)
	logs.Log(err)
	if strings.ToLower(r.Platform) == "dos" && d.DOSee != "" {
		r.DOSeeBinary = d.DOSee
	}
	if d.NFO != "" {
		r.Readme = d.NFO
	}
	if strings.ToLower(r.Platform) == "dos" && d.DOSee != "" {
		r.DOSeeBinary = d.DOSee
	}
}

func (r Record) variations() (names []string) {
	if r.GroupBy != "" {
		names = append(names, groups.Variations(r.GroupBy)...)
	}
	if r.GroupFor != "" {
		names = append(names, groups.Variations(r.GroupFor)...)
	}
	return names
}
