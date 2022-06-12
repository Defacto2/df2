package filter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/insert"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prod"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/logs"
)

// Productions API production request.
type Productions struct {
	Filter     releases.Filter
	Count      int           // Total productions
	Finds      int           // Found productions to add
	Timeout    time.Duration // HTTP request timeout in seconds (default 5)
	Link       string        // URL link to send the request
	StatusCode int           // received HTTP statuscode
	Status     string        // received HTTP status
}

// Production List result.
type ProductionList struct {
	Count    int                     `json:"count"`    // Total productions
	Next     string                  `json:"next"`     // URL for the next page
	Previous interface{}             `json:"previous"` // URL for the previous page
	Results  []releases.ProductionV1 `json:"results"`
}

func empty() []releases.ProductionV1 {
	return []releases.ProductionV1{}
}

// Prods gets all the productions of a releaser and normalises the results.
func (p *Productions) Prods(quiet bool) ([]releases.ProductionV1, error) {
	const endOfRecords, maxPage = "", 1000
	var next []releases.ProductionV1
	var dz ProductionList
	totalFinds := 0

	url, err := releases.URLFilter(p.Filter)
	if err != nil {
		return empty(), err
	}
	req := download.Request{
		Link: url,
	}
	if !quiet {
		fmt.Printf("Fetch the first 100 of many records from Demozoo\n")
	}
	if err := req.Body(); err != nil {
		return empty(), fmt.Errorf("filter data body: %w", err)
	}
	p.Status = req.Status
	p.StatusCode = req.StatusCode
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &dz); err != nil {
			return empty(), fmt.Errorf("filter data json unmarshal: %w", err)
		}
	}
	p.Count = dz.Count
	finds, prods := Filter(dz.Results, quiet)
	if !quiet {
		pp(1, finds)
	}

	nu, page := dz.Next, 1
	for {
		page++
		next, nu, err = Next(nu)
		if err != nil {
			return empty(), err
		}
		finds, next = Filter(next, quiet)
		totalFinds += finds
		if !quiet {
			pp(page, finds)
		}
		prods = append(prods, next...)
		if nu == endOfRecords {
			break
		}
		if page > maxPage {
			break
		}
	}
	p.Finds = totalFinds
	return prods, nil
}

func pp(page, finds int) {
	if finds == 0 {
		fmt.Printf("   Page %d, no new records found\n", page)
		return
	}
	if finds == 1 {
		fmt.Printf("   Page %d, new record found, 1\n", page)
		return
	}
	fmt.Printf("   Page %d, new records found, %d\n", page, finds)
}

// Next gets all the next page of productions.
func Next(url string) ([]releases.ProductionV1, string, error) {
	req := download.Request{
		Link: url,
	}
	if err := req.Body(); err != nil {
		return empty(), "", fmt.Errorf("filter data body: %w", err)
	}
	var dz ProductionList
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &dz); err != nil {
			return empty(), "", fmt.Errorf("filter data json unmarshal: %w", err)
		}
	}
	return dz.Results, dz.Next, nil
}

// Filter productions removes any records that are not suitable for Defacto2.
func Filter(p []releases.ProductionV1, quiet bool) (int, []releases.ProductionV1) {
	finds := 0
	var prods []releases.ProductionV1 //nolint:prealloc
	for _, prod := range p {
		if !prodType(prod.Types) {
			continue
		}
		if lost(prod.Tags) {
			continue
		}
		// confirm ID is not already used in a defacto2 file record
		if id, _ := database.DemozooID(uint(prod.ID)); id > 0 {
			continue
		}
		if l, _ := linked(prod.ID); l != "" {
			sync(prod.ID, database.DeObfuscate(l), quiet)
			continue
		}
		rec := make(releases.Productions, 1)
		rec[0] = prod
		if err := insert.Prods(&rec, true); err != nil {
			logs.Println(err)
			continue
		}
		fmt.Printf("%d. (%d) %s\n", finds+1, prod.ID, prod.Title)
		finds++
		prods = append(prods, prod)
	}
	return finds, prods
}

func sync(demozooID, recordID int, quiet bool) {
	i, err := update(demozooID, recordID)
	if quiet {
		return
	}
	if err != nil {
		fmt.Printf(" Found an unlinked Demozoo record %d, that points to Defacto2 ID %d\n",
			demozooID, recordID)
		logs.Println(err)
		return
	}
	fmt.Printf(" Updated %d record, the Demozoo ID %d was saved to record id: %v\n", i,
		demozooID, recordID)
}

func update(demozooID, recordID int) (int64, error) {
	var up database.Update
	up.Query = "UPDATE files SET web_id_demozoo=? WHERE `id` = ?"
	up.Args = []any{demozooID, recordID}
	count, err := database.Execute(up)
	if err != nil {
		return 0, fmt.Errorf("update installers: %w", err)
	}
	return count, nil
}

// linked returns the Defacto2 URL linked to a Demozoo ID that points to a download or external link.
func linked(id int) (string, error) {
	const domain = "defacto2.net"
	var p prod.Production
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

func lost(tags []string) bool {
	for _, tag := range tags {
		if strings.ToLower(tag) == "lost" {
			return true
		}
	}
	return false
}
