package people

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/campoy/unique"
)

// Request flags for people functions.
type Request struct {
	// Filter people by category.
	Filter string
	// Counts the person's total files.
	Counts bool
	// Progress counter when requesting database data.
	Progress bool
}

// Filters are peoples' roles.
const Filters = "artists,coders,musicians,writers"

// List people filtered by a role.
func List(role string) ([]string, int) {
	db := database.Connect()
	defer db.Close()
	s := sqlPeople(role, false)
	total := database.Total(s)
	// interate through records
	rows, err := db.Query(s)
	logs.Check(err)
	var persons sql.NullString
	i := 0
	g := ""
	ppl := []string{}
	for rows.Next() {
		err = rows.Scan(&persons)
		logs.Check(err)
		if _, err = persons.Value(); err != nil {
			continue
		}
		i++
		g = fmt.Sprintf("%v", persons.String)
		ppl = append(ppl, g)
	}
	return ppl, total
}

// Print lists people filtered by a role and summaries the results.
func Print(r Request) {
	ppl, total := List(r.Filter)
	println(total, "matching", r.Filter, "records found")
	var a []string

	for i := range ppl {
		if r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		// role
		n := ppl[i]
		s := n
		// file totals
		// if r.Counts {
		// 	if c := Count(n); c > 0 {
		// 		s = fmt.Sprintf("%v (%d)", s, c)
		// 	}
		// }
		x := strings.Split(s, ",")
		a = append(a, x...)
	}
	// title and sort names
	for i := range a {
		if r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		a[i] = strings.Title(a[i])
	}
	sort.Strings(a)
	// remove duplicates
	less := func(i, j int) bool { return a[i] < a[j] }
	unique.Slice(&a, less)
	fmt.Println()
	fmt.Println(strings.Join(a, ","))
	fmt.Println("Total authors", len(a))
}

// sqlPeople returns a complete SQL WHERE statement where the people are filtered by a role.
func sqlPeople(role string, includeSoftDeletes bool) string {
	inc := includeSoftDeletes
	f := ""
	switch role {
	case "writers", "writer", "w":
		f = "w"
	case "musicians", "music", "m":
		f = "m"
	case "coders", "code", "c":
		f = "c"
	case "artists", "art", "a":
		f = "a"
	case "", "all":
		f = "wmca"
	}
	var sql string
	if strings.ContainsAny(f, "w") {
		sql += " UNION (SELECT DISTINCT credit_text AS pubCombined FROM files WHERE Length(credit_text) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	if strings.ContainsAny(f, "m") {
		sql += " UNION (SELECT DISTINCT credit_audio AS pubCombined FROM files WHERE Length(credit_audio) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	if strings.ContainsAny(f, "c") {
		sql += " UNION (SELECT DISTINCT credit_program AS pubCombined FROM files WHERE Length(credit_program) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	if strings.ContainsAny(f, "a") {
		sql += " UNION (SELECT DISTINCT credit_illustration AS pubCombined FROM files WHERE Length(credit_illustration) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	sql += " ORDER BY pubCombined"
	sql = strings.Replace(sql, "UNION (SELECT DISTINCT ", "(SELECT DISTINCT ", 1)
	return sql
}

// sqlPeopleDel returns a partial SQL WHERE to handle soft deleted entries.
func sqlPeopleDel(includeSoftDeletes bool) string {
	sql := ""
	if includeSoftDeletes {
		return sql
	}
	sql += "AND `deletedat` IS NULL"
	return sql
}
