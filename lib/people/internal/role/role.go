package role

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
)

var (
	ErrRenAll    = errors.New("cannot rename everyone, only people tied to specific roles can be renamed")
	ErrNoName    = errors.New("a name of the person must be provided")
	ErrNoReplace = errors.New("a replacement name must be provided")
	ErrRole      = errors.New("unknown role")
)

// Role are jobs that categorize persons.
type Role int

const (
	Everyone  Role = iota // Everyone displays all people.
	Artists               // Artists are graphic or video.
	Coders                // Coders are programmers.
	Musicians             // Musicians create music or audio.
	Writers               // Writers for the documents.
)

const (
	all       = "all"
	artists   = "artists"
	coders    = "coders"
	musicians = "musicians"
	writers   = "writers"
)

func (r Role) String() string {
	switch r {
	case Everyone:
		return all
	case Artists:
		return artists
	case Coders:
		return coders
	case Musicians:
		return musicians
	case Writers:
		return writers
	default:
		return ""
	}
}

// List people filtered by a role.
func List(role Role) (people []string, total int, err error) {
	db := database.Connect()
	defer db.Close()
	s := PeopleStmt(role, false)
	if s == "" {
		return nil, 0, fmt.Errorf("list statement %v: %w", role, ErrRole)
	}
	if total, err = database.Total(&s); err != nil {
		return nil, 0, fmt.Errorf("list totals: %w", err)
	}
	// interate through records
	rows, err := db.Query(s)
	if err != nil {
		return nil, 0, fmt.Errorf("list query: %w", err)
	} else if rows.Err() != nil {
		return nil, 0, fmt.Errorf("list rows: %w", rows.Err())
	}
	defer rows.Close()
	var p sql.NullString
	i := 0
	for rows.Next() {
		if err = rows.Scan(&p); err != nil {
			return nil, 0, fmt.Errorf("list row scan: %w", err)
		}
		if _, err = p.Value(); err != nil {
			continue
		}
		i++
		people = append(people, fmt.Sprintf("%v", p.String))
	}
	return people, total, nil
}

// PeopleStmt returns a complete SQL WHERE statement for people that are filtered by roles.
func PeopleStmt(role Role, softDel bool) string {
	del := func() string {
		if softDel {
			return ""
		}
		return "AND `deletedat` IS NULL"
	}
	d := func(s string) string {
		return fmt.Sprintf(" UNION (SELECT DISTINCT %s AS pubCombined FROM files WHERE Length(%s) <> 0 %s)",
			s, s, del())
	}
	s := ""
	switch role {
	case Writers:
		s += d("credit_text")
	case Musicians:
		s += d("credit_audio")
	case Coders:
		s += d("credit_program")
	case Artists:
		s += d("credit_illustration")
	case Everyone:
		s += d("credit_text")
		s += d("credit_audio")
		s += d("credit_program")
		s += d("credit_illustration")
	}
	if s == "" {
		return s
	}
	s += " ORDER BY pubCombined"
	s = strings.Replace(s, "UNION (SELECT DISTINCT ", "(SELECT DISTINCT ", 1)
	return s
}

func Roles(r string) Role {
	switch r {
	case writers, "w":
		return Writers
	case musicians, "m":
		return Musicians
	case coders, "c":
		return Coders
	case artists, "a":
		return Artists
	case "", all:
		return Everyone
	}
	return -1
}

// Rename replaces the persons using name with the replacement.
// The task must be limited names associated to a Role.
func Rename(replacement, name string, r Role) (count int64, err error) {
	if replacement == "" {
		return 0, ErrNoReplace
	}
	if name == "" {
		return 0, ErrNoName
	}
	query := ""
	switch r {
	case Artists:
		query = "UPDATE `files` SET credit_illustration=? WHERE credit_illustration=?"
	case Coders:
		query = "UPDATE `files` SET credit_program=? WHERE credit_program=?"
	case Musicians:
		query = "UPDATE `files` SET credit_audio=? WHERE credit_audio=?"
	case Writers:
		query = "UPDATE `files` SET credit_text=? WHERE credit_text=?"
	case Everyone:
		return 0, ErrRenAll
	}
	db := database.Connect()
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("rename people statement: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(replacement, name)
	if err != nil {
		return 0, fmt.Errorf("rename people exec: %w", err)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rename people rows affected: %w", err)
	}
	return count, db.Close()
}

// Clean fixes and saves a malformed name.
func Clean(name string, r Role, sim bool) (ok bool) {
	rep := CleanS(Trim(name))
	if rep == name {
		return false
	}
	if sim {
		logs.Printf("\n%s %q %s %s", color.Question.Sprint("?"), name,
			color.Question.Sprint("!="), color.Info.Sprint(r))
		return true
	}
	s := str.Y()
	ok = true
	c, err := Rename(rep, name, r)
	if err != nil {
		s = str.X()
		ok = false
	}
	logs.Printf("\n%s %q %s %s (%d)", s, name,
		color.Question.Sprint("⟫"), color.Info.Sprint(rep), c)
	return ok
}

// CleanS seperates and fixes the substrings of s.
func CleanS(s string) string {
	ppl := strings.Split(s, ",")
	for i, person := range ppl {
		ss := database.StripStart(person)
		if ss != person {
			ppl[i] = ss
		}
	}
	return strings.Join(ppl, ",")
}

// Trim fixes any malformed strings.
func Trim(s string) string {
	f := database.TrimSP(s)
	f = database.StripChars(f)
	f = database.StripStart(f)
	f = strings.TrimSpace(f)
	return f
}
