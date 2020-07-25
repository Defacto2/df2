package demozoo

import (
	"fmt"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
)

// Fix repairs imported Demozoo data conflicts.
func Fix() error {
	res, err := updateApplications()
	if err != nil {
		return fmt.Errorf("demozoo fix: %w", err)
	}
	logs.Println("moved", res, "Demozoo #releaseadvert records to #groupapplication")
	res, err = updateInstallers()
	if err != nil {
		return fmt.Errorf("demozoo fix: %w", err)
	}
	logs.Println("moved", res, "Demozoo #releaseadvert records to #releaseinstall")
	return nil
}

func updateApplications() (count int64, err error) {
	var app database.Update
	app.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%application%'"
	app.Args = []interface{}{"groupapplication"}
	if count, err = app.Execute(); err != nil {
		return 0, fmt.Errorf("update applications: %w", err)
	}
	return count, nil
}

func updateInstallers() (count int64, err error) {
	var inst database.Update
	inst.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%installer%'"
	inst.Args = []interface{}{"releaseinstall"}
	if count, err = inst.Execute(); err != nil {
		return 0, fmt.Errorf("update installers: %w", err)
	}
	return count, nil
}
