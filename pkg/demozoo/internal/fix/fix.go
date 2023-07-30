// Package fix repairs any imported Demozoo production data that may conflict
// or be incompatible with the Defacto2 database.
package fix

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/str"
)

// Fix repairs imported Demozoo data conflicts.
func Configs(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	res1, err := updateApps(db)
	if err != nil {
		return fmt.Errorf("demozoo app fix: %w", err)
	}
	if res1 > 0 {
		fmt.Fprintln(w, "\tMOVED", res1, "Demozoo #releaseadvert records to #groupapplication")
	}
	res2, err := updateInstallers(db)
	if err != nil {
		return fmt.Errorf("demozoo installer fix: %w", err)
	}
	if res2 > 0 {
		fmt.Fprintln(w, "\tMOVED", res2, "Demozoo #releaseadvert records to #releaseinstall")
	}
	if res1 == 0 && res2 == 0 {
		fmt.Fprintf(w, "%s%s\n", str.PrePad, str.NothingToDo)
	}
	return nil
}

func updateApps(db *sql.DB) (int64, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	app := database.Update{}
	app.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" " +
		"AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%application%'"
	app.Args = []any{"groupapplication"}
	count, err := database.Execute(db, app)
	if err != nil {
		return 0, fmt.Errorf("update applications: %w", err)
	}
	return count, nil
}

func updateInstallers(db *sql.DB) (int64, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	inst := database.Update{}
	inst.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" " +
		"AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%installer%'"
	inst.Args = []any{"releaseinstall"}
	count, err := database.Execute(db, inst)
	if err != nil {
		return 0, fmt.Errorf("update installers: %w", err)
	}
	return count, nil
}
