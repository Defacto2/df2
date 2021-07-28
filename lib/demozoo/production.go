package demozoo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/Defacto2/df2/lib/download"
)

// Production API production request.
type Production struct {
	ID         int64         // Demozoo production ID
	Timeout    time.Duration // HTTP request timeout in seconds (default 5)
	link       string        // URL link to sent the request // ??
	StatusCode int           // received HTTP statuscode
	Status     string        // received HTTP status
}

// URL creates a productions API v.1 request link.
// example: https://demozoo.org/api/v1/productions/158411/?format=json
func (p *Production) URL() error {
	s, err := production(p.ID)
	if err != nil {
		return fmt.Errorf("production url: %w", err)
	}
	p.link = s
	return nil
}

// data gets a production API link and normalises the results.
func (p *Production) data() (ProductionsAPIv1, error) {
	if err := p.URL(); err != nil {
		return ProductionsAPIv1{}, fmt.Errorf("production data: %w", err)
	}
	r := download.Request{
		Link: p.link,
	}
	if err := r.Body(); err != nil {
		return ProductionsAPIv1{}, fmt.Errorf("production data body: %w", err)
	}
	p.Status = r.Status
	p.StatusCode = r.StatusCode
	var dz ProductionsAPIv1
	if len(r.Read) > 0 {
		if err := json.Unmarshal(r.Read, &dz); err != nil {
			return ProductionsAPIv1{}, fmt.Errorf("production data json unmarshal: %w", err)
		}
	}
	return dz, nil
}

// production creates a production URL from a Demozoo ID.
func production(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("production id %v: %w", id, ErrNegativeID)
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
