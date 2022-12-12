package recd

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database/internal/connect"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

const (
	// Datetime MySQL format.
	Datetime = "2006-01-02T15:04:05Z"
	// UpdateID is a user id to use with the updatedby column.
	UpdateID = "b66dc282-a029-4e99-85db-2cf2892fffcc"

	fm os.FileMode = 0o666

	filename       = 4
	filesize       = 5
	filezipcontent = 7
	platform       = 9
	hashstrong     = 10
	hashweak       = 11
	groupbrandfor  = 13
	groupbrandby   = 14
	section        = 15

	z7  = ".7z"
	arj = ".arj"
	png = ".png"
	rar = ".rar"
	zip = ".zip"

	changeme = "changeme"
)

type Record struct {
	C          int
	Save       bool
	Verbose    bool
	ID         uint
	UUID       string
	Filename   string
	filesize   uint
	zipContent string
	groupBy    string
	groupFor   string
	platform   string
	tag        string
	hashStrong string
	hashWeak   string
}

func (r *Record) String() string {
	status := str.Y()
	if !r.Save {
		status = str.X()
	}
	return fmt.Sprintf("%s item %04d (%v) %s %s", status, r.C, r.ID,
		color.Primary.Sprint(r.UUID), color.Info.Sprint(r.Filename))
}

// approve sets the record to be publically viewable.
func (r *Record) Approve() error {
	db := connect.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	if err != nil {
		return fmt.Errorf("record approve prepare: %w", err)
	}
	defer update.Close()
	if _, err := update.Exec(UpdateID, r.ID); err != nil {
		return fmt.Errorf("record approve exec: %w", err)
	}
	return nil
}

func (r *Record) AutoID(data string) (id uint) {
	i, err := strconv.Atoi(data)
	if err != nil {
		return 0
	}
	id = uint(i)
	r.ID = id
	return id
}

func (r *Record) Check(values []sql.RawBytes, dir *directories.Dir) (ok bool) {
	v := r.Verbose
	if !r.checkFileName(string(values[filename])) {
		verbose(v, "!filename")
		return false
	}
	if !r.checkFileSize(string(values[filesize])) {
		verbose(v, "!filesize")
		return false
	}
	if !r.checkHash(string(values[hashstrong]), string(values[hashweak])) {
		verbose(v, "!hash")
		return false
	}
	if !r.checkFileContent(string(values[filezipcontent])) {
		verbose(v, "!file content")
		return false
	}
	if !r.CheckGroups(string(values[groupbrandby]), string(values[groupbrandfor])) {
		verbose(v, "!group")
		return false
	}
	if !r.checkTags(string(values[platform]), string(values[section])) {
		verbose(v, "!tag")
		return false
	}
	if !r.checkDownload(dir.UUID) {
		verbose(v, "!download")
		return false
	}
	if string(values[platform]) != "audio" {
		if !r.checkImage(dir.Img000) {
			verbose(v, "!000x")
			return false
		}
		if !r.checkImage(dir.Img400) {
			verbose(v, "!400x")
			return false
		}
	}
	return true
}

func (r *Record) checkDownload(path string) (ok bool) {
	file := filepath.Join(fmt.Sprint(path), r.UUID)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return r.recoverDownload(path)
	}
	return true
}

func (r *Record) checkFileContent(fc string) (ok bool) {
	r.zipContent = fc
	switch filepath.Ext(r.Filename) {
	case z7, arj, rar, zip:
		if r.zipContent == "" {
			return false
		}
	}
	return true
}

func (r *Record) checkFileName(fn string) (ok bool) {
	if r.Filename = fn; r.Filename == "" {
		return false
	}
	return true
}

func (r *Record) checkFileSize(fs string) (ok bool) {
	i, err := strconv.Atoi(fs)
	if err != nil {
		return false
	}
	r.filesize = uint(i)
	return true
}

func (r *Record) CheckGroups(g1, g2 string) (ok bool) {
	r.groupBy, r.groupFor = g1, g2
	if r.groupBy == "" && r.groupFor == "" {
		return false
	}
	if strings.EqualFold(r.groupBy, changeme) || strings.EqualFold(r.groupFor, changeme) {
		return false
	}
	return true
}

func (r *Record) checkHash(h1, h2 string) (ok bool) {
	if r.hashStrong = h1; r.hashStrong == "" {
		return false
	}
	if r.hashWeak = h2; r.hashWeak == "" {
		return false
	}
	return true
}

func (r *Record) checkImage(path string) (ok bool) {
	if _, err := os.Stat(r.ImagePath(path)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *Record) checkTags(t1, t2 string) (ok bool) {
	if r.platform = t1; r.platform == "" {
		return false
	}
	if r.tag = t2; r.tag == "" {
		return false
	}
	return true
}

func (r *Record) ImagePath(path string) string {
	return filepath.Join(fmt.Sprint(path), r.UUID+png)
}

func (r *Record) recoverDownload(path string) (ok bool) {
	src, v := viper.GetString("directory.incoming.files"), r.Verbose
	if src == "" {
		return false
	}
	file := filepath.Join(src, r.Filename)
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

func (r *Record) Summary(rows int) {
	if rows == 0 {
		const m = "No files were approved"
		l := strings.Repeat("─", len(m))
		logs.Printf("%s\n%s\n%s\n", l, m, l)
		return
	}
	t := fmt.Sprintf("%d new files approved", r.C)
	if r.C == 1 {
		t = fmt.Sprintf("%d new file approved", r.C)
	}
	if rows <= r.C {
		l := strings.Repeat("─", len(t))
		logs.Printf("%s\n%s\n%s\n", l, t, l)
		return
	}
	d := fmt.Sprintf("%d database new records were skipped", rows-r.C)
	l := strings.Repeat("─", len(d))
	logs.Printf("%s\n%s\n%s\n%s\n", l, t, d, l)
}

func verbose(v bool, i any) {
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

// -----------------

// NewApprove reports if a new file record is set to unapproved.
func NewApprove(b []sql.RawBytes) bool {
	// SQL column names can be found in the newFilesSQL statement in approve.go
	const deletedat, createdat = 2, 3
	if b[deletedat] == nil {
		return false
	}
	n, err := Valid(b[deletedat], b[createdat])
	if err != nil {
		logs.Log(err)
	}
	return n
}

func Valid(deletedat, updatedat sql.RawBytes) (bool, error) {
	const (
		min = -5
		max = 5
	)
	// normalise the date values as sometimes updatedat & deletedat can be off by a second.
	del, err := time.Parse(time.RFC3339, string(deletedat))
	if err != nil {
		return false, fmt.Errorf("valid deleted time: %w", err)
	}
	upd, err := time.Parse(time.RFC3339, string(updatedat))
	if err != nil {
		return false, fmt.Errorf("valid updated time: %w", err)
	}
	if diff := upd.Sub(del); diff.Seconds() > max || diff.Seconds() < min {
		return false, nil
	}
	return true, nil
}

func ColLen(s *sql.ColumnType) string {
	l, ok := s.Length()
	if !ok {
		return ""
	}
	if l > 0 {
		return strconv.Itoa(int(l))
	}
	return ""
}

// ReverseInt swaps the direction of the value, 12345 would return 54321.
func ReverseInt(i uint) (uint, error) {
	var (
		n int
		s string
	)
	v := strconv.Itoa(int(i))
	for x := len(v); x > 0; x-- {
		s += string(v[x-1])
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return i, fmt.Errorf("reverse int %q: %w", s, err)
	}
	return uint(n), nil
}

func Verbose(v bool, i any) {
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

const newFilesSQL = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`," +
	"`web_id_demozoo`,`file_zip_content`,`updatedat`,`platform`,`file_integrity_strong`," +
	"`file_integrity_weak`,`web_id_pouet`,`group_brand_for`,`group_brand_by`,`section`\n" +
	"FROM `files`\n" +
	"WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"

// queries parses all records waiting for approval skipping those that
// are missing expected data or assets such as thumbnails.
func Queries(v bool) error {
	db := connect.Connect()
	defer db.Close()
	rows, err := db.Query(newFilesSQL)
	if err != nil {
		return fmt.Errorf("queries query: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("queries query rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("queries columns: %w", err)
	}
	return query(v, rows, columns)
}

func query(v bool, rows *sql.Rows, columns []string) error {
	x := func() string {
		return fmt.Sprintf(" %s", str.X())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	dir := directories.Init(false)
	// fetch the rows
	var r Record
	r.C = 0
	r.Verbose = v
	rowCnt := 0
	for rows.Next() {
		rowCnt++
		if err := rows.Scan(scanArgs...); err != nil {
			return fmt.Errorf("queries row scan: %w", err)
		}
		Verbose(v, fmt.Sprintf("\nitem %04d (%v) %s %s ",
			rowCnt, string(values[0]), color.Primary.Sprint(r.UUID), color.Info.Sprint(r.Filename)))
		if na, dz := NewApprove(values), IsDemozoo(values); !na && !dz {
			Verbose(v, x())
			continue
		}
		r.UUID = string(values[1])
		if ok := r.Check(values, &dir); !ok {
			Verbose(v, x())
			continue
		}
		r.Save = true
		if r.AutoID(string(values[0])) == 0 {
			r.Save = false
		} else if err := r.Approve(); err != nil {
			Verbose(v, x())
			logs.Log(err)
			r.Save = false
		}
		Verbose(v, fmt.Sprintf(" %s", str.Y()))
		r.C++
	}
	if rowCnt > 0 {
		Verbose(v, "\n")
	}
	r.Summary(rowCnt)
	return nil
}

// IsDemozoo reports if a fetched demozoo file record is set to unapproved.
func IsDemozoo(b []sql.RawBytes) bool {
	// SQL column names can be found in the selectSQL statement in database.go
	const deletedat, updatedat = 2, 8
	if b[deletedat] == nil {
		return false
	}
	n, err := Valid(b[deletedat], b[updatedat])
	if err != nil {
		logs.Log(err)
	}
	return n
}
