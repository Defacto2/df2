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
	var app Update
	app.query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%application%'"
	app.args = "groupapplication"
	return app.execute()
}

func updateInstallers() (int64, error) {
	var inst Update
	inst.query = "UPDATE files SET section=? WHERE `section` = \"releaseadvert\" AND `web_id_demozoo` IS NOT NULL AND `record_title` LIKE '%installer%'"
	inst.args = "releaseinstall"
	return inst.execute()
}

// Update row values based on conditions and returns the total number of changes.
type Update struct {
	query string
	args  interface{}
}

func (u Update) execute() (int64, error) {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare(u.query)
	if err != nil {
		return 0, err
	}
	res, err := update.Exec(u.args)
	if err != nil {
		return 0, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return ra, nil
}
