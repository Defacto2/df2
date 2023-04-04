// Package acronym handles the initalisms and acronyms frequently used by
// the groups.
package acronym

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
)

var ErrName = errors.New("group name cannot be empty")

type Group struct {
	Name       string
	Initialism string
}

// Get the initialism for the named group.
func (g *Group) Get(db *sql.DB) error {
	if db == nil {
		return database.ErrDB
	}
	if g.Name == "" {
		return ErrName
	}
	row := db.QueryRow("SELECT `initialisms` FROM `groupnames` WHERE `pubname`=?", g.Name)
	if err := row.Scan(&g.Initialism); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get row scan: %w", err)
	}
	return nil
}

// Fix deletes any malformed initialisms in the database and returns the number of rows affected.
func Fix(db *sql.DB) (int64, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	row, err := db.Exec("DELETE FROM `groupnames` WHERE `pubname`=? OR `initialisms`=?", "", "")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("fix exec: %w", err)
	}
	return row.RowsAffected()
}

// Get a group's initialism or acronym.
// For example "Defacto2" would return "df2".
func Get(db *sql.DB, s string) (string, error) {
	if db == nil {
		return "", database.ErrDB
	}
	i := ""
	row := db.QueryRow("SELECT `initialisms` FROM groupnames WHERE `pubname` = ?", s)
	if err := row.Scan(&i); err != nil &&
		strings.Contains(err.Error(), "no rows in result set") {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("initialism %q: %w", s, err)
	}
	return i, nil
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
