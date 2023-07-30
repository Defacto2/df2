// Package filter confirms a Demozoo Production is suitable for Defacto2.
// A MS-DOS intro would be okay, while a demo for the Atari 2600 would be rejected.
package filter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/insert"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prod"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/str"
)

const (
	domain   = "defacto2.net"
	maxTries = 5
)

// Productions API production request.
type Productions struct {
	Filter  releases.Filter
	Count   int           // Count the total productions.
	Finds   int           // Finds are the number of productions to use.
	Link    string        // Link URL to receive the request.
	Code    int           // Code received by the HTTP request.
	Status  string        // Status received by the HTTP request.
	Timeout time.Duration // Timeout in seconds for the HTTP request (default 5).
}

// ProductionList result.
type ProductionList struct {
	Count    int                     `json:"count"`    // Count the total productions.
	Next     string                  `json:"next"`     // Next page URL.
	Previous interface{}             `json:"previous"` // Previous page object.
	Results  []releases.ProductionV1 `json:"results"`  // Results of the productions.
}

func empty() []releases.ProductionV1 {
	return []releases.ProductionV1{}
}

// Prods gets all the productions of a releaser and normalises the results.
// The maxPage are the maximum number of API pages to iterate, but if set to 0 it will default to 1000.
func (p *Productions) Prods(db *sql.DB, w io.Writer, maxPage int) ([]releases.ProductionV1, error) { //nolint:funlen
	if db == nil {
		return nil, database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	const endOfRecords = ""
	dz := ProductionList{}
	finds := 0

	link, err := p.Filter.URL(0)
	if err != nil {
		return empty(), err
	}
	req := download.Request{Link: link}
	fmt.Fprintf(w, "Fetching the first 100 of many records from Demozoo\n")
	tries := 0
	for {
		tries++
		if err := req.Body(); err != nil {
			if tries <= maxTries {
				continue
			}
			return empty(), fmt.Errorf("filter data body: %w", err)
		}
		break
	}
	p.Status = req.Status
	p.Code = req.Code
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &dz); err != nil {
			return empty(), fmt.Errorf("filter data json unmarshal: %w", err)
		}
	}
	p.Count = dz.Count
	fmt.Fprintf(w, "There are %d %s production matches\n", dz.Count, p.Filter)
	rels, err := Filter(db, w, dz.Results)
	if err != nil {
		return nil, err
	}
	f := len(rels)
	wp(w, f, 1)
	np, page := dz.Next, 1
	if maxPage < 1 {
		maxPage = 1000
	}
	for {
		page++
		rel, next, err := Next(np)
		if err != nil {
			return empty(), err
		}
		rel, err = Filter(db, w, rel)
		if err != nil {
			return nil, err
		}
		f = len(rel)
		finds += f
		wp(w, f, page)
		rels = append(rels, rel...)
		if next == endOfRecords {
			break
		}
		if page > maxPage {
			break
		}
	}
	p.Finds = finds
	return rels, nil
}

func wp(w io.Writer, finds, page int) {
	if finds == 0 {
		fmt.Fprintf(w, "   Page %d, no new records found\n", page)
		return
	}
	if finds == 1 {
		fmt.Fprintf(w, "   Page %d, new record found, 1\n", page)
		return
	}
	fmt.Fprintf(w, "   Page %d, new records found, %d\n", page, finds)
}

// Next gets all the next page of productions.
func Next(link string) ([]releases.ProductionV1, string, error) {
	req := download.Request{Link: link}
	tries := 0
	for {
		tries++
		if err := req.Body(); err != nil {
			if tries <= maxTries {
				continue
			}
			return empty(), "", fmt.Errorf("filter data body: %w", err)
		}
		break
	}
	dz := ProductionList{}
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &dz); err != nil {
			return empty(), "", fmt.Errorf("filter data json unmarshal: %w", err)
		}
	}
	return dz.Results, dz.Next, nil
}

// Filter removes any productions that are not suitable for Defacto2.
func Filter(db *sql.DB, w io.Writer, prods []releases.ProductionV1) ([]releases.ProductionV1, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	finds := 0
	p := []releases.ProductionV1{}
	for _, prod := range prods {
		if !prodType(prod.Types) {
			continue
		}
		if invalid(prod.Tags) {
			continue
		}
		// confirm ID is not already used in a defacto2 file record
		if id, _ := database.DemozooID(db, uint(prod.ID)); id > 0 {
			continue
		}
		if l, _ := linked(prod.ID); l != "" {
			if err := sync(db, w, prod.ID, database.DeObfuscate(l)); err != nil {
				fmt.Fprintln(w, err)
			}
			continue
		}
		rec := make(releases.Productions, 1)
		rec[0] = prod
		if err := insert.Prods(db, w, &rec); err != nil {
			fmt.Fprintln(w, err)
			continue
		}
		finds++
		fmt.Fprintf(w, "%s%d. (%d) %s\n", str.PrePad, finds, prod.ID, prod.Title)
		p = append(p, prod)
	}
	return p, nil
}

func sync(db *sql.DB, w io.Writer, demozooID, recordID int) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	i, err := update(db, demozooID, recordID)
	if err != nil {
		fmt.Fprintf(w, " Found an unlinked Demozoo record %d, that points to Defacto2 ID %d\n",
			demozooID, recordID)
		return err
	}
	fmt.Fprintf(w, " Updated %d record, the Demozoo ID %d was saved to record id: %v\n", i,
		demozooID, recordID)
	return nil
}

func update(db *sql.DB, demozooID, recordID int) (int64, error) {
	up := database.Update{}
	up.Query = "UPDATE files SET web_id_demozoo=? WHERE `id` = ?"
	up.Args = []any{demozooID, recordID}
	count, err := database.Execute(db, up)
	if err != nil {
		return 0, fmt.Errorf("update installers: %w", err)
	}
	return count, nil
}

// linked returns the Defacto2 URL linked to a Demozoo ID that points to a download or external link.
func linked(id int) (string, error) {
	p := prod.Production{}
	p.ID = int64(id)
	api, err := p.Get()
	if err != nil {
		return "", err
	}
	for _, link := range api.DownloadLinks {
		if strings.Contains(link.URL, domain) {
			return link.URL, nil
		}
	}
	for _, link := range api.ExternalLinks {
		if strings.Contains(link.URL, domain) {
			return link.URL, nil
		}
	}
	return "", nil
}

func prodType(types []releases.Type) bool {
	const (
		bbstro   = 41
		cracktro = 13
	)
	for _, t := range types {
		if t.ID == bbstro {
			return true
		}
		if t.ID == cracktro {
			return true
		}
	}
	return false
}

func invalid(tags []string) bool {
	for _, tag := range tags {
		if strings.ToLower(tag) == "lost" {
			return true
		}
		if strings.ToLower(tag) == "no-binaries" {
			return true
		}
	}
	return false
}
