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
	"github.com/dustin/go-humanize"
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
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`,`web_id_demozoo`,`file_zip_content`,`updatedat`,`platform`,`file_integrity_strong`,`file_integrity_weak`,`web_id_pouet`,`group_brand_for`,`group_brand_by`"
	w := " WHERE `web_id_demozoo` IS NOT NULL AND `platform` = 'dos'"
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

type row struct {
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
		GroupFor: string(values[13]),
		GroupBy:  string(values[14]),
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
	// fetch the rows
	rw := row{count: 0, missing: 0}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := database.IsNew(values); !new && !req.All {
			continue
		}
		rw.count++
		r := newRecord(rw.count, values)
		logs.Print(r.String()) // counter and record intro
		// confirm API request
		code, status, api := Fetch(r.WebIDDemozoo)
		if code < 200 || code > 299 {
			logs.Printf("(%s)\n", download.StatusColor(code, status))
			continue
		}
		// pouet id
		if id, c := api.PouetID(); id > 0 && c < 300 {
			r.WebIDPouet = uint(id)
		}
		// confirm & handle missing UUID host download file
		r.AbsFile = filepath.Join(dir.UUID, r.UUID)
		if rw.absNotExist(r) || req.Overwrite {
			name, link := api.DownloadLink()
			if len(link) == 0 {
				logs.Print(color.Note.Sprint("no suitable downloads found\n"))
				continue
			}
			logs.Printf("%s %s\n", color.Primary.Sprint(link), download.StatusColor(200, "200 OK"))
			head, err := download.LinkDownload(r.AbsFile, link)
			if err != nil {
				logs.Log(err)
				continue
			}
			logs.EL()
			logs.Print("\r")
			// reset existing data due to the new download
			r.Filename = name
			logs.Printf(" • %s", r.Filename)
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
						logs.Printf(" • mod %s", t.Format("2 Jan"))
					} else {
						logs.Printf(" • mod %s", t.Format("Jan 06"))
					}
				}
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
			logs.Printf(" • %s", humanize.Bytes(uint64(stat.Size())))
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
			logs.Printf(" • md5 %s", logs.Truncate(r.SumMD5, 10))
			h2 := sha512.New384()
			if _, err := io.Copy(h2, f); err != nil {
				logs.Log(err)
				continue
			}
			r.Sum384 = fmt.Sprintf("%x", h2.Sum(nil))
			logs.Printf(" • sha %s", logs.Truncate(r.Sum384, 10))
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
				c := strings.Split(r.FileZipContent, "\n")
				logs.Printf(" • %d files", len(c))
				if strings.ToLower(r.Platform) == "dos" && edz.DOSee != "" {
					logs.Printf(" • dosee %q", edz.DOSee)
					r.DOSeeBinary = edz.DOSee
				}
				if edz.NFO != "" {
					logs.Printf(" • text %q", edz.NFO)
					r.Readme = edz.NFO
				}
			}
		default: // skip record as nothing needs updating
			logs.Printf("skipped, no changes needed %v\n", logs.Y())
			continue
		}
		// save results
		switch req.Simulate {
		case true:
			logs.Printf(" • simulated %v\n", logs.Y())
		default:
			if err = r.Save(); err != nil {
				logs.Printf(" • saved %v ", logs.X())
				logs.Log(err)
			} else {
				logs.Printf(" • saved %v\n", logs.Y())
			}
		}
	}
	logs.Check(rows.Err())
	if prodID != "" {
		if rw.count == 0 {
			t := fmt.Sprintf("id %q is not a Demozoo sourced file record", prodID)
			logs.Println(t)
		}
		return nil
	}
	rw.summary()
	return nil
}

func (rw row) summary() {
	t := fmt.Sprintf("Total Demozoo items handled: %v", rw.count)
	logs.Println(strings.Repeat("─", len(t)))
	logs.Println(t)
	if rw.missing > 0 {
		logs.Println("UUID files not found:", rw.missing)
	}
}
