package demozoo

import (
	"fmt"

	"github.com/Defacto2/df2/pkg/database"
)

const selectSQL = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`," +
	"`web_id_demozoo`,`file_zip_content`,`updatedat`,`platform`,`file_integrity_strong`," +
	"`file_integrity_weak`,`web_id_pouet`,`group_brand_for`,`group_brand_by`,`record_title`" +
	",`section`,`credit_illustration`,`credit_audio`,`credit_program`,`credit_text`"

func selectByID(id string) string {
	const w = " FROM `files` WHERE `web_id_demozoo` IS NOT NULL"
	where := w
	if id != "" {
		switch {
		case database.IsUUID(id):
			where = fmt.Sprintf("%v AND `uuid`=%q", w, id)
		case database.IsID(id):
			where = fmt.Sprintf("%v AND `id`=%q", w, id)
		}
	}
	return selectSQL + where
}
