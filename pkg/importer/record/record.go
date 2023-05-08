// Package record creates an object that can be written as JSON or used as a new
// database record.
package record

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/importer/arcade"
	"github.com/Defacto2/df2/pkg/importer/zone"
	"github.com/Defacto2/df2/pkg/importer/zwt"
	models "github.com/Defacto2/df2/pkg/models/mysql"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
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
	LastMod    time.Time `json:"file_last_modified"`
	ZipContent string    `json:"file_zip_content"`
	Published  time.Time `json:"date_issued"`
	Section    string    `json:"section"`
	Platform   string    `json:"platform"`
	Comment    string    `json:"comment"`
	Readme     string    `json:"retrotxt_readme"`
	TempPath   string    `json:"temp_path"` // TempPath to the temporary UUID named file download.
}

func (rec Record) Insert(ctx context.Context, db *sql.DB, newpath string) error {
	f1 := models.File{}
	if _, err := uuid.Parse(rec.UUID); err != nil {
		return fmt.Errorf("%w: %s", err, rec.UUID)
	}
	f1.UUID = null.NewString(rec.UUID, true)
	f1.RecordTitle = null.NewString(rec.Title, true)
	f1.GroupBrandFor = null.NewString(rec.Group, true)
	if !rec.Published.IsZero() {
		f1.DateIssuedYear = null.Int16From(int16(rec.Published.Year()))
		f1.DateIssuedMonth = null.Int8From(int8(rec.Published.Month()))
		f1.DateIssuedDay = null.Int8From(int8(rec.Published.Day()))
	}
	f1.Filename = null.NewString(rec.FileName, true)
	f1.Filesize = null.NewInt(int(rec.FileSize), true)
	f1.FileMagicType = null.NewString(rec.FileMagic, true)
	if rec.ZipContent != "" {
		f1.FileZipContent = null.NewString(rec.ZipContent, true)
	}
	f1.FileIntegrityStrong = null.NewString(rec.HashStrong, true)
	f1.FileIntegrityWeak = null.NewString(rec.HashWeak, true)
	f1.FileLastModified = null.NewTime(rec.LastMod, true)
	f1.Platform = null.NewString(rec.Platform, true)
	f1.Section = null.NewString(rec.Section, true)
	f1.Comment = null.NewString(rec.Comment, true)
	if rec.Readme != "" {
		f1.RetrotxtReadme = null.NewString(rec.Readme, true)
	}
	// hide the new record from public access
	f1.Updatedby = null.NewString(database.UpdateID, true)
	f1.Deletedat = null.NewTime(time.Now(), true)
	err := f1.Insert(ctx, db, boil.Infer()) // Insert the first pilot with name "Larry"
	if err != nil {
		defer os.Remove(newpath)
		return err
	}
	return nil
}

// Records are a collection of Record items to insert into the database.
type Records []Record

func (imports Records) Insert(
	ctx context.Context, db *sql.DB, l *zap.SugaredLogger, path string, limit uint,
) (int, error) {
	inserts := 0
	for i, rec := range imports {
		if limit > 0 && i > int(limit) {
			break
		}
		query := fmt.Sprintf("SELECT COUNT(*) AS count FROM `files` WHERE `file_integrity_strong` = '%s'",
			rec.HashStrong)
		rows, err := db.Query(query)
		if err != nil {
			return 0, err
		}
		count := 0
		for rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return 0, err
			}
		}
		if count > 0 {
			l.Errorf("SKIP, the hash matches a database entry: %q\n\t%s",
				rec.Title, rec.HashStrong)
			continue
		}
		newpath := filepath.Join(path, rec.UUID)
		if err := os.Rename(rec.TempPath, newpath); err != nil {
			return 0, err
		}
		if err := rec.Insert(ctx, db, newpath); err != nil {
			return 0, err
		}
		l.Infof("✽ ADDED %s", rec.Title)
		inserts++
	}
	return inserts, nil
}

// New creates a Record.
// The uid must be a valid UUID or returns an error.
// The name must be the subdirectory of the release.
// The group must be the formal release-group name.
func New(uid, name, group string) (Record, error) {
	if uid == "" || name == "" || group == "" {
		return Record{}, ErrNew
	}
	if _, err := uuid.Parse(uid); err != nil {
		return Record{}, fmt.Errorf("%w: %s", err, uid)
	}
	return Record{
		UUID:     uid,
		Slug:     name,
		Group:    group,
		Section:  Section,
		Platform: Platform,
		Comment:  fmt.Sprintf("release directory: %s", name),
	}, nil
}

// Download file metadata, the download is usually either a ZIP archive
// or a single textfile such as an NFO or file_id.diz.
type Download struct {
	Path       string    // Path to the file that is open for reading and checksums.
	Name       string    // Name of the base file.
	Bytes      int64     // Bytes size of the file.
	HashStrong string    // HashStrong is the SHA-386 checksum.
	HashWeak   string    // HashWeak is the MD5 checksum.
	Magic      string    // Magic file type.
	ModTime    time.Time // ModTime is the last modification date of the file.
	ReadTitle  string    // Title of the release, read from a file_id.diz.
	ReadDate   time.Time // Publish date of the release, read from a file_id.diz.
}

// Create a download from the named file.
// The group must be the formal release-group name.
func (dl *Download) Create(name, group string) error {
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
	dl.Path = name
	dl.Name = st.Name()
	dl.Bytes = st.Size()
	dl.ModTime = st.ModTime()
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()

	// strong hashes require the named file to be reopened after being read.
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
	// weak hashes require the named file to be reopened after being read.
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

// ReadDIZ sets the title and publish date of the download using
// the body string sourced from a file_id.diz metadata file.
func (dl *Download) ReadDIZ(body, group string) error {
	var (
		m          time.Month
		y, d       int
		pub, title string
	)
	switch strings.ToLower(group) {
	case "":
		return ErrGroup
	case "zone":
		y, m, d = zone.DizDate(body)
		title = zone.DizTitle(body)
	case "zwt", strings.ToLower(zwt.Name):
		y, m, d = zwt.DizDate(body)
		title, pub = zwt.DizTitle(body)
	default:
		// in the future, handle generic dizdate, title etc?
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

// ReadNfo sets the publish date of the download using
// the body string sourced from a group.nfo file.
// This is a fallback for when ReadDIZ cannot be used.
func (dl *Download) ReadNfo(body, group string) error {
	var (
		m    time.Month
		y, d int
	)
	switch strings.ToLower(group) {
	case "":
		return ErrGroup
	case "arcade":
		y, m, d = arcade.NfoDate(body)
	case "zone", strings.ToLower(zone.Name):
		y, m, d = zone.NfoDate(body)
	case "zwt", strings.ToLower(zwt.Name):
		y, m, d = zwt.NfoDate(body)
	default:
		return nil
	}
	if y > 0 {
		dl.ReadDate = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}
	return nil
}
