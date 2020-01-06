package proof

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

// Record of a file item.
type Record struct {
	ID   string // mysql auto increment id
	UUID string // record unique id
	File string // absolute path to file
	Name string // filename
}

var (
	proofID string // ID used for proofs, either a UUID or ID string
)

// Query parses a single proof.
func Query(id string, ow bool, all bool) error {
	if !database.UUID(id) && !database.ID(id) {
		return fmt.Errorf("invalid id given %q it needs to be an auto-generated MySQL id or an uuid", id)
	}
	proofID = id
	return Queries(ow, all, false)
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func Queries(ow bool, all bool, miss bool) error {
	db := database.Connect()
	defer db.Close()
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`file_zip_content`,`updatedat`,`platform`"
	w := "WHERE `section` = 'releaseproof'"
	if proofID != "" {
		switch {
		case database.UUID(proofID):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, proofID)
		case database.ID(proofID):
			w = fmt.Sprintf("%v AND `id`=%q", w, proofID)
		}
	}
	rows, err := db.Query(s + "FROM `files`" + w)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	// more information: https://github.com/go-sql-driver/mysql/wiki/Examples
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	dir := directories.Init(false)
	// fetch the rows
	cnt := 0
	missing := 0
	base := logs.Path(dir.UUID)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := recordNew(values); !new && !all {
			continue
		}
		cnt++
		r := Record{ID: string(values[0]), UUID: string(values[1]), Name: string(values[4])}
		r.File = filepath.Join(dir.UUID, r.UUID)
		// ping file
		if _, err := os.Stat(r.File); os.IsNotExist(err) {
			missing++
			if !miss {
				fmt.Printf("%s item %v (%v) missing %v\n", logs.X(), cnt, r.ID, filepath.Join(base, color.Danger.Sprint(r.UUID)))
			}
			continue
		}
		// iterate through each value
		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "id":
				fmt.Printf("%s item %04d (%v) ", logs.Y(), cnt, value) // cnt has 3 leading zeros
			case "uuid":
				fmt.Printf("%v ", value)
			case "createdat":
				t, err := time.Parse("2006-01-02T15:04:05Z", value)
				logs.Check(err)
				if t.UTC().Format("01 2006") != time.Now().Format("01 2006") {
					fmt.Printf("%v ", color.Info.Sprint(t.UTC().Format("2 Jan 2006")))
				} else {
					fmt.Printf("%v ", color.Info.Sprint(t.UTC().Format("2 Jan 15:04")))
				}
			case "filename":
				fmt.Printf("%v\n   • ", value)
			case "file_zip_content":
				if col == nil || ow {
					if u := r.fileZipContent(); !u {
						continue
					}
					if err := archive.Extract(r.File, r.UUID); err != nil {
						logs.Log(err)
					} else {
						err = r.approve()
						logs.Log(err)
					}
				}
			case "deletedat":
			case "updatedat": // ignore
			default:
				//fmt.Printf("  %v: %v\n", columns[i], value)
			}
		}
		println()
	}
	logs.Check(rows.Err())
	t := fmt.Sprintf("Total proofs handled: %v", cnt)
	fmt.Println(strings.Repeat("─", len(t)))
	fmt.Println(t)
	if missing > 0 {
		fmt.Println("UUID files not found:", missing)
	}
	return nil
}

// approve sets the record to be publically viewable.
func (r Record) approve() error {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	logs.Check(err)
	_, err = update.Exec(database.UpdateID, r.ID)
	logs.Check(err)
	print(fmt.Sprintf("  %s approved", logs.Y()))
	return nil
}

// fileZipContent reads an archive and saves its content to the database
func (r Record) fileZipContent() bool {
	a, err := archive.Read(r.File)
	if err != nil {
		logs.Log(err)
		return false
	}
	updateZipContent(r.ID, strings.Join(a, "\n"))
	return true
}

// recordNew reports if a proof is set to unapproved
func recordNew(values []sql.RawBytes) bool {
	if values[2] == nil || string(values[2]) != string(values[3]) {
		return false
	}
	return true
}

// updateZipContent sets the file_zip_content column to match content and platform to "image".
func updateZipContent(id string, content string) {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,platform=? WHERE id=?")
	logs.Check(err)
	_, err = update.Exec(content, database.UpdateID, "image", id)
	logs.Check(err)
	fmt.Printf("%s file_zip_content", logs.Y())
}
