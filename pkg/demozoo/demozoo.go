// Package demozoo interacts with the demozoo.org API for data scraping and file downloads.
package demozoo

import (
	"database/sql"
	"fmt"
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
	"github.com/Defacto2/df2/pkg/logs"
)

// Product is a demozoo production.
type Product struct {
	Code   int
	Status string
	API    prods.ProductionsAPIv1
}

func (p *Product) Get(id uint) error {
	d := prod.Production{ID: int64(id)}
	api, err := d.Get()
	if err != nil {
		return fmt.Errorf("get product id %d: %w", id, err)
	}
	p.Code = d.StatusCode
	p.Status = d.Status
	p.API = api
	return nil
}

// Releaser is a demozoo scener or group.
type Releaser struct {
	Code   int
	Status string
	API    releaser.ReleaserV1
}

func (r *Releaser) Get(id uint) error {
	d := releaser.Releaser{ID: int64(id)}
	api, err := d.Get()
	if err != nil {
		return fmt.Errorf("get releaser id %d: %w", id, err)
	}
	r.Code = d.StatusCode
	r.Status = d.Status
	r.API = api
	return nil
}

// ReleaserProducts are the productions of a demozoo releaser.
type ReleaserProducts struct {
	Code   int
	Status string
	API    releases.Productions
}

func (r *ReleaserProducts) Get(id uint) error {
	d := releaser.Releaser{ID: int64(id)}
	api, err := d.Prods()
	if err != nil {
		return fmt.Errorf("get releaser prods id %d: %w", id, err)
	}
	r.Code = d.StatusCode
	r.Status = d.Status
	r.API = api
	return nil
}

// MsDosProducts are productions that match platforms id 4, MS-DOS.
// Productions with the tag "lost" are skipped.
// Productions created on or newer than 1 Jan. 2000 are skipped.
type MsDosProducts struct {
	Code   int
	Status string
	API    []releases.ProductionV1
	Count  int
	Finds  int
}

func (m *MsDosProducts) Get() error {
	d := filter.Productions{Filter: releases.MsDos}
	api, err := d.Prods(false)
	if err != nil {
		return fmt.Errorf("get msdos prods: %w", err)
	}
	m.Code = d.StatusCode
	m.Status = d.Status
	m.Count = d.Count
	m.Finds = d.Finds
	m.API = api
	return nil
}

// Fix repairs imported Demozoo data conflicts.
func Fix() error {
	return fix.Configs()
}

// NewRecord initialises a new file record.
func NewRecord(c int, values []sql.RawBytes) (Record, error) {
	const sep, want = ",", 21
	if l := len(values); l < want {
		return Record{}, fmt.Errorf("new records = %d, want %d: %w", l, want, ErrTooFew)
	}
	const id, uuid, createdat, filename, filesize, webiddemozoo = 0, 1, 3, 4, 5, 6
	const filezipcontent, updatedat, platform, fileintegritystrong, fileintegrityweak = 7, 8, 9, 10, 11
	const webidpouet, groupbrandfor, groupbrandby, recordtitle, section = 12, 13, 14, 15, 16
	const creditillustration, creditaudio, creditprogram, credittext = 17, 18, 19, 20
	r := Record{
		Count: c,
		ID:    string(values[id]),
		UUID:  string(values[uuid]),
		// deletedat placeholder
		CreatedAt: database.DateTime(values[createdat]),
		Filename:  string(values[filename]),
		Filesize:  string(values[filesize]),
		// web_id_demozoo placeholder
		FileZipContent: string(values[filezipcontent]),
		UpdatedAt:      database.DateTime(values[updatedat]),
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

// RefreshMeta synchronises missing file entries with Demozoo sourced metadata.
func RefreshMeta() error {
	return refresh(meta)
}

// RefreshPouet synchronises missing file entries with Demozoo sourced metadata.
func RefreshPouet() error {
	return refresh(pouet)
}

func refresh(r request) error {
	start := time.Now()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(selectByID(""))
	if err != nil {
		return fmt.Errorf("meta query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("meta rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("meta columns: %w", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// fetch the rows
	var st Stat
	switch r {
	case meta:
		for rows.Next() {
			if err := st.NextRefresh(Records{rows, scanArgs, values}); err != nil {
				logs.Println(fmt.Errorf("meta rows: %w", err))
			}
		}
	case pouet:
		for rows.Next() {
			if err := st.NextPouet(Records{rows, scanArgs, values}); err != nil {
				logs.Println(fmt.Errorf("meta rows: %w", err))
			}
		}
	}
	st.summary(time.Since(start))
	return nil
}
