package demozoo

import (
	"fmt"

	"github.com/Defacto2/df2/pkg/database"
)

const selectStmt = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`," +
	"`web_id_demozoo`,`file_zip_content`,`updatedat`,`platform`,`file_integrity_strong`," +
	"`file_integrity_weak`,`web_id_pouet`,`group_brand_for`,`group_brand_by`,`record_title`" +
	",`section`,`credit_illustration`,`credit_audio`,`credit_program`,`credit_text`"

func selectByID(id string) string {
	stmt := " FROM `files` WHERE `web_id_demozoo` IS NOT NULL "
	if id == "" {
		return selectStmt + stmt
	}
	switch {
	case database.IsUUID(id):
		stmt += fmt.Sprintf("AND `uuid`=%q", id)
	case database.IsID(id):
		stmt += fmt.Sprintf("AND `id`=%q", id)
	}
	return selectStmt + stmt
}

func countPouet() string {
	s := "SELECT COUNT(*) FROM `files` "
	s += "WHERE `web_id_pouet` IS NOT NULL"
	return s
}

func countDemozoo() string {
	s := "SELECT COUNT(*) FROM `files` "
	s += "WHERE `web_id_demozoo` IS NOT NULL"
	return s
}
