package demozoo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/insert"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
)

var ErrNoQuery = errors.New("query statement is empty")

// Prod mutates the raw Demozoo API releaser production data to database ready values.
// Except for errors, setting quiet to false disables all stdout feedback.
func Prod(prod releases.ProductionV1, quiet bool) insert.Record {
	dbID, _ := database.DemozooID(uint(prod.ID))
	if dbID > 0 {
		prod.ExistsInDB = true
		if !quiet {
			fmt.Printf(": skipped, production already exists")
		}
		return insert.Record{}
	}

	p, t := "", ""
	if len(prod.Platforms) > 0 {
		p = prod.Platforms[0].Name
	}
	if len(prod.Types) > 0 {
		t = prod.Types[0].Name
	}
	platform, section := releases.Tags(p, t)
	if platform == "" && section == "" {
		s := ""
		if p != "" {
			s = p
		}
		if t != "" {
			s += " " + t
		}
		if !quiet {
			fmt.Printf(": skipped, unsuitable production [%s]", strings.TrimSpace(s))
		}
		return insert.Record{}
	}
	if !quiet {
		fmt.Printf(" [%s/%s]", platform, section)
	}

	a, b := prod.Groups()
	if a != "" {
		fmt.Printf(" for: %s", a)
	}
	if b != "" {
		fmt.Printf(" by: %s", b)
	}

	y, m, d := prod.Released()

	var rec insert.Record
	rec.WebIDDemozoo = uint(prod.ID)
	rec.Title = strings.TrimSpace(prod.Title)
	rec.Platform = platform
	rec.Section = section
	rec.GroupFor = a
	rec.GroupBy = b
	rec.IssuedYear = uint16(y)
	rec.IssuedMonth = uint8(m)
	rec.IssuedDay = uint8(d)
	return rec
}

// InsertProds adds the Demozoo releasers productions to the database.
// API: https://demozoo.org/api/v1/releasers/
// Except for errors, setting quiet to false disables all stdout feedback.
func InsertProds(p *releases.Productions, quiet bool) error {
	recs := 0
	for i, prod := range *p {
		item := fmt.Sprintf("%d. ", i)
		if !quiet {
			fmt.Printf("\n%s%s", item, prod.Title)
		}
		rec := Prod(prod, quiet)
		if reflect.DeepEqual(rec, insert.Record{}) {
			continue
		}
		res, err := rec.Insert()
		if err != nil {
			return err
		}
		newID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if !quiet {
			pad := strings.Repeat(" ", len(item))
			fmt.Printf("\n%s ↳ production added using auto-id: %d", pad, newID)
		}
		recs++
	}
	if !quiet && recs > 0 {
		fmt.Printf("\nAdded %d new releaser productions from Demozoo.\n", recs)
	}
	return nil
}
