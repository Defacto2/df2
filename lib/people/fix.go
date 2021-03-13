package people

import (
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
)

// Fix any malformed group names found in the database.
func Fix(simulate bool) error {
	c, start := 0, time.Now()
	for _, r := range []Role{Artists, Coders, Musicians, Writers} {
		fmt.Println(r.String())
		credits, _, err := List(r)
		if err != nil {
			return err
		}
		for _, credit := range credits {
			if r := cleanPeople(credit, r, simulate); r {
				c++
			}
		}
	}
	switch {
	case c > 0 && simulate:
		logs.Printcrf("%d fixes required", c)
		logs.Simulate()
	case c == 1:
		logs.Printcr("1 fix applied")
	case c > 0:
		logs.Printcrf("%d fixes applied", c)
	default:
		logs.Printcr("no fixes needed")
	}
	elapsed := time.Since(start).Seconds()
	logs.Print(fmt.Sprintf(", time taken %.1f seconds\n", elapsed))

	return nil
}

// cleanPeople fixes and saves a malformed group name.
func cleanPeople(credit string, r Role, sim bool) (ok bool) {
	rep := cleanString(credit)
	if rep == credit {
		return false
	}
	if sim {
		logs.Printf("\n%s %q %s %s", color.Question.Sprint("?"), credit,
			color.Question.Sprint("!="), color.Info.Sprint(r))
		return true
	}
	s := str.Y()
	ok = true
	c, err := rename(rep, credit, r)
	if err != nil {
		s = str.X()
		ok = false
	}
	logs.Printf("\n%s %q %s %s (%d)", s, credit, color.Question.Sprint("⟫"), color.Info.Sprint(rep), c)
	return ok
}

func rename(replacement, credits string, r Role) (count int64, err error) {
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
	}
	db := database.Connect()
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("rename people statement: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(replacement, credits)
	if err != nil {
		return 0, fmt.Errorf("rename people exec: %w", err)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rename people rows affected: %w", err)
	}
	return count, db.Close()
}

// cleanString fixes any malformed strings.
func cleanString(s string) string {
	f := database.TrimSP(s)
	f = database.StripChars(f)
	f = database.StripStart(f)
	f = strings.TrimSpace(f)
	return f
}
