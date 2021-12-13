package acronym

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Defacto2/df2/lib/database"
)

type Group struct {
	Name       string
	Initialism string
}

// Get the initialism for the named group.
func (g *Group) Get() error {
	db, err := database.ConnectErr()
	if err != nil {
		return fmt.Errorf("get connect: %w", err)
	}
	defer db.Close()
	row := db.QueryRow("SELECT `initialisms` FROM `groups` WHERE `pubname`=?", g.Name)
	if err = row.Scan(&g.Initialism); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get row scan: %w", err)
	}
	return db.Close()
}

// Get a group's initialism or acronym.
// For example "Defacto2" would return "df2".
func Get(s string) (string, error) {
	db := database.Connect()
	defer db.Close()
	var i string
	row := db.QueryRow("SELECT `initialisms` FROM groups WHERE `pubname` = ?", s)
	if err := row.Scan(&i); err != nil &&
		strings.Contains(err.Error(), "no rows in result set") {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("initialism %q: %w", s, err)
	}
	return i, db.Close()
}

// Trim removes a (bracketed initialism) from s.
// For example "Defacto2 (DF2)" would return "Defacto2".
func Trim(s string) string {
	s = strings.TrimSpace(s)
	a := strings.Split(s, " ")
	l := a[len(a)-1]
	if l[:1] == "(" && l[len(l)-1:] == ")" {
		return strings.Join(a[:len(a)-1], " ")
	}
	return s
}
