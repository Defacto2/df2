package record

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/dizzer/zwt"
	"github.com/google/uuid"
)

var (
	ErrGroup   = errors.New("record group field cannot be empty")
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

type Record struct {
	UUID       string    // uuid *+
	Title      string    // record_title *+
	Group      string    // group_brand_for *+
	FileName   string    // filename *
	FileSize   int64     // filesize *
	FileMagic  string    // file_magic_type *
	HashStrong string    // file_integrity_strong *
	HashWeak   string    // file_integrity_weak *
	LastMod    time.Time // file_last_modified *
	Published  time.Time // date_issued_year,date_issued_month,date_issued_day *
	Section    string    // todo use a constant default *+
	Platform   string    // todo use a const default *+
	Comment    string    // key *
}

func New(name, group string) Record {
	uid, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}
	return Record{
		UUID:     uid.String(),
		Group:    group,
		Section:  Section,
		Platform: Platform,
		Comment:  fmt.Sprintf("release directory: %s", name),
	}
}

func (r *Record) Copy(d *Download, title string) {
	r.Title = d.ReadTitle // create a fallback
	if r.Title == "" {
		r.Title = title
	}
	r.FileName = d.Name
	r.FileSize = d.Bytes
	r.FileMagic = d.Magic
	r.HashStrong = d.HashStrong
	r.HashWeak = d.HashWeak
	r.LastMod = d.LastMod
	r.Published = d.ReadDate
}

type Download struct {
	Path       string
	Name       string
	Bytes      int64
	HashStrong string
	HashWeak   string
	Magic      string
	LastMod    time.Time
	ReadTitle  string
	ReadDate   time.Time
}

func (dl *Download) New(name, group string) error {
	st, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	// dl.LastMod must be set using the RAR archive metadata.
	dl.Path = name
	dl.Name = st.Name()
	dl.Bytes = st.Size()
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()

	if filepath.Base(name) == "file_id.diz" {
		diz := strings.Builder{}
		fileScanner := bufio.NewScanner(f)
		for fileScanner.Scan() {
			fmt.Fprintln(&diz, fileScanner.Text())
		}
		if err := dl.ReadDIZ(diz.String(), group); err != nil {
			return err
		}
	}

	strong, err := Sum386(f)
	if err != nil {
		return err
	}
	dl.HashStrong = strong

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

	y, m, d := 0, time.Month(0), 0
	title, pub := "", ""
	switch strings.ToLower(group) {
	case "":
		return ErrGroup
	case "zwt", zwt.Name:
		y, m, d = zwt.DizDate(body)
		title, pub = zwt.DizTitle(body)
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
