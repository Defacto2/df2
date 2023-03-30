package record

import (
	"errors"
	"fmt"
	"log"
	"os"
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
	UUID       string    // uuid *
	Title      string    // record_title *
	Group      string    // group_brand_for *
	FileName   string    // filename *
	FileSize   int64     // filesize *
	FileMagic  string    // file_magic_type *
	HashStrong string    // file_integrity_strong *
	HashWeak   string    // file_integrity_weak *
	LastMod    time.Time // file_last_modified *
	Published  time.Time // date_issued_year,date_issued_month,date_issued_day *
	Section    string    // todo use a constant default *
	Platform   string    // todo use a const default *
	Comment    string    // key *
}

func New(name string) Record {
	uid, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}
	return Record{
		UUID:     uid.String(),
		Section:  Section,
		Platform: Platform,
		Comment:  fmt.Sprintf("release directory: %s", name),
	}
}

func (r *Record) ReadDIZ(f *os.File) error {
	if f == nil {
		return fmt.Errorf("f %w", ErrPointer)
	}

	y, m, d := 0, time.Month(0), 0
	s, p := "", ""
	switch strings.ToLower(r.Group) {
	case "zwt":
		r.Group = zwt.Name
		y, m, d = zwt.DizDate(f)
		s, p = zwt.DizTitle(f)
	case "":
		return ErrGroup
	default:
		// todo: generic dizdate, title etc?
		return nil
	}

	r.Published = r.LastMod
	if y > 0 {
		r.Published = time.Date(y, m, d, 0, 0, 0, 0, nil)
	}

	r.Title = s + "??"
	if p != "" {
		r.Title = fmt.Sprintf("%s by %s", s, p)
	}
	return nil
}

func (r *Record) Read(name string) error {
	diz := strings.ToLower(name) == "file_id.diz"
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()
	if diz {
		if err := r.ReadDIZ(f); err != nil {
			return err
		}
	}

	strong, err := Sum386(f)
	if err != nil {
		return err
	}
	r.HashStrong = strong

	weak, err := SumMD5(f)
	if err != nil {
		return err
	}
	r.HashWeak = weak

	magic, err := Determine(name)
	if err != nil {
		return err
	}
	r.FileMagic = magic
	return nil
}

func (r *Record) Stat(name string) error {
	st, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	r.FileName = st.Name()
	r.FileSize = st.Size()
	// note: do no use lastmod time value
	return nil
}

type Download struct {
	Path       string
	Name       string
	Bytes      int64
	HashStrong string
	HashWeak   string
	Magic      string
	LastMod    time.Time
}

func (dl *Download) New(name string) error {
	st, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	dl.Path = name
	dl.Name = st.Name()
	dl.Bytes = st.Size()
	dl.LastMod = st.ModTime()
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()

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
