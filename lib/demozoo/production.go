package demozoo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
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
func (p *Production) URL() {
	rawurl, err := prodURL(p.ID)
	logs.Check(err)
	p.link = rawurl
}

// data gets a production API link and normalises the results.
func (p *Production) data() *ProductionsAPIv1 {
	var err error
	p.URL()
	var r = download.Request{
		Link: p.link,
	}
	err = r.Body()
	logs.Log(err)
	p.Status = r.Status
	p.StatusCode = r.StatusCode
	dz := ProductionsAPIv1{}
	if len(r.Read) > 0 {
		err = json.Unmarshal(r.Read, &dz)
	}
	if err != nil && logs.Panic {
		logs.Println(string(r.Read))
	}
	logs.Check(err)
	return &dz
}

// prodURL creates a production URL from a Demozoo ID.
func prodURL(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("unexpected negative id value %v", id)
	}
	u, err := url.Parse(prodAPI) // base URL
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, strconv.FormatInt(id, 10)) // append ID
	q := u.Query()
	q.Set("format", "json") // append format=json
	u.RawQuery = q.Encode()
	return u.String(), nil
}
