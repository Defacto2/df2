package database

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

const newFilesSQL = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`,`web_id_demozoo`," +
	"`file_zip_content`,`updatedat`,`platform`,`file_integrity_strong`,`file_integrity_weak`,`web_id_pouet`" +
	",`group_brand_for`,`group_brand_by`,`section`\n" +
	"FROM `files`\n" +
	"WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"

var verb = false

// Approve automatically checks and clears file records for live.
func Approve(verbose bool) {
	verb = verbose
	err := queries()
	logs.Check(err)
}

type record struct {
	c          int
	save       bool
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

func (r record) String() string {
	status := logs.Y()
	if !r.save {
		status = logs.X()
	}
	return fmt.Sprintf("%s item %04d (%v) %s %s", status, r.c, r.id, color.Primary.Sprint(r.uuid), color.Info.Sprint(r.filename))
}

// queries parses all records waiting for approval skipping those that are missing expected data or assets such as thumbnails.
func queries() error {
	db := Connect()
	defer db.Close()
	rows, err := db.Query(newFilesSQL)
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
	var r record
	r.c = 0
	rowCnt := 0
	for rows.Next() {
		rowCnt++
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := IsNew(values); !new {
			continue
		}
		r.uuid = string(values[1])
		printV(fmt.Sprintf("\n%s item %04d (%v) %s %s ", logs.X(), rowCnt, string(values[0]), color.Primary.Sprint(r.uuid), color.Info.Sprint(r.filename)))
		if !r.checkFileName(string(values[4])) {
			printV("!filename")
			continue
		}
		if !r.checkFileSize(string(values[5])) {
			printV("!filesize")
			continue
		}
		if !r.checkHash(string(values[10]), string(values[11])) {
			printV("!hash")
			continue
		}
		if !r.checkFileContent(string(values[7])) {
			printV("!file content")
			continue
		}
		if !r.checkGroups(string(values[14]), string(values[13])) {
			printV("!group")
			continue
		}
		if !r.checkTags(string(values[9]), string(values[15])) {
			printV("!tag")
			continue
		}
		if !r.checkDownload(dir.UUID) {
			printV("!download")
			continue
		}
		if string(values[9]) != "audio" {
			if !r.checkImage(dir.Img000) {
				printV("!000x")
				continue
			}
			if !r.checkImage(dir.Img400) {
				printV("!400x")
				continue
			}
			if !r.checkImage(dir.Img150) {
				printV("!150x")
				continue
			}
		}
		r.save = true
		if r.autoID(string(values[0])) == 0 {
			r.save = false
		} else if err := r.approve(); err != nil {
			logs.Log(err)
			r.save = false
		}
		r.c++
		fmt.Println(r)
	}
	logs.Check(rows.Err())
	printV("\n")
	r.summary(rowCnt)
	return nil
}

func printV(i interface{}) {
	if !verb {
		return
	}
	switch v := i.(type) {
	case string:
		if len(v) > 0 && v[0] == 33 {
			logs.Printf("%s", color.Warn.Sprint(v))
		} else {
			logs.Printf("%s", v)
		}
	}
}

// approve sets the record to be publically viewable.
func (r record) approve() error {
	db := Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	logs.Check(err)
	_, err = update.Exec(UpdateID, r.id)
	logs.Check(err)
	return nil
}

func (r *record) autoID(data string) uint {
	i, err := strconv.Atoi(data)
	if err != nil {
		return 0
	}
	r.id = uint(i)
	return uint(i)
}

func (r record) checkDownload(path string) bool {
	file := filepath.Join(fmt.Sprint(path), r.uuid)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return r.recoverDownload(path)
	}
	return true
}

func (r record) recoverDownload(path string) bool {
	src := viper.GetString("directory.incoming.files")
	if src == "" {
		return false
	}
	file := filepath.Join(src, r.filename)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		printV("!incoming:" + file + " ")
		return false
	}
	fc, err := fileCopy(file, path)
	if err != nil {
		printV("!filecopy ")
		logs.Log(err)
		return false
	}
	printV(fmt.Sprintf("copied %v", humanize.Bytes(uint64(fc))))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		printV("!!filecopy ")
		logs.Log(err)
		return false
	}
	return true
}

func (r *record) checkFileContent(fc string) bool {
	r.zipContent = fc
	switch filepath.Ext(r.filename) {
	case ".7z", ".arj", ".rar", ".zip":
		if r.zipContent == "" {
			return false
		}
	}
	return true
}

func (r *record) checkFileName(fn string) bool {
	if r.filename = string(fn); r.filename == "" {
		return false
	}
	return true
}

func (r *record) checkFileSize(fs string) bool {
	i, err := strconv.Atoi(fs)
	if err != nil {
		return false
	}
	r.filesize = uint(i)
	return true
}

func (r *record) checkGroups(g1, g2 string) bool {
	r.groupBy = g1
	r.groupFor = g2
	if r.groupBy == "" && r.groupFor == "" {
		return false
	}
	if strings.ToLower(r.groupBy) == "changeme" || strings.ToLower(r.groupFor) == "changeme" {
		return false
	}
	return true
}

func (r *record) checkHash(h1, h2 string) bool {
	if r.hashStrong = string(h1); r.hashStrong == "" {
		return false
	}
	if r.hashWeak = string(h2); r.hashWeak == "" {
		return false
	}
	return true
}

func (r record) checkImage(path string) bool {
	if _, err := os.Stat(r.imagePath(path)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *record) checkTags(t1, t2 string) bool {
	if r.platform = t1; r.platform == "" {
		return false
	}
	if r.tag = t2; r.tag == "" {
		return false
	}
	return true
}

func (r record) imagePath(path string) string {
	return filepath.Join(fmt.Sprint(path), r.uuid+".png")
}

func (r record) summary(rows int) {
	t := fmt.Sprintf("Total items handled: %d", r.c)
	d := fmt.Sprintf("Database records fetched: %d", rows)
	if rows <= r.c {
		logs.Println(strings.Repeat("─", len(t)))
		logs.Println(t)
	} else {
		logs.Println(strings.Repeat("─", len(d)))
		logs.Println(t)
		logs.Println(d)
	}
}

// fileCopy copies a file to the destination.
func fileCopy(name, dest string) (int64, error) {
	src, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer src.Close()
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	i, err := io.Copy(file, src)
	if err != nil {
		return 0, err
	}
	return i, nil
}
