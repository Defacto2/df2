package fix

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/database"
)

// Fix repairs imported Demozoo data conflicts.
func Configs(db *sql.DB, w io.Writer) error {
	res, err := updateApps(db)
	if err != nil {
		return fmt.Errorf("demozoo fix: %w", err)
	}
	fmt.Fprintln(w, "moved", res, "Demozoo #releaseadvert records to #groupapplication")
	res, err = updateInstallers(db)
	if err != nil {
		return fmt.Errorf("demozoo fix: %w", err)
	}
	fmt.Fprintln(w, "moved", res, "Demozoo #releaseadvert records to #releaseinstall")
	return nil
}

func updateApps(db *sql.DB) (int64, error) {
	var app database.Update
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
	var inst database.Update
	inst.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" " +
		"AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%installer%'"
	inst.Args = []any{"releaseinstall"}
	count, err := database.Execute(db, inst)
	if err != nil {
		return 0, fmt.Errorf("update installers: %w", err)
	}
	return count, nil
}
