package data

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrApprove  = errors.New("cannot shrink files as there are database records waiting for public approval")
	ErrDel      = errors.New("delete errors")
	ErrArcStore = errors.New("archive store error")
	ErrUnknown  = errors.New("unknown command")
)

// Months of the year.
type Months uint

const (
	non Months = iota // Unknown or non month.
	jan
	feb
	mar
	apr
	may
	jun
	jul
	aug
	sep
	oct
	nov
	dec
)

// Month returns the Months value of s, which should be a three letter English abbreviation.
func Month(s string) Months {
	const monthPrefix = 3
	if len(s) < monthPrefix {
		return non
	}
	months := map[string]Months{
		"jan": jan,
		"feb": feb,
		"mar": mar,
		"apr": apr,
		"may": may,
		"jun": jun,
		"jul": jul,
		"aug": aug,
		"sep": sep,
		"oct": oct,
		"nov": nov,
		"dec": dec,
	}
	for prefix, val := range months {
		if strings.ToLower(s)[:monthPrefix] == prefix {
			return val
		}
	}
	return non
}

type Approvals string

const (
	Preview  Approvals = "previews"
	Incoming Approvals = "incoming"
)

// Approve prints the number of records waiting for approval for public display.
func (cmd Approvals) Approve(db *sql.DB) error {
	switch cmd {
	case Preview, Incoming:
	default:
		return ErrUnknown
	}
	wait, err := database.Waiting(db)
	if err != nil {
		return fmt.Errorf("approve count: %w", err)
	}
	if wait > 0 {
		return fmt.Errorf("%d %s files waiting approval: %w", wait, cmd, ErrApprove)
	}
	return nil
}

// Init SQL directory.
func Init(w io.Writer, directory string) error { //nolint:funlen
	const (
		layout   = "2-1-2006"
		minDash  = 2
		oneMonth = 730
	)
	s := directory
	color.Primary.Printf("SQL directory: %s\n", s)
	entries, err := os.ReadDir(s)
	if err != nil {
		return fmt.Errorf("sql read directory: %w", err)
	}
	cnt, freed, inUse := 0, 0, 0
	files := []string{}
	var create time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		f, err := entry.Info()
		if err != nil {
			fmt.Fprintf(w, "error with file info: %s\n", err)
			continue
		}
		exts := strings.Split(f.Name(), ".")
		dashes := strings.Split(exts[0], "-")
		if len(dashes) < minDash {
			continue
		}
		m := dashes[len(dashes)-minDash]
		if Month(m) == non {
			continue
		}
		cnt++
		inUse += int(f.Size())
		create, err = time.Parse(layout,
			fmt.Sprintf("1-%d-%s", Month(m), dashes[len(dashes)-1]))
		if err != nil {
			log.Printf("error parsing date from %s: %s\n", f.Name(), err)
			continue
		}
		const expire = time.Hour * oneMonth * 2
		if time.Since(create) > expire {
			if filepath.Ext(f.Name()) == ".sql" {
				fmt.Fprintf(w, "%s is to be moved.\n", f.Name())
			}
			files = append(files, filepath.Join(s, f.Name()))
			freed += int(f.Size())
		}
	}
	fmt.Fprintf(w, "SQL found %d files using %s", cnt, humanize.Bytes(uint64(inUse)))
	if len(files) == 0 {
		fmt.Fprintln(w, ", but there is nothing to do.")
		return nil
	}
	fmt.Fprintln(w, ".")
	fmt.Fprintf(w, "SQL will move %d items totaling %s, leaving %s used.\n",
		len(files), humanize.Bytes(uint64(freed)), humanize.Bytes(uint64(inUse-freed)))
	return sqlProcess(w, files)
}

// SaveDir returns a usable path to store backups.
func SaveDir() string {
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir
	}
	var dir string
	dir, err = os.Getwd()
	if err != nil {
		log.Println("shrink saveDir failed to get the user home or the working directory:", err)
		return ""
	}
	return dir
}

// Store incoming or preview files as a tar archive.
func (cmd Approvals) Store(w io.Writer, path, partial string) error {
	switch cmd {
	case Preview, Incoming:
	default:
		return ErrUnknown
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("store read: %w", err)
	}

	cnt, inUse := 0, 0
	files := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		f, err := entry.Info()
		if err != nil {
			return fmt.Errorf("store entry info: %w", err)
		}
		files = append(files, filepath.Join(path, f.Name()))
		cnt++
		inUse += int(f.Size())
	}

	if len(files) == 0 {
		return nil
	}
	fmt.Fprintf(w, "%s found %d files using %s for backup.\n", cmd, cnt, humanize.Bytes(uint64(inUse)))

	n := time.Now()
	filename := filepath.Join(SaveDir(),
		fmt.Sprintf("d2-%s_%d-%02d-%02d.tar", partial, n.Year(), n.Month(), n.Day()))

	if err := storer(files, filename, partial); err != nil {
		return err
	}
	fmt.Fprintf(w, "%s freeing up space is complete.\n", cmd)
	return nil
}

func storer(files []string, filename, partial string) error {
	store, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("store create: %w", err)
	}
	defer store.Close()

	if err := archive.Store(store, files); err != nil {
		return fmt.Errorf("%w: %s", ErrArcStore, partial)
	}
	return archive.Delete(files)
}

// Compress the collection of files into a named archive.
func Compress(w io.Writer, files []string, name string) error {
	if w == nil {
		w = io.Discard
	}
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("sql create: %w", err)
	}
	defer f.Close()
	if err := archive.Compress(f, files); err != nil {
		return err
	}
	fmt.Fprintln(w, "SQL archiving is complete.")
	return nil
}

func sqlProcess(w io.Writer, files []string) error {
	n := time.Now()
	name := filepath.Join(SaveDir(), fmt.Sprintf("d2-sql_%d-%02d-%02d.tar.gz",
		n.Year(), n.Month(), n.Day()))
	if err := Compress(w, files, name); err != nil {
		return err
	}
	return Remove(w, files)
}

// Remove files from the host file system.
func Remove(w io.Writer, files []string) error {
	if w == nil {
		w = io.Discard
	}
	if err := archive.Delete(files); err != nil {
		return err
	}
	fmt.Fprintln(w, "SQL freeing up space is complete.")
	return nil
}
