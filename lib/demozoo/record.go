package demozoo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

// Record of a file item.
type Record struct {
	count          int
	AbsFile        string // absolute path to file
	ID             string // mysql auto increment id
	UUID           string // record unique id
	WebIDDemozoo   uint   // demozoo production id
	WebIDPouet     uint
	Filename       string
	Filesize       string
	FileZipContent string
	CreatedAt      string
	UpdatedAt      string
	SumMD5         string    // file download MD5 hash
	Sum384         string    // file download SHA384 hash
	LastMod        time.Time // file download last modified time
	Readme         string
	DOSeeBinary    string
	Platform       string
	GroupFor       string
	GroupBy        string
}

func (r Record) String() string {
	return fmt.Sprintf("%s item %04d (%v) %v DZ:%v %v",
		logs.Y(), r.count, r.ID,
		color.Primary.Sprint(r.UUID), color.Note.Sprint(r.WebIDDemozoo),
		r.CreatedAt)
}

// Save a file item record to the database.
func (r Record) Save() error {
	db := database.Connect()
	defer db.Close()
	query, args := r.sql()
	update, err := db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = update.Exec(args...)
	if err != nil {
		return err
	}
	return nil
}

func (rw row) absNotExist(r Record) bool {
	if s, err := os.Stat(r.AbsFile); os.IsNotExist(err) || s.IsDir() {
		rw.missing++
		return true
	}
	return false
}

// fileZipContent reads an archive and saves its content to the database
func (r *Record) fileZipContent() bool {
	if r.AbsFile == "" {
		logs.Log(fmt.Errorf("r.absfile required by fileZipContent is empty"))
		return false
	}
	a, err := archive.Read(r.AbsFile)
	if err != nil {
		if err.Error() == "unarr: File not found" {
			logs.Log(fmt.Errorf("file not found: %s", r.AbsFile))
			return false
		}
		logs.Log(err)
		return false
	}
	r.FileZipContent = strings.Join(a, "\n")
	//updateZipContent(r.ID, strings.Join(a, "\n"))
	return true
}

func (r Record) sql() (string, []interface{}) {
	var set []string
	var args []interface{}

	if r.Filename != "" {
		set = append(set, "filename=?")
		args = append(args, []interface{}{r.Filename}...)
	}
	if r.Filesize != "" {
		set = append(set, "filesize=?")
		args = append(args, []interface{}{r.Filesize}...)
	}
	if r.FileZipContent != "" {
		set = append(set, "file_zip_content=?")
		args = append(args, []interface{}{r.FileZipContent}...)
	}
	if r.SumMD5 != "" {
		set = append(set, "file_integrity_weak=?")
		args = append(args, []interface{}{r.SumMD5}...)
	}
	if r.Sum384 != "" {
		set = append(set, "file_integrity_strong=?")
		args = append(args, []interface{}{r.Sum384}...)
	}
	if r.LastMod.Year() != 0001 {
		set = append(set, "file_last_modified=?")
		args = append(args, []interface{}{r.LastMod}...)
	}
	if r.WebIDPouet != 0 {
		set = append(set, "web_id_pouet=?")
		args = append(args, []interface{}{r.WebIDPouet}...)
	}
	if r.DOSeeBinary != "" {
		set = append(set, "dosee_run_program=?")
		args = append(args, []interface{}{r.DOSeeBinary}...)
	}
	if r.Readme != "" {
		set = append(set, "retrotxt_readme=?")
		args = append(args, []interface{}{r.Readme}...)
	}
	if len(set) == 0 {
		return "", args
	}
	set = append(set, "updatedat=?")
	args = append(args, []interface{}{time.Now()}...)
	set = append(set, "updatedby=?")
	args = append(args, []interface{}{database.UpdateID}...)
	query := "UPDATE files SET " + strings.Join(set, ",") + " WHERE id=?"
	args = append(args, []interface{}{r.ID}...)
	return query, args
}
