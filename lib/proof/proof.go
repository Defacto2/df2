package proof

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
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
	if !database.IsUUID(id) && !database.IsID(id) {
		return fmt.Errorf("invalid id given %q it needs to be an auto-generated MySQL id or an uuid", id)
	}
	proofID = id
	return Queries(ow, all)
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func Queries(ow bool, all bool) error {
	db := database.Connect()
	defer db.Close()
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`file_zip_content`,`updatedat`,`platform`"
	w := "WHERE `section` = 'releaseproof'"
	if proofID != "" {
		switch {
		case database.IsUUID(proofID):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, proofID)
		case database.IsID(proofID):
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
	// todo move to sep func to allow individual record parsing
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
			fmt.Printf("✗ item %v (%v) missing %v\n", cnt, r.ID, r.File)
			missing++
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
				fmt.Printf("✓ item %v (%v) ", cnt, value)
			case "uuid":
				fmt.Printf("%v, ", value)
			case "createdat":
				fmt.Printf("%v, ", value)
			case "filename":
				fmt.Printf("%v\n", value)
			case "file_zip_content":
				if col == nil || ow {
					if u := fileZipContent(r); !u {
						continue
					}
					// todo: tag platform based on found files
					err := archive.Extract(r.File, r.UUID)
					logs.Log(err)
				}
			case "deletedat":
			case "updatedat": // ignore
			default:
				fmt.Printf("   %v: %v\n", columns[i], value)
			}
		}
		fmt.Println("---------------")
	}
	logs.Check(rows.Err())
	fmt.Println("Total proofs handled: ", cnt)
	if missing > 0 {
		fmt.Println("UUID files not found: ", missing)
	}
	return nil
}

func fileZipContent(r Record) bool {
	a, err := archive.Read(r.File)
	if err != nil {
		logs.Log(err)
		return false
	}
	database.Update(r.ID, strings.Join(a, "\n"))
	return true
}

func recordNew(values []sql.RawBytes) bool {
	if values[2] == nil || string(values[2]) != string(values[3]) {
		return false
	}
	return true
}
