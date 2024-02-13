// Package insert adds Demozoo productions to the Defacto2 database.
package insert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
	"github.com/google/uuid"
)

var (
	ErrID      = errors.New("production id must be 1 or higher")
	ErrNoQuery = errors.New("query statement is empty")
	ErrProd    = errors.New("productions pointer cannot be nil")
)

const (
	sep     = ","
	timeout = 10 * time.Second
)

// Record contains the values for a new Demozoo releaser production to be added to the database file table.
type Record struct {
	WebIDDemozoo uint   // Demozoo production ID.
	ID           string // MySQL auto increment ID.
	UUID         string // Unique ID for the record.
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

// Insert the new Demozoo releaser production into the database.
func (r *Record) Insert(db *sql.DB) (sql.Result, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("insert uuid: %w", err)
	}
	query, args, err := r.stmt(id)
	if err != nil {
		return nil, fmt.Errorf("insert stmt: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("insert prepare: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("insert exec: %w", err)
	}
	return res, nil
}

// Prods adds the Demozoo releasers productions to the database.
// API: https://demozoo.org/api/v1/releasers/
func Prods(db *sql.DB, w io.Writer, prods *releases.Productions) error {
	if db == nil {
		return database.ErrDB
	}
	if prods == nil {
		return ErrProd
	}
	if w == nil {
		w = io.Discard
	}
	recs := 0
	for i, prod := range *prods {
		item := fmt.Sprintf("%d. ", i)
		fmt.Fprintf(w, "\n%s%s", item, prod.Title)
		rec, err := Prod(db, w, prod)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(rec, Record{}) {
			continue
		}
		res, err := rec.Insert(db)
		if err != nil {
			return err
		}
		newID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		pad := strings.Repeat(" ", len(item))
		fmt.Fprintf(w, "\n%s ↳ production added using auto-id: %d", pad, newID)
		recs++
	}
	if recs > 0 {
		fmt.Fprintf(w, "\nAdded %d new releaser productions from Demozoo.\n", recs)
	}
	return nil
}

// Prod mutates the raw Demozoo API releaser production data to database ready values.
func Prod(db *sql.DB, w io.Writer, prod releases.ProductionV1) (Record, error) {
	if db == nil {
		return Record{}, database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if prod.ID < 1 {
		return Record{}, fmt.Errorf("%w %d", ErrID, prod.ID)
	}
	id, err := database.DemozooID(db, uint(prod.ID))
	if err != nil {
		return Record{}, err
	}
	if id > 0 {
		prod.ExistsInDB = true
		fmt.Fprintf(w, ": skipped, production already exists")
		return Record{}, nil
	}

	p, t := "", ""
	if len(prod.Platforms) > 0 {
		p = prod.Platforms[0].Name
	}
	if len(prod.Types) > 0 {
		t = prod.Types[0].Name
	}
	platform, section := releases.Tags(p, t, prod.Title)
	if platform == "" && section == "" {
		s := ""
		if p != "" {
			s = p
		}
		if t != "" {
			s += " " + t
		}
		fmt.Fprintf(w, ": skipped, unsuitable production [%s]", strings.TrimSpace(s))
		return Record{}, nil
	}
	fmt.Fprintf(w, " [%s/%s]", platform, section)

	a, b := prod.Groups()
	if a != "" {
		fmt.Fprintf(w, " for: %s", a)
	}
	if b != "" {
		fmt.Fprintf(w, " by: %s", b)
	}
	y, m, d := prod.Released()
	return Record{
		WebIDDemozoo: uint(prod.ID),
		Title:        strings.TrimSpace(prod.Title),
		Platform:     platform,
		Section:      section,
		GroupFor:     a,
		GroupBy:      b,
		IssuedYear:   uint16(y),
		IssuedMonth:  uint8(m),
		IssuedDay:    uint8(d),
	}, nil
}

// stmt creates the SQL prepare statement and values to insert a new Demozoo releaser production.
func (r *Record) stmt(id uuid.UUID) (string, []any, error) {
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
	query := "INSERT INTO files (" + strings.Join(set, sep) + ") VALUES (" + strings.Join(vals, sep) + ")"
	return query, args, nil
}

func inserts(r *Record) ([]string, []any) {
	var args []any
	set := []string{}
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
	s, a := credits(r)
	set = append(set, s...)
	args = append(args, a...)
	return set, args
}

func credits(r *Record) ([]string, []any) {
	args, set := []any{}, []string{}
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
