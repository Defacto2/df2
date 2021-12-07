package record

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof/internal/stat"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
)

// Record of a file item.
type Record struct {
	ID   string // mysql auto increment id
	UUID string // record unique id
	File string // absolute path to file
	Name string // filename
}

func New(values []sql.RawBytes, path string) Record {
	const id, uuid, filename = 0, 1, 4
	return Record{
		ID:   string(values[id]),
		UUID: string(values[uuid]),
		Name: string(values[filename]),
		File: filepath.Join(path, string(values[uuid])),
	}
}

// approve sets the record to be publically viewable.
func (r Record) approve() error {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	if err != nil {
		return fmt.Errorf("approve update prepare: %w", err)
	}
	defer update.Close()
	_, err = update.Exec(database.UpdateID, r.ID)
	if err != nil {
		return fmt.Errorf("approve update exec: %w", err)
	}
	logs.Printf(" %s", str.Y())
	return nil
}

// iterate through each value.
func (r Record) Iterate(s stat.Stat) error {
	var value string
	for i, raw := range *s.Values {
		value = database.Val(raw)
		switch s.Columns[i] {
		case "id":
			r.printID(&s)
		case "createdat":
			database.DateTime(raw)
		case "filename":
			logs.Printf("%v", value)
		case "file_zip_content":
			if err := r.zip(raw, &s); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

// fileZipContent reads an archive and saves its content to the database.
func (r Record) fileZipContent() (ok bool, err error) {
	a, err := archive.Read(r.File, r.Name)
	if err != nil {
		return false, fmt.Errorf("file zip content archive read: %w", err)
	}
	if err := updateZipContent(r.ID, len(a), strings.Join(a, "\n")); err != nil {
		return false, fmt.Errorf("file zip content update: %w", err)
	}
	return true, nil
}

// updateZipContent sets the file_zip_content column to match content and platform to "image".
func updateZipContent(id string, items int, content string) error {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,platform=? WHERE id=?")
	if err != nil {
		return fmt.Errorf("update zip content db prepare: %w", err)
	}
	defer update.Close()
	if _, err := update.Exec(content, database.UpdateID, "image", id); err != nil {
		return fmt.Errorf("update zip content update exec: %w", err)
	}
	logs.Printf("%d items", items)
	return db.Close()
}

func (r Record) printID(s *stat.Stat) {
	logs.Printcrf("%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.Total)),
		s.Count,
		color.Primary.Sprint(r.ID))
}

func (r Record) zip(col sql.RawBytes, s *stat.Stat) error {
	if col == nil || s.Overwrite {
		logs.Print(" • ")
		if u, err := r.fileZipContent(); !u {
			return nil
		} else if err != nil {
			return fmt.Errorf("zip content: %w", err)
		}
		if err := archive.Proof(r.File, r.Name, r.UUID); err != nil {
			return fmt.Errorf("zip proof: %w", err)
		} else if err := r.approve(); err != nil {
			return fmt.Errorf("zip approve: %w", err)
		}
	}
	return nil
}

// Skip checks if the file of the proof exists.
func Skip(s stat.Stat, r Record, hide bool) (skip bool, err error) {
	if _, err := os.Stat(r.File); os.IsNotExist(err) {
		s.Missing++
		if !hide {
			fmt.Printf("%s %0*d. %v is missing %v %s\n",
				color.Question.Sprint("→"),
				len(strconv.Itoa(s.Total)),
				s.Count,
				color.Primary.Sprint(r.ID),
				filepath.Join(s.Base, color.Danger.Sprint(r.UUID)),
				str.X())
		}
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("file skip stat %q: %w", r.File, err)
	}
	return false, nil
}
