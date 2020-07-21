package groups

import (
	"database/sql"
	"fmt"

	"github.com/Defacto2/df2/lib/database"
)

type group struct {
	name       string
	initialism string
}

// Initialism returns a named group initialism or acronym.
func Initialism(name string) (string, error) {
	g := group{name: name}
	if err := g.get(); err != nil {
		return "", err
	}
	return g.initialism, nil
}

func (g group) sqlInitialism() string {
	return fmt.Sprintf("SELECT `initialisms` FROM `groups` WHERE `pubname`=%q", g.name)
}

func (g *group) get() error {
	db, err := database.ConnectErr()
	if err != nil {
		return err
	}
	defer db.Close()
	var pubname string
	row := db.QueryRow(g.sqlInitialism())
	if err = row.Scan(&pubname); err != nil && err != sql.ErrNoRows {
		return err
	}
	g.initialism = pubname
	return db.Close()
}
