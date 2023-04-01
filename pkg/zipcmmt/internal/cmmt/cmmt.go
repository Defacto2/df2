package cmmt

import (
	"archive/zip"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bengarrett/retrotxtgo/lib/convert"
	"go.uber.org/zap/buffer"
)

var (
	ErrID       = errors.New("zipfile id cannot be zero")
	ErrPath     = errors.New("path directory cannot be empty")
	ErrPointer  = errors.New("pointer value cannot be nil")
	ErrUUID     = errors.New("uuid zipfile id cannot be empty")
	ErrZipIsDir = errors.New("path to the named zipfile comment is a directory")
)

const (
	SuffixName = `_zipcomment.txt` // SuffixName is the suffix string to append to the file name.
	SceneOrg   = `scene.org`       // SceneOrg is a search term used to ignore unwanted zipcomments.

	resetCmd = "\033[0m" // ansi command to reset colors and styles.
)

// Zipfile structure for files archived or compressed with a ZIP format.
type Zipfile struct {
	ID        uint           // ID is the database auto increment ID.
	UUID      string         // Universal Unique ID.
	Name      string         // Name of the file.
	Ext       string         // Ext is the extension of the filename.
	Size      int            // Size of the size in bytes.
	Magic     sql.NullString // Magic or MIME type of the file.
	ASCII     bool           // ASCII, plaintext encoded zip comment.
	CP437     bool           // CP437, MS-DOS text encoded zip comment.
	Overwrite bool           // Overwrite a preexisting zip comment.
}

// Exists returns true when the UUID comment file exists in the path.
func (z *Zipfile) Exist(path string) (bool, error) {
	if z.UUID == "" {
		return false, ErrUUID
	}
	if path == "" {
		return false, ErrPath
	}
	src := filepath.Join(path, z.UUID)
	if _, err := os.Stat(src); err != nil {
		return false, fmt.Errorf("%w: %s", err, z.UUID)
	}
	name := filepath.Join(path, z.UUID+SuffixName)
	st, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		// okay when zip comment does not exist
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("%w: %s", err, name)
	}
	if st.IsDir() {
		return false, fmt.Errorf("%w: %s", ErrZipIsDir, name)
	}
	if st.Size() > 0 {
		if z.Overwrite {
			return true, nil
		}
		// skip when the overwrite flag is not set
		return false, nil
	}
	return true, nil
}

// Save an embedded zipfile, text comment to the path.
func (z *Zipfile) Save(w io.Writer, path string) (string, error) {
	if z.UUID == "" {
		return "", ErrUUID
	}
	if path == "" {
		return "", ErrPath
	}
	if w == nil {
		w = io.Discard
	}
	src := filepath.Join(path, z.UUID)
	dest := src + SuffixName
	// Open a zip archive for reading.
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, src)
	}
	defer r.Close()
	// Parse and save zip comment
	cmmt := r.Comment
	if cmmt == "" {
		return "", nil
	}
	if strings.TrimSpace(cmmt) == "" {
		return "", nil
	}
	if strings.Contains(cmmt, SceneOrg) {
		return "", nil
	}
	s, err := z.Format(&cmmt)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, src)
	}
	fmt.Fprint(w, s.String())
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, src)
	}
	defer f.Close()
	if i, err := f.WriteString(cmmt); err != nil {
		return "", fmt.Errorf("%w: %s", err, src)
	} else if i == 0 {
		defer os.Remove(dest)
		return "", nil
	}
	if err = f.Sync(); err != nil {
		return "", err
	}
	return dest, nil
}

func (z *Zipfile) Format(cmmt *string) (buffer.Buffer, error) {
	if cmmt == nil {
		return buffer.Buffer{}, fmt.Errorf("cmmt %w", ErrPointer)
	}
	if *cmmt == "" {
		return buffer.Buffer{}, nil
	}
	if z.ID < 1 {
		return buffer.Buffer{}, ErrID
	}
	b := &buffer.Buffer{}
	fmt.Fprintf(b, "\n%v. - %s", z.ID, z.Name)
	if z.Magic.Valid {
		fmt.Fprintf(b, " [%s]", z.Magic.String)
	}
	if z.CP437 {
		fmt.Fprint(b, z.unicode(cmmt))
	} else {
		fmt.Fprint(b, z.ascii(cmmt))
	}
	return *b, nil
}

func (z *Zipfile) ascii(cmmt *string) string {
	if cmmt == nil || *cmmt == "" {
		return ""
	}
	return fmt.Sprintf("%s%s\n", *cmmt, resetCmd)
}

func (z *Zipfile) unicode(cmmt *string) string {
	if cmmt == nil || *cmmt == "" {
		return ""
	}
	b, err := convert.D437(*cmmt)
	if err != nil {
		return fmt.Sprintf("Could not convert to Unicode:\n%s%s\n", *cmmt, resetCmd)
	}
	return fmt.Sprintf("%s%s\n", b, resetCmd)
}
