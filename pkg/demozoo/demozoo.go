// Package demozoo interacts with the demozoo.org API for data scraping and file downloads.
package demozoo

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/filter"
	"github.com/Defacto2/df2/pkg/demozoo/internal/fix"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prod"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releaser"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
)

var (
	ErrRequest = errors.New("unknown request value")
	ErrValues  = errors.New("too few record values")
)

// Product is a Demozoo production item.
type Product struct {
	Code   int                    // Code is the HTTP status.
	Status string                 // Status is the HTTP status.
	API    prods.ProductionsAPIv1 // API v1 for a Demozoo production.
}

// Get a Demozoo production.
func (p *Product) Get(id uint) error {
	d := prod.Production{ID: int64(id)}
	api, err := d.Get()
	if err != nil {
		return fmt.Errorf("get product id %d: %w", id, err)
	}
	p.Code = d.Code
	p.Status = d.Status
	p.API = api
	return nil
}

// Releaser is a Demozoo scener or group.
type Releaser struct {
	Code   int                 // Code is the HTTP status.
	Status string              // Status is the HTTP status.
	API    releaser.ReleaserV1 // API v1 for a Demozoo releaser.
}

// Get a Demozoo scener or group.
func (r *Releaser) Get(id uint) error {
	d := releaser.Releaser{ID: int64(id)}
	api, err := d.Get()
	if err != nil {
		return fmt.Errorf("get releaser id %d: %w", id, err)
	}
	r.Code = d.Code
	r.Status = d.Status
	r.API = api
	return nil
}

// ReleaserProducts are the productions of a Demozoo releaser.
type ReleaserProducts struct {
	Code   int                  // Code is the HTTP status.
	Status string               // Status is the HTTP status.
	API    releases.Productions // API for the Demozoo productions.
}

// Get the productions of a Demozoo scener or group.
func (r *ReleaserProducts) Get(id uint) error {
	d := releaser.Releaser{ID: int64(id)}
	api, err := d.Prods()
	if err != nil {
		return fmt.Errorf("get releaser prods id %d: %w", id, err)
	}
	r.Code = d.Code
	r.Status = d.Status
	r.API = api
	return nil
}

// MsDosProducts are Demozoo productions that match platforms id 4, MS-DOS.
// Productions with the tag "lost" are skipped.
// Productions created on or newer than 1 Jan. 2000 are skipped.
type MsDosProducts struct {
	Code   int                     // Code is the HTTP status.
	Status string                  // Status is the HTTP status.
	API    []releases.ProductionV1 // API v1 for a Demozoo production.
	Count  int                     // Count the total productions.
	Finds  int                     // Finds are the number of usable productions.
}

// Get all the productions on Demozoo that are for the MS-DOS platform.
func (m *MsDosProducts) Get(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	d := filter.Productions{Filter: releases.MsDos}
	api, err := d.Prods(db, w, 0)
	if err != nil {
		return fmt.Errorf("get msdos prods: %w", err)
	}
	m.Code = d.Code
	m.Status = d.Status
	m.Count = d.Count
	m.Finds = d.Finds
	m.API = api
	return nil
}

// WindowsProducts are Demozoo productions that match the Windows platform.
type WindowsProducts struct {
	Code   int                     // Code is the HTTP status.
	Status string                  // Status is the HTTP status.
	API    []releases.ProductionV1 // API v1 for a Demozoo production.
	Count  int                     // Count the total productions.
	Finds  int                     // Finds are the number of usable productions.
}

// Get all the productions on Demozoo that are for the Windows platform.
func (m *WindowsProducts) Get(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	d := filter.Productions{Filter: releases.Windows}
	api, err := d.Prods(db, w, 0)
	if err != nil {
		return fmt.Errorf("get msdos prods: %w", err)
	}
	m.Code = d.Code
	m.Status = d.Status
	m.Count = d.Count
	m.Finds = d.Finds
	m.API = api
	return nil
}

// Fix any Demozoo data import conflicts.
func Fix(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	return fix.Configs(db, w)
}

// NewRecord initialises a new file record.
func NewRecord(count int, values []sql.RawBytes) (Record, error) {
	const sep, want = ",", 21
	if l := len(values); l < want {
		return Record{}, fmt.Errorf("new records = %d, want %d: %w", l, want, ErrValues)
	}
	const id, uuid, createdat, filename, filesize, webiddemozoo = 0, 1, 3, 4, 5, 6
	const filezipcontent, updatedat, platform, fileintegritystrong, fileintegrityweak = 7, 8, 9, 10, 11
	const webidpouet, groupbrandfor, groupbrandby, recordtitle, section = 12, 13, 14, 15, 16
	const creditillustration, creditaudio, creditprogram, credittext = 17, 18, 19, 20
	r := Record{
		Count: count,
		ID:    string(values[id]),
		UUID:  string(values[uuid]),
		// deletedat placeholder
		Filename: string(values[filename]),
		Filesize: string(values[filesize]),
		// web_id_demozoo placeholder
		FileZipContent: string(values[filezipcontent]),
		Platform:       string(values[platform]),
		Sum384:         string(values[fileintegritystrong]),
		SumMD5:         string(values[fileintegrityweak]),
		// web_id_pouet placeholder
		GroupFor:    string(values[groupbrandfor]),
		GroupBy:     string(values[groupbrandby]),
		Title:       string(values[recordtitle]),
		Section:     string(values[section]),
		CreditArt:   strings.Split(string(values[creditillustration]), sep),
		CreditAudio: strings.Split(string(values[creditaudio]), sep),
		CreditCode:  strings.Split(string(values[creditprogram]), sep),
		CreditText:  strings.Split(string(values[credittext]), sep),
	}
	ca, err := database.DateTime(values[createdat])
	if err != nil {
		return Record{}, fmt.Errorf("create date for new record %d: %w", count, err)
	}
	r.CreatedAt = ca
	ua, err := database.DateTime(values[updatedat])
	if err != nil {
		return Record{}, fmt.Errorf("update date for new record %d: %w", count, err)
	}
	r.UpdatedAt = ua
	if i, err := strconv.Atoi(string(values[webiddemozoo])); err == nil {
		r.WebIDDemozoo = uint(i)
	}
	if i, err := strconv.Atoi(string(values[webidpouet])); err == nil {
		r.WebIDPouet = uint(i)
	}
	return r, nil
}

type request uint

const (
	meta request = iota
	pouet
)

func (r request) String() string {
	return []string{"Demozoo", "Pouet"}[r]
}

// RefreshMeta synchronises missing file entries with Demozoo sourced metadata.
func RefreshMeta(db *sql.DB, w io.Writer) error {
	return refresh(db, w, meta)
}

// RefreshPouet synchronises missing file entries with Demozoo sourced metadata.
func RefreshPouet(db *sql.DB, w io.Writer) error {
	return refresh(db, w, pouet)
}

func refresh(db *sql.DB, w io.Writer, r request) error { //nolint:cyclop
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	switch r {
	case meta, pouet: // ok
	default:
		return ErrRequest
	}
	start := time.Now()
	if err := Counter(db, w, r); err != nil {
		return err
	}
	rows, err := db.Query(selectByID(""))
	if err != nil {
		return fmt.Errorf("meta query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("meta rows: %w", rows.Err())
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("meta columns: %w", err)
	}
	values := make([]sql.RawBytes, len(cols))
	args := make([]any, len(values))
	for i := range values {
		args[i] = &values[i]
	}
	// fetch the rows
	var st Stat
	switch r {
	case meta:
		for rows.Next() {
			if err := st.NextRefresh(db, w, Records{rows, args, values}); err != nil {
				fmt.Fprintf(w, "meta rows: %s\n", err)
			}
		}
	case pouet:
		for rows.Next() {
			if err := st.NextPouet(db, w, Records{rows, args, values}); err != nil {
				fmt.Fprintf(w, "meta rows: %s\n", err)
			}
		}
	}
	st.summary(w, time.Since(start))
	return nil
}

// Counter prints to the writer the number of records with links to the request.
func Counter(db *sql.DB, w io.Writer, r request) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	cnt := 0
	var stmt string
	switch r {
	case meta:
		stmt = countDemozoo()
	case pouet:
		stmt = countPouet()
	default:
		return ErrRequest
	}
	if err := db.QueryRow(stmt).Scan(&cnt); err != nil {
		return fmt.Errorf("counter row query: %w", err)
	}
	fmt.Fprintf(w, "There are %d records with %s links\n", cnt, r)
	return nil
}
