package demozoo

import (
	"encoding/json"
	"fmt"
	"log"
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
	rawurl, err := prodURL(p.ID)
	if err != nil {
		return fmt.Errorf("production url: %w", err)
	}
	p.link = rawurl
	return nil
}

// data gets a production API link and normalises the results.
func (p *Production) data() *ProductionsAPIv1 {
	if err := p.URL(); err != nil {
		log.Fatal(fmt.Errorf("production data: %w", err))
	}
	r := download.Request{
		Link: p.link,
	}
	if err := r.Body(); err != nil {
		log.Fatal(fmt.Errorf("production data body: %w", err))
	}
	p.Status = r.Status
	p.StatusCode = r.StatusCode
	dz := ProductionsAPIv1{}
	if len(r.Read) > 0 {
		if err := json.Unmarshal(r.Read, &dz); err != nil {
			log.Fatal(fmt.Errorf("production data json unmarshal: %w", err))
		}
	}
	return &dz
}

// prodURL creates a production URL from a Demozoo ID.
func prodURL(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("prod url id %v: %w", id, ErrNegativeID)
	}
	u, err := url.Parse(prodAPI) // base URL
	if err != nil {
		return "", fmt.Errorf("prod url parse: %w", err)
	}
	u.Path = path.Join(u.Path, strconv.FormatInt(id, 10)) // append ID
	q := u.Query()
	q.Set("format", "json") // append format=json
	u.RawQuery = q.Encode()
	return u.String(), nil
}
