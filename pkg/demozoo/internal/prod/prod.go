// Package prod obtains a Demozoo Production.
package prod

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/Defacto2/df2/pkg/download"
)

var ErrID = errors.New("demozoo production id cannot be a negative integer")

const api = "https://demozoo.org/api/v1/productions"

// Production API production request.
type Production struct {
	ID      int64         // Demozoo production ID.
	Link    string        // Link URL to receive the request.
	Code    int           // Code is the HTTP status.
	Status  string        // Status is the HTTP status.
	Timeout time.Duration // Timeout in seconds for the HTTP request (default 5).
}

// URL creates a productions API v.1 request link.
// example: https://demozoo.org/api/v1/productions/158411/?format=json
func (p *Production) URL() error {
	s, err := URL(p.ID)
	if err != nil {
		return fmt.Errorf("production url: %w", err)
	}
	p.Link = s
	return nil
}

// Get a production API link and normalises the results.
func (p *Production) Get() (prods.ProductionsAPIv1, error) {
	if err := p.URL(); err != nil {
		return prods.ProductionsAPIv1{}, fmt.Errorf("production data: %w", err)
	}
	r := download.Request{
		Link: p.Link,
	}
	if err := r.Body(); err != nil {
		return prods.ProductionsAPIv1{}, fmt.Errorf("production data body: %w", err)
	}
	p.Status = r.Status
	p.Code = r.Code
	if len(r.Read) == 0 {
		return prods.ProductionsAPIv1{}, nil
	}
	dz := prods.ProductionsAPIv1{}
	if err := json.Unmarshal(r.Read, &dz); err != nil {
		return prods.ProductionsAPIv1{}, fmt.Errorf("production data json unmarshal: %w", err)
	}
	return dz, nil
}

// URL creates a production URL from a Demozoo ID.
func URL(id int64) (string, error) {
	if id < 1 {
		return "", fmt.Errorf("production id %v: %w", id, ErrID)
	}
	u, err := url.Parse(api) // base URL
	if err != nil {
		return "", fmt.Errorf("production parse: %w", err)
	}
	const decimal = 10
	u.Path = path.Join(u.Path, strconv.FormatInt(id, decimal)) // append ID
	q := u.Query()
	q.Set("format", "json") // append format=json
	u.RawQuery = q.Encode()
	return u.String(), nil
}
