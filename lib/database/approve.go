package database

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize" //nolint:misspell
	"github.com/gookit/color"       //nolint:misspell
	"github.com/spf13/viper"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

const (
	readWriteEveryone             = 0666
	fm                os.FileMode = readWriteEveryone
)

const newFilesSQL = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`,`web_id_demozoo`," +
	"`file_zip_content`,`updatedat`,`platform`,`file_integrity_strong`,`file_integrity_weak`,`web_id_pouet`" +
	",`group_brand_for`,`group_brand_by`,`section`\n" +
	"FROM `files`\n" +
	"WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"

const countWaiting = "SELECT COUNT(*)\nFROM `files`\n" +
	"WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"

// Approve automatically checks and clears file records for live.
func Approve(verbose bool) error {
	if err := queries(verbose); err != nil {
		return err
	}
	return nil
}

// Waiting returns the number of files requiring approval for public display.
func Waiting() (count uint, err error) {
	db := Connect()
	defer db.Close()
	if err = db.QueryRow(countWaiting).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// duplicate or copy a file to the destination.
func dupe(name, dest string) (written int64, err error) {
	src, err := os.Open(name)
	if err != nil {
		return 0, fmt.Errorf("dupe open %s: %w", name, err)
	}
	defer src.Close()
	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, fm)
	if err != nil {
		return 0, fmt.Errorf("dupe open new %s: %w", dest, err)
	}
	defer dst.Close()
	written, err = io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf("dupe io: %w", err)
	}
	return written, dst.Close()
}

func verbose(v bool, i interface{}) {
	if !v {
		return
	}
	const exclamationMark = 33
	if val, ok := i.(string); ok {
		if len(val) > 0 && val[0] == exclamationMark {
			logs.Printf("%s", color.Warn.Sprint(val))
			return
		}
		logs.Printf("%s", val)
	}
}

// queries parses all records waiting for approval skipping those that
// are missing expected data or assets such as thumbnails.
func queries(v bool) error {
	db := Connect()
	defer db.Close()
	rows, err := db.Query(newFilesSQL)
	if err != nil {
		return fmt.Errorf("queries query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("queries query rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("queries columns: %w", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	dir := directories.Init(false)
	// fetch the rows
	var r record
	r.c = 0
	r.verbose = v
	rowCnt := 0
	for rows.Next() {
		rowCnt++
		if err = rows.Scan(scanArgs...); err != nil {
			return fmt.Errorf("queries row scan: %w", err)
		}
		if n := NewApprove(values); !n {
			continue
		}
		r.uuid = string(values[1])
		verbose(v, fmt.Sprintf("\n%s item %04d (%v) %s %s ", str.X(),
			rowCnt, string(values[0]), color.Primary.Sprint(r.uuid), color.Info.Sprint(r.filename)))
		if ok := r.check(values, &dir); !ok {
			continue
		}
		r.save = true
		if r.autoID(string(values[0])) == 0 {
			r.save = false
		} else if err := r.approve(); err != nil {
			logs.Log(err)
			r.save = false
		}
		r.c++
	}
	verbose(v, "\n")
	r.summary(rowCnt)
	return nil
}

type record struct {
	c          int
	save       bool
	verbose    bool
	id         uint
	uuid       string
	filename   string
	filesize   uint
	zipContent string
	groupBy    string
	groupFor   string
	platform   string
	tag        string
	hashStrong string
	hashWeak   string
}

func (r *record) String() string {
	status := str.Y()
	if !r.save {
		status = str.X()
	}
	return fmt.Sprintf("%s item %04d (%v) %s %s", status, r.c, r.id,
		color.Primary.Sprint(r.uuid), color.Info.Sprint(r.filename))
}

// approve sets the record to be publically viewable.
func (r *record) approve() error {
	db := Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	if err != nil {
		return fmt.Errorf("record approve prepare: %w", err)
	}
	defer update.Close()
	if _, err := update.Exec(UpdateID, r.id); err != nil {
		return fmt.Errorf("record approve exec: %w", err)
	}
	return nil
}

func (r *record) autoID(data string) (id uint) {
	i, err := strconv.Atoi(data)
	if err != nil {
		return 0
	}
	id = uint(i)
	r.id = id
	return id
}

func (r *record) check(values []sql.RawBytes, dir *directories.Dir) (ok bool) {
	v := r.verbose
	if !r.checkFileName(string(values[4])) {
		verbose(v, "!filename")
		return false
	}
	if !r.checkFileSize(string(values[5])) {
		verbose(v, "!filesize")
		return false
	}
	if !r.checkHash(string(values[10]), string(values[11])) {
		verbose(v, "!hash")
		return false
	}
	if !r.checkFileContent(string(values[7])) {
		verbose(v, "!file content")
		return false
	}
	if !r.checkGroups(string(values[14]), string(values[13])) {
		verbose(v, "!group")
		return false
	}
	if !r.checkTags(string(values[9]), string(values[15])) {
		verbose(v, "!tag")
		return false
	}
	if !r.checkDownload(dir.UUID) {
		verbose(v, "!download")
		return false
	}
	if string(values[9]) != "audio" {
		if !r.checkImage(dir.Img000) {
			verbose(v, "!000x")
			return false
		}
		if !r.checkImage(dir.Img400) {
			verbose(v, "!400x")
			return false
		}
		if !r.checkImage(dir.Img150) {
			verbose(v, "!150x")
			return false
		}
	}
	return true
}

func (r *record) checkDownload(path string) (ok bool) {
	file := filepath.Join(fmt.Sprint(path), r.uuid)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return r.recoverDownload(path)
	}
	return true
}

func (r *record) checkFileContent(fc string) (ok bool) {
	r.zipContent = fc
	switch filepath.Ext(r.filename) {
	case z7, arj, rar, zip:
		if r.zipContent == "" {
			return false
		}
	}
	return true
}

func (r *record) checkFileName(fn string) (ok bool) {
	if r.filename = fn; r.filename == "" {
		return false
	}
	return true
}

func (r *record) checkFileSize(fs string) (ok bool) {
	i, err := strconv.Atoi(fs)
	if err != nil {
		return false
	}
	r.filesize = uint(i)
	return true
}

func (r *record) checkGroups(g1, g2 string) (ok bool) {
	r.groupBy, r.groupFor = g1, g2
	if r.groupBy == "" && r.groupFor == "" {
		return false
	}
	if strings.EqualFold(r.groupBy, changeme) || strings.EqualFold(r.groupFor, changeme) {
		return false
	}
	return true
}

func (r *record) checkHash(h1, h2 string) (ok bool) {
	if r.hashStrong = h1; r.hashStrong == "" {
		return false
	}
	if r.hashWeak = h2; r.hashWeak == "" {
		return false
	}
	return true
}

func (r *record) checkImage(path string) (ok bool) {
	if _, err := os.Stat(r.imagePath(path)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *record) checkTags(t1, t2 string) (ok bool) {
	if r.platform = t1; r.platform == "" {
		return false
	}
	if r.tag = t2; r.tag == "" {
		return false
	}
	return true
}

func (r *record) imagePath(path string) string {
	return filepath.Join(fmt.Sprint(path), r.uuid+png)
}

func (r *record) recoverDownload(path string) (ok bool) {
	src, v := viper.GetString("directory.incoming.files"), r.verbose
	if src == "" {
		return false
	}
	file := filepath.Join(src, r.filename)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		verbose(v, "!incoming:"+file+" ")
		return false
	}
	fc, err := dupe(file, path)
	if err != nil {
		verbose(v, "!filecopy ")
		logs.Log(err)
		return false
	}
	verbose(v, fmt.Sprintf("copied %v", humanize.Bytes(uint64(fc))))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		verbose(v, "!!filecopy ")
		logs.Log(err)
		return false
	}
	return true
}

func (r *record) summary(rows int) {
	t := fmt.Sprintf("Total items handled: %d", r.c)
	if rows <= r.c {
		logs.Println(strings.Repeat("─", len(t)))
		logs.Println(t)
		return
	}
	d := fmt.Sprintf("Database records fetched: %d", rows)
	logs.Println(strings.Repeat("─", len(d)))
	logs.Println(t)
	logs.Println(d)
}
