package acronym

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Defacto2/df2/lib/database"
)

type Group struct {
	Name       string
	Initialism string
}

func (g *Group) Get() error {
	db, err := database.ConnectErr()
	if err != nil {
		return fmt.Errorf("get db connect: %w", err)
	}
	defer db.Close()
	row := db.QueryRow("SELECT `initialisms` FROM `groups` WHERE `pubname`=?", g.Name)
	if err = row.Scan(&g.Initialism); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get row scan: %w", err)
	}
	return db.Close()
}
