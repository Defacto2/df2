package demozoo

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
	"github.com/google/uuid"
)

// Insert contains the values for a new Demozoo releaser production to be added as a database file record.
type Insert struct {
	WebIDDemozoo uint   // Demozoo production id
	ID           string // MySQL auto increment id
	UUID         string // record unique id
	Title        string
	Platform     string
	Section      string
	GroupFor     string
	GroupBy      string
	CreditText   []string
	CreditCode   []string
	CreditArt    []string
	CreditAudio  []string
	IssuedYear   uint16
	IssuedMonth  uint8
	IssuedDay    uint8
}

var ErrNoQuery = errors.New("query statement is empty")

// Prod mutates the raw Demozoo API data to database ready values.
// Except for errors, setting quiet to false disables all stdout feedback.
func Prod(prod releases.ProductionV1, quiet bool) Insert {
	dbID, _ := database.DemozooID(uint(prod.ID))
	if dbID > 0 {
		prod.ExistsInDB = true
		if !quiet {
			fmt.Printf(": skipped, production already exists")
		}
		return Insert{}
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
		if !quiet {
			fmt.Printf("[%s/%s] : skipped, unsuitable production", p, t)
		}
		return Insert{}
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

	var rec Insert
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
	new := 0
	for i, prod := range *p {
		item := fmt.Sprintf("%d. ", i)
		if !quiet {
			fmt.Printf("\n%s%s", item, prod.Title)
		}
		rec := Prod(prod, quiet)
		if reflect.DeepEqual(rec, Insert{}) {
			continue
		}
		res, err := rec.Insert()
		if err != nil {
			return err
		}
		newId, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if !quiet {
			pad := strings.Repeat(" ", len(item))
			fmt.Printf("\n%s ↳ production added using auto-id: %d", pad, newId)
		}
		new++
		if new > 1 {
			break
		}
	}
	if !quiet && new > 0 {
		fmt.Printf("\nAdded %d new releaser productions from Demozoo.\n", new)
	}
	return nil
}

// Insert the new Demozoo releaser production into the database.
func (r *Insert) Insert() (sql.Result, error) {
	db := database.Connect()
	defer db.Close()
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("insert uuid: %w", err)
	}
	query, args, err := r.InsertStmt(id)
	if err != nil {
		return nil, fmt.Errorf("insert stmt: %w", err)
	}
	new, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("insert prepare: %w", err)
	}
	defer new.Close()
	res, err := new.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("insert exec: %w", err)
	}
	return res, nil
}

// InsertStmt creates the SQL prepare statement and values to insert a new Demozoo releaser production.
func (r *Insert) InsertStmt(id uuid.UUID) (query string, args []any, err error) {
	set, args := inserts(r)
	if len(set) == 0 {
		return "", args, ErrNoQuery
	}
	// create an uuid that's required by the file table.
	set = append(set, "uuid")
	args = append(args, id.String())
	// create time values for the new record.
	// setting createdat, updatedat and deletedat tells the webapp that the record is new, unmodifed and not public.
	now := time.Now()
	set = append(set, "createdat")
	args = append(args, []any{now}...)
	set = append(set, "updatedat")
	args = append(args, []any{now}...)
	set = append(set, "deletedat")
	args = append(args, []any{now}...)

	vals := strings.Split(strings.TrimSpace(strings.Repeat("? ", len(args))), " ")
	query = "INSERT INTO files (" + strings.Join(set, sep) + ") VALUES (" + strings.Join(vals, sep) + ")"
	return query, args, nil
}

func inserts(r *Insert) (set []string, args []any) {
	if r.WebIDDemozoo != 0 {
		set = append(set, "web_id_demozoo")
		args = append(args, []any{r.WebIDDemozoo}...)
	}
	if r.Title != "" {
		set = append(set, "record_title")
		args = append(args, []any{r.Title}...)
	}
	if r.Platform != "" {
		set = append(set, "platform")
		args = append(args, []any{r.Platform}...)
	}
	if r.Section != "" {
		set = append(set, "section")
		args = append(args, []any{r.Section}...)
	}
	if r.GroupFor != "" {
		set = append(set, "group_brand_for")
		args = append(args, []any{r.GroupFor}...)
	}
	if r.GroupBy != "" {
		set = append(set, "group_brand_by")
		args = append(args, []any{r.GroupBy}...)
	}
	if r.IssuedYear != 0 {
		set = append(set, "date_issued_year")
		args = append(args, []any{r.IssuedYear}...)
	}
	if r.IssuedMonth != 0 {
		set = append(set, "date_issued_month")
		args = append(args, []any{r.IssuedMonth}...)
	}
	if r.IssuedDay != 0 {
		set = append(set, "date_issued_day")
		args = append(args, []any{r.IssuedDay}...)
	}
	s, a := insertCredits(r)
	set = append(set, s...)
	args = append(args, a...)
	return set, args
}

func insertCredits(r *Insert) (set []string, args []any) {
	if len(r.CreditText) > 0 {
		set = append(set, "credit_text")
		j := strings.Join(r.CreditText, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditCode) > 0 {
		set = append(set, "credit_program")
		j := strings.Join(r.CreditCode, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditArt) > 0 {
		set = append(set, "credit_illustration")
		j := strings.Join(r.CreditArt, sep)
		args = append(args, []any{j}...)
	}
	if len(r.CreditAudio) > 0 {
		set = append(set, "credit_audio")
		j := strings.Join(r.CreditAudio, sep)
		args = append(args, []any{j}...)
	}
	return set, args
}
