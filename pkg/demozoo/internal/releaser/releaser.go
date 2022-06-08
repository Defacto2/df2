package releaser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
	"github.com/Defacto2/df2/pkg/download"
)

var ErrNegativeID = errors.New("demozoo production id cannot be a negative integer")

const api = "https://demozoo.org/api/v1/releasers"

// Releaser API production request.
type Releaser struct {
	ID         int64         // Demozoo releaser ID
	Timeout    time.Duration // HTTP request timeout in seconds (default 5)
	Link       string        // URL link to send the request
	StatusCode int           // received HTTP statuscode
	Status     string        // received HTTP status
}

// ReleaserV1 releasers API v1.
// This can be dynamically generated at https://mholt.github.io/json-to-go/
// Get the Demozoo JSON output from https://demozoo.org/api/v1/releasers/{{.ID}}/?format=json
type ReleaserV1 struct {
	URL        string `json:"url"`
	DemozooURL string `json:"demozoo_url"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
	IsGroup    bool   `json:"is_group"`
	Nicks      []struct {
		Name          string   `json:"name"`
		Abbreviation  string   `json:"abbreviation"`
		IsPrimaryNick bool     `json:"is_primary_nick"`
		Variants      []string `json:"variants"`
	} `json:"nicks"`
	MemberOf []interface{} `json:"member_of"`
	Members  []struct {
		Member struct {
			URL  string `json:"url"`
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"member"`
		IsCurrent bool `json:"is_current"`
	} `json:"members"`
	Subgroups     []interface{} `json:"subgroups"`
	ExternalLinks []struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	} `json:"external_links"`
}

// Print to stdout the releaser API results as tabbed JSON.
func (r *ReleaserV1) Print() error {
	js, err := json.MarshalIndent(&r, "", "  ")
	if err != nil {
		return fmt.Errorf("print json marshal indent: %w", err)
	}
	// ignore --quiet
	fmt.Println(string(js))
	return nil
}

// URL creates a releasers API v1 request link.
// example: https://demozoo.org/api/v1/releasers/10000/?format=json
func (r *Releaser) URL() error {
	s, err := URL(r.ID)
	if err != nil {
		return fmt.Errorf("releaser url: %w", err)
	}
	r.Link = s
	return nil
}

// Get a releaser API link and normalises the results.
func (r *Releaser) Get() (ReleaserV1, error) {
	if err := r.URL(); err != nil {
		return ReleaserV1{}, fmt.Errorf("releaser data: %w", err)
	}
	req := download.Request{
		Link: r.Link,
	}
	if err := req.Body(); err != nil {
		return ReleaserV1{}, fmt.Errorf("releaser data body: %w", err)
	}
	r.Status = req.Status
	r.StatusCode = req.StatusCode
	var rel ReleaserV1
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &rel); err != nil {
			return ReleaserV1{}, fmt.Errorf("releaser data json unmarshal: %w", err)
		}
	}
	return rel, nil
}

// Prods gets all the productions of a releaser and normalises the results.
func (r *Releaser) Prods() (releases.Productions, error) {
	url, err := releases.URL(int64(r.ID))
	if err != nil {
		return releases.Productions{}, err
	}
	req := download.Request{
		Link: url,
	}
	if err := req.Body(); err != nil {
		return releases.Productions{}, fmt.Errorf("releaser data body: %w", err)
	}
	r.Status = req.Status
	r.StatusCode = req.StatusCode
	var dz releases.Productions
	if len(req.Read) > 0 {
		if err := json.Unmarshal(req.Read, &dz); err != nil {
			return releases.Productions{}, fmt.Errorf("releaser data json unmarshal: %w", err)
		}
	}
	return dz, nil
}

// URL creates a releaser URL from a Demozoo ID.
func URL(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("releaser id %v: %w", id, ErrNegativeID)
	}
	u, err := url.Parse(api) // base URL
	if err != nil {
		return "", fmt.Errorf("releaser parse: %w", err)
	}
	const decimal = 10
	u.Path = path.Join(u.Path, strconv.FormatInt(id, decimal)) // append ID
	q := u.Query()
	q.Set("format", "json") // append format=json
	u.RawQuery = q.Encode()
	return u.String(), nil
}
