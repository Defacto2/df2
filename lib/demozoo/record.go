package demozoo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/gookit/color"
)

const sep = ","

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

// Save a file item record to the database.
func (r *Record) Save() error {
	db := database.Connect()
	defer db.Close()
	query, args := r.sql()
	update, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("record save db prepare: %w", err)
	}
	defer update.Close()
	_, err = update.Exec(args...)
	if err != nil {
		return fmt.Errorf("record save db exec: %w", err)
	}
	return nil
}

// fileZipContent reads an archive and saves its content to the database.
func (r *Record) fileZipContent() (ok bool, err error) {
	if r.FilePath == "" {
		return false, fmt.Errorf("record filezipcontent: %w", ErrFilePath)
	} else if r.Filename == "" {
		return false, fmt.Errorf("record filezipcontent: %w", ErrFilename)
	}
	a, err := archive.Read(r.FilePath, r.Filename)
	if err != nil {
		return false, fmt.Errorf("record filezipcontent archive read: %w", err)
	}
	r.FileZipContent = strings.Join(a, "\n")
	return true, nil
}

func (r *Record) sql() (query string, args []interface{}) {
	// an range map iternation is not used due to the varied comparisons
	set := r.sqlSets()
	args = r.sqlArgs(set)
	if len(set) == 0 {
		return "", args
	}
	set = append(set, "updatedat=?")
	args = append(args, []interface{}{time.Now()}...)
	set = append(set, "updatedby=?")
	args = append(args, []interface{}{database.UpdateID}...)
	query = "UPDATE files SET " + strings.Join(set, sep) + " WHERE id=?"
	args = append(args, []interface{}{r.ID}...)
	return query, args
}

func (r *Record) sqlSets() []string {
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
	set = append(set, r.sqlCredits()...)
	if r.Platform != "" {
		set = append(set, "platform=?")
	}
	return set
}

func (r *Record) sqlCredits() []string {
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

func (r *Record) sqlArgs(set []string) (args []interface{}) {
	if r.Filename != "" {
		args = append(args, []interface{}{r.Filename}...)
	}
	if r.Filesize != "" {
		args = append(args, []interface{}{r.Filesize}...)
	}
	if r.FileZipContent != "" {
		args = append(args, []interface{}{r.FileZipContent}...)
	}
	if r.SumMD5 != "" {
		args = append(args, []interface{}{r.SumMD5}...)
	}
	if r.Sum384 != "" {
		args = append(args, []interface{}{r.Sum384}...)
	}
	const errYear = 0o001
	if r.LastMod.Year() != errYear {
		args = append(args, []interface{}{r.LastMod}...)
	}
	if r.WebIDPouet != 0 {
		args = append(args, []interface{}{r.WebIDPouet}...)
	}
	if r.WebIDDemozoo == 0 && len(set) > 0 {
		args = append(args, []interface{}{""}...)
	}
	if r.DOSeeBinary != "" {
		args = append(args, []interface{}{r.DOSeeBinary}...)
	}
	if r.Readme != "" {
		args = append(args, []interface{}{r.Readme}...)
	}
	if r.Title != "" {
		args = append(args, []interface{}{r.Title}...)
	}
	args = append(args, r.sqlCreditsArgs()...)
	if r.Platform != "" {
		args = append(args, []interface{}{r.Platform}...)
	}
	return args
}

func (r *Record) sqlCreditsArgs() (args []interface{}) {
	if len(r.CreditText) > 0 {
		j := strings.Join(r.CreditText, sep)
		args = append(args, []interface{}{j}...)
	}
	if len(r.CreditCode) > 0 {
		j := strings.Join(r.CreditCode, sep)
		args = append(args, []interface{}{j}...)
	}
	if len(r.CreditArt) > 0 {
		j := strings.Join(r.CreditArt, sep)
		args = append(args, []interface{}{j}...)
	}
	if len(r.CreditAudio) > 0 {
		j := strings.Join(r.CreditAudio, sep)
		args = append(args, []interface{}{j}...)
	}
	return args
}

func (st *stat) fileExist(r *Record) (missing bool) {
	if s, err := os.Stat(r.FilePath); os.IsNotExist(err) || s.IsDir() {
		st.Missing++
		return true
	}
	return false
}
