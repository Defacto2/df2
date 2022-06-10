package insert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/google/uuid"
)

var ErrNoQuery = errors.New("query statement is empty")

const (
	sep     = ","
	timeout = 5 * time.Second
)

// Record contains the values for a new Demozoo releaser production to be added to the database file table.
type Record struct {
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

// Insert the new Demozoo releaser production into the database.
func (r *Record) Insert() (sql.Result, error) {
	db := database.Connect()
	defer db.Close()
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

// stmt creates the SQL prepare statement and values to insert a new Demozoo releaser production.
func (r *Record) stmt(id uuid.UUID) (query string, args []any, err error) {
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

func inserts(r *Record) (set []string, args []any) {
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

func credits(r *Record) (set []string, args []any) {
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
