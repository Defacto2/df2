package demozoo

import (
	"crypto/md5"
	"crypto/sha512"
	"database/sql"
	"fmt"
	"io"
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

// Fetch a Demozoo production by its ID.
func Fetch(id uint) (int, string, ProductionsAPIv1) {
	var d = Production{
		ID: int64(id),
	}
	api := *d.data()
	return d.StatusCode, d.Status, api
}

// Request proofs.
type Request struct {
	All       bool // parse all demozoo entries
	Overwrite bool // overwrite existing files
	Refresh   bool // refresh all demozoo entries
	Simulate  bool // simulate database save
}

var prodID = ""

// Query parses a single Demozoo entry.
func (req Request) Query(id string) error {
	if err := database.CheckID(id); err != nil {
		return err
	}
	prodID = id
	return req.Queries()
}

func sqlSelect() string {
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`,`web_id_demozoo`,`file_zip_content`,"
	s += "`updatedat`,`platform`,`file_integrity_strong`,`file_integrity_weak`,`web_id_pouet`,`group_brand_for`,"
	s += "`group_brand_by`,`record_title`,`section`,`credit_illustration`,`credit_audio`,`credit_program`,`credit_text`"
	w := " WHERE `web_id_demozoo` IS NOT NULL"
	if prodID != "" {
		switch {
		case database.IsUUID(prodID):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, prodID)
		case database.IsID(prodID):
			w = fmt.Sprintf("%v AND `id`=%q", w, prodID)
		}
	}
	return s + " FROM `files`" + w
}

type stat struct {
	count   int
	missing int
}

func newRecord(c int, values []sql.RawBytes) Record {
	r := Record{
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
	if i, err := strconv.Atoi(string(values[6])); err == nil { // deletedat
		r.WebIDDemozoo = uint(i)
	}
	if i, err := strconv.Atoi(string(values[12])); err == nil { // web_id_pouet
		r.WebIDPouet = uint(i)
	}
	return r
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (req Request) Queries() error {
	start := time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(sqlSelect())
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
	dir := directories.Init(false)
	// count the total number of rows
	total := 0
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := database.IsNew(values); !new && !req.All && !req.Refresh {
			continue
		}
		total++
	}
	if total > 1 {
		logs.Println("Total records", total)
	}
	// fetch the rows
	rows, err = db.Query(sqlSelect())
	if err != nil {
		return err
	}
	st := stat{count: 0, missing: 0}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := database.IsNew(values); !new && !req.All && !req.Refresh {
			continue
		}
		st.count++
		r := newRecord(st.count, values)
		logs.Printfcr(r.String(total)) // counter and record intro
		// confirm API request
		code, status, api := Fetch(r.WebIDDemozoo)
		if code == 404 {
			r.WebIDDemozoo = 0
			if err = r.Save(); err == nil {
				logs.Printf("(%s)\n", download.StatusColor(code, status))
			}
			continue
		}
		if code < 200 || code > 299 {
			logs.Printf("(%s)\n", download.StatusColor(code, status))
			continue
		}
		// pouet id
		if id, c := api.PouetID(true); id > 0 && c < 300 {
			r.WebIDPouet = uint(id)
		}
		// confirm & handle missing UUID host download file
		r.AbsFile = filepath.Join(dir.UUID, r.UUID)
		if st.absNotExist(r) || req.Overwrite {
			name, link := api.DownloadLink()
			if len(link) == 0 {
				logs.Print(color.Note.Sprint("no suitable downloads found\n"))
				continue
			}
			logs.Printfcr("%s%s %s", r.String(total), color.Primary.Sprint(link), download.StatusColor(200, "200 OK"))
			head, err := download.LinkDownload(r.AbsFile, link)
			if err != nil {
				logs.Log(err)
				continue
			}
			// reset existing data due to the new download
			logs.Printfcr(r.String(total))
			r.Filename = name
			logs.Printf("• %s", r.Filename)
			r.Filesize = ""
			r.SumMD5 = ""
			r.Sum384 = ""
			r.FileZipContent = ""
			// last modified time passed via HTTP
			if lm := head.Get("Last-Modified"); len(lm) > 0 {
				if t, err := time.Parse(download.RFC5322, lm); err != nil {
					logs.Log(err)
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
		switch {
		case r.Platform == "", r.Filesize == "", r.Sum384 == "", r.SumMD5 == "", r.FileZipContent == "":
		default: // skip record as nothing needs updating
			logs.Printf("skipped, no changes needed %v", logs.Y())
			continue
		}
		switch {
		case r.Platform == "":
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
		switch {
		case r.Filesize == "", r.SumMD5 == "", r.Sum384 == "":
			stat, err := os.Stat(r.AbsFile)
			if err != nil {
				logs.Log(err)
				continue
			}
			r.Filesize = strconv.Itoa(int(stat.Size()))
			// file hashes
			f, err := os.Open(r.AbsFile)
			if err != nil {
				logs.Log(err)
				continue
			}
			defer f.Close()
			h1 := md5.New()
			if _, err := io.Copy(h1, f); err != nil {
				logs.Log(err)
				continue
			}
			r.SumMD5 = fmt.Sprintf("%x", h1.Sum(nil))
			h2 := sha512.New384()
			if _, err := io.Copy(h2, f); err != nil {
				logs.Log(err)
				continue
			}
			r.Sum384 = fmt.Sprintf("%x", h2.Sum(nil))
			fallthrough
		case r.FileZipContent == "":
			if zip := r.fileZipContent(); zip {
				var names []string
				if r.GroupBy != "" {
					names = append(names, groups.Variations(r.GroupBy)...)
				}
				if r.GroupFor != "" {
					names = append(names, groups.Variations(r.GroupFor)...)
				}
				edz, err := archive.ExtractDemozoo(r.AbsFile, r.UUID, &names)
				logs.Log(err)
				if strings.ToLower(r.Platform) == "dos" && edz.DOSee != "" {
					r.DOSeeBinary = edz.DOSee
				}
				if edz.NFO != "" {
					r.Readme = edz.NFO
				}
			}
		}
		// save results
		if total < 2 {
			break
		}
		switch req.Simulate {
		case true:
			logs.Printf(" • dry-run %v", logs.Y())
		default:
			if err = r.Save(); err != nil {
				logs.Printf(" %v \n", logs.X())
				logs.Log(err)
			} else {
				logs.Printf(" • saved %v", logs.Y())
			}
		}
	}
	logs.Check(rows.Err())
	if prodID != "" {
		if st.count == 0 {
			t := fmt.Sprintf("id %q is not a Demozoo sourced file record", prodID)
			logs.Println(t)
		}
		return nil
	}
	st.summary(time.Since(start))
	return nil
}

func (st stat) summary(elapsed time.Duration) {
	t := fmt.Sprintf("Total Demozoo items handled: %v, time elapsed %s", st.count, elapsed)
	logs.Println("\n" + strings.Repeat("─", len(t)))
	logs.Println(t)
	if st.missing > 0 {
		logs.Println("UUID files not found:", st.missing)
	}
}
