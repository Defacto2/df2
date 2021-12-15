package fix

import (
	"fmt"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
)

// Fix repairs imported Demozoo data conflicts.
func Configs() error {
	res, err := updateApps()
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

func updateApps() (int64, error) {
	var app database.Update
	app.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" " +
		"AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%application%'"
	app.Args = []interface{}{"groupapplication"}
	count, err := database.Execute(app)
	if err != nil {
		return 0, fmt.Errorf("update applications: %w", err)
	}
	return count, nil
}

func updateInstallers() (int64, error) {
	var inst database.Update
	inst.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" " +
		"AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%installer%'"
	inst.Args = []interface{}{"releaseinstall"}
	count, err := database.Execute(inst)
	if err != nil {
		return 0, fmt.Errorf("update installers: %w", err)
	}
	return count, nil
}
