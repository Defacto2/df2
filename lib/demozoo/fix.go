package demozoo

import (
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
)

// Fix repairs imported Demozoo data conflicts.
func Fix() {
	var err error
	var res int64
	res, err = updateApplications()
	logs.Check(err)
	logs.Println("moved", res, "Demozoo #releaseadvert records to #groupapplication")
	res, err = updateInstallers()
	logs.Check(err)
	logs.Println("moved", res, "Demozoo #releaseadvert records to #releaseinstall")
}

func updateApplications() (int64, error) {
	var app database.Update
	app.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%application%'"
	app.Args = []interface{}{"groupapplication"}
	return app.Execute()
}

func updateInstallers() (int64, error) {
	var inst database.Update
	inst.Query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%installer%'"
	inst.Args = []interface{}{"releaseinstall"}
	return inst.Execute()
}
