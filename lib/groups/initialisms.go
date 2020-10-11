package groups

import (
	"database/sql"
	"errors"
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
		return "", fmt.Errorf("initialism get %q: %w", name, err)
	}
	return g.initialism, nil
}

func (g *group) get() error {
	db, err := database.ConnectErr()
	if err != nil {
		return fmt.Errorf("get db connect: %w", err)
	}
	defer db.Close()
	row := db.QueryRow("SELECT `initialisms` FROM `groups` WHERE `pubname`=?", g.name)
	if err = row.Scan(&g.initialism); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get row scan: %w", err)
	}
	return db.Close()
}
