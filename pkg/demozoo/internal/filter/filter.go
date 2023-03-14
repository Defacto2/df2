package filter

import (
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
)

const maxTries = 5

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
func (p *Productions) Prods(w io.Writer) ([]releases.ProductionV1, error) { //nolint:funlen
	const endOfRecords, maxPage = "", 1000
	var next []releases.ProductionV1
	var dz ProductionList
	totalFinds := 0

	url, err := p.Filter.URL(0)
	if err != nil {
		return empty(), err
	}
	req := download.Request{Link: url}
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
	p.StatusCode = req.StatusCode
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &dz); err != nil {
			return empty(), fmt.Errorf("filter data json unmarshal: %w", err)
		}
	}
	p.Count = dz.Count
	fmt.Fprintf(w, "There are %d %s production matches\n", dz.Count, p.Filter)
	finds, prods := Filter(w, dz.Results)
	pp(w, 1, finds)
	nu, page := dz.Next, 1
	for {
		page++
		next, nu, err = Next(nu)
		if err != nil {
			return empty(), err
		}
		finds, next = Filter(w, next)
		totalFinds += finds
		pp(w, page, finds)
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

func pp(w io.Writer, page, finds int) {
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
func Next(url string) ([]releases.ProductionV1, string, error) {
	req := download.Request{Link: url}
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
	var dz ProductionList
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &dz); err != nil {
			return empty(), "", fmt.Errorf("filter data json unmarshal: %w", err)
		}
	}
	return dz.Results, dz.Next, nil
}

// Filter productions removes any records that are not suitable for Defacto2.
func Filter(w io.Writer, p []releases.ProductionV1) (int, []releases.ProductionV1) {
	finds := 0
	var prods []releases.ProductionV1 //nolint:prealloc
	for _, prod := range p {
		if !prodType(prod.Types) {
			continue
		}
		if invalid(prod.Tags) {
			continue
		}
		// confirm ID is not already used in a defacto2 file record
		if id, _ := database.DemozooID(w, uint(prod.ID)); id > 0 {
			continue
		}
		if l, _ := linked(prod.ID); l != "" {
			sync(w, prod.ID, database.DeObfuscate(l))
			continue
		}
		rec := make(releases.Productions, 1)
		rec[0] = prod
		if err := insert.Prods(w, &rec); err != nil {
			fmt.Fprintln(w, err)
			continue
		}
		fmt.Fprintf(w, "%d. (%d) %s\n", finds+1, prod.ID, prod.Title)
		finds++
		prods = append(prods, prod)
	}
	return finds, prods
}

func sync(w io.Writer, demozooID, recordID int) error {
	i, err := update(w, demozooID, recordID)
	if err != nil {
		fmt.Fprintf(w, " Found an unlinked Demozoo record %d, that points to Defacto2 ID %d\n",
			demozooID, recordID)
		return err
	}
	fmt.Fprintf(w, " Updated %d record, the Demozoo ID %d was saved to record id: %v\n", i,
		demozooID, recordID)
	return nil
}

func update(w io.Writer, demozooID, recordID int) (int64, error) {
	var up database.Update
	up.Query = "UPDATE files SET web_id_demozoo=? WHERE `id` = ?"
	up.Args = []any{demozooID, recordID}
	count, err := database.Execute(w, up)
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
