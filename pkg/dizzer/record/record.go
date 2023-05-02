// Package record creates an object that can be written as JSON or used as a new
// database record.
package record

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/dizzer/zwt"
)

var (
	ErrDir     = errors.New("the named file points to a directory")
	ErrGroup   = errors.New("record group field cannot be empty")
	ErrNew     = errors.New("new record name and group cannot be empty")
	ErrPointer = errors.New("pointer value cannot be nil")
)

const (
	FileID   = "file_id.diz"        // FileID is the filename of the release mini-descriptor.
	Section  = "releaseinformation" // Section releaseinformation is the default category tag for NFOs and file_id.diz.
	Platform = "text"               // Platform text is the default file format or NFOs and file_id.diz.
)

// todo: preview, readme, thumb400x
// new cli arg to --limit number of records to process
// need to lookup strong checksum before commit record add, file generation

// Record contains the fields that will be used as database cell values.
type Record struct {
	UUID       string    `json:"uuid"`
	Slug       string    `json:"slug"`
	Title      string    `json:"record_title"`
	Group      string    `json:"group_brand_for"`
	FileName   string    `json:"filename"`
	FileSize   int64     `json:"filesize"`
	FileMagic  string    `json:"file_magic_type"`
	HashStrong string    `json:"file_integrity_strong"`
	HashWeak   string    `json:"file_integrity_weak"`
	LastMod    time.Time `json:"file_last_modified"` // X
	Published  time.Time `json:"date_issued"`
	Section    string    `json:"section"`
	Platform   string    `json:"platform"`
	Comment    string    `json:"comment"`
	//ZipContent string    `json:"file_zip_content"`
	//Package    bool      `json:"package"` // X
}

// New creates a Record.
// The required name must be the subdirectory of the release.
// The required group is the formal release group name.
func New(uid, name, group string) (Record, error) {
	if uid == "" || name == "" || group == "" {
		return Record{}, ErrNew
	}
	// todo: validate uuid
	return Record{
		UUID:     uid,
		Slug:     name,
		Group:    group,
		Section:  Section,
		Platform: Platform,
		Comment:  fmt.Sprintf("release directory: %s", name),
	}, nil
}

// Copy the Download values to a new Record.
// The optional pathTitle should be the result of the PathTitle func.
func (r *Record) Copy(d *Download, pathTitle string) error {
	// if d == nil {
	// 	return fmt.Errorf("d %w", ErrPointer)
	// }
	// r.Title = d.ReadTitle // create a fallback
	// if r.Title == "" {
	// 	r.Title = pathTitle
	// }
	// r.FileName = d.Name
	// r.FileSize = d.Bytes
	// r.FileMagic = d.Magic
	// r.HashStrong = d.HashStrong
	// r.HashWeak = d.HashWeak
	// r.LastMod = d.LastMod
	// r.Published = d.ReadDate
	return nil
}

type Download struct {
	Path       string
	Name       string
	Bytes      int64
	HashStrong string
	HashWeak   string
	Magic      string
	ReadTitle  string
	ReadDate   time.Time
}

// New creates a Download from the named file.
// The required group is the formal release group name.
// TODO lastmod arg?
func (dl *Download) New(name, group string) error {
	if group == "" {
		return ErrGroup
	}
	st, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %s", ErrDir, name)
	}
	// the dl.LastMod value should not be set here,
	// it MUST be set using the RAR archive metadata.
	dl.Path = name
	dl.Name = st.Name()
	dl.Bytes = st.Size()
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()

	// hashes require the named file to be reopened after being read.
	f, err = os.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()
	strong, err := Sum386(f)
	if err != nil {
		return err
	}
	dl.HashStrong = strong
	// hashes require the named file to be reopened after being read.
	f, err = os.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()
	weak, err := SumMD5(f)
	if err != nil {
		return err
	}
	dl.HashWeak = weak

	magic, err := Determine(name)
	if err != nil {
		return err
	}
	dl.Magic = magic
	return nil
}

func (dl *Download) ReadDIZ(body string, group string) error {
	var (
		m          time.Month
		y, d       int
		pub, title string
	)
	switch strings.ToLower(group) {
	case "":
		return ErrGroup
	case "zwt", strings.ToLower(zwt.Name):
		y, m, d = zwt.DizDate(body)
		title, pub = zwt.DizTitle(body)
		fmt.Println("==================")
		fmt.Println("==================")
	default:
		// todo: generic dizdate, title etc?
		return nil
	}

	if y > 0 {
		dl.ReadDate = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}

	dl.ReadTitle = title
	if pub != "" && !strings.Contains(title, pub) {
		dl.ReadTitle = fmt.Sprintf("%s by %s", title, pub)
	}
	return nil
}
