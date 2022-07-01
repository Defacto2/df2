package prods

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/logs"
)

const (
	cd  = "Content-Disposition"
	cls = "PouetProduction"
	df2 = "defacto2.net"
)

// ProductionsAPIv1 productions API v1.
// This can be dynamically generated at https://mholt.github.io/json-to-go/
// Get the Demozoo JSON output from https://demozoo.org/api/v1/productions/{{.ID}}/?format=json
type ProductionsAPIv1 struct {
	URL         string `json:"url"`
	DemozooURL  string `json:"demozoo_url"`
	ID          int    `json:"id"`
	Title       string `json:"title"`
	AuthorNicks []struct {
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Releaser     struct {
			URL     string `json:"url"`
			ID      int    `json:"id"`
			Name    string `json:"name"`
			IsGroup bool   `json:"is_group"`
		} `json:"releaser"`
	} `json:"author_nicks"`
	AuthorAffiliationNicks []any  `json:"author_affiliation_nicks"`
	ReleaseDate            string `json:"release_date"`
	Supertype              string `json:"supertype"`
	Platforms              []struct {
		URL  string `json:"url"`
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"platforms"`
	Types []struct {
		URL       string `json:"url"`
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Supertype string `json:"supertype"`
	} `json:"types"`
	Credits []struct {
		Nick struct {
			Name         string `json:"name"`
			Abbreviation string `json:"abbreviation"`
			Releaser     struct {
				URL     string `json:"url"`
				ID      int    `json:"id"`
				Name    string `json:"name"`
				IsGroup bool   `json:"is_group"`
			} `json:"releaser"`
		} `json:"nick"`
		Category string `json:"category"`
		Role     string `json:"role"`
	} `json:"credits"`
	DownloadLinks []struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	} `json:"download_links"`
	ExternalLinks []struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	} `json:"external_links"`
	ReleaseParties      []any `json:"release_parties"`
	CompetitionPlacings []any `json:"competition_placings"`
	InvitationParties   []any `json:"invitation_parties"`
	Screenshots         []struct {
		OriginalURL     string `json:"original_url"`
		OriginalWidth   int    `json:"original_width"`
		OriginalHeight  int    `json:"original_height"`
		StandardURL     string `json:"standard_url"`
		StandardWidth   int    `json:"standard_width"`
		StandardHeight  int    `json:"standard_height"`
		ThumbnailURL    string `json:"thumbnail_url"`
		ThumbnailWidth  int    `json:"thumbnail_width"`
		ThumbnailHeight int    `json:"thumbnail_height"`
	} `json:"screenshots"`
}

// JSON returns the production API results as tabbed JSON.
// This is used by internal/generator.go.
func (p *ProductionsAPIv1) JSON() ([]byte, error) {
	js, err := json.MarshalIndent(&p, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("json marshal indent: %w", err)
	}
	return js, nil
}

// PouetID returns the ID value used by Pouet's which prod URL syntax
// and its HTTP status code.
// example: https://www.pouet.net/prod.php?which=30352
func (p *ProductionsAPIv1) PouetID(ping bool) (id, statusCode int, err error) {
	for _, l := range p.ExternalLinks {
		if l.LinkClass != cls {
			continue
		}
		id, err := Parse(l.URL)
		if err != nil {
			return 0, 0, fmt.Errorf("pouet id parse: %w", err)
		}
		if ping {
			resp, err := download.Ping(l.URL)
			if err != nil {
				return 0, 0, fmt.Errorf("pouet id ping: %w", err)
			}
			resp.Body.Close()
			return id, resp.StatusCode, nil
		}
		return id, 0, nil
	}
	return 0, 0, nil
}

// Print to stdout the production API results as tabbed JSON.
func (p *ProductionsAPIv1) Print() error {
	js, err := json.MarshalIndent(&p, "", "  ")
	if err != nil {
		return fmt.Errorf("print json marshal indent: %w", err)
	}
	fmt.Println(string(js))
	return nil
}

// Filename is obtained from the http header metadata.
func Filename(h http.Header) string {
	gh := h.Get(cd)
	if gh == "" {
		return ""
	}
	rh := strings.Split(gh, ";")
	const want = 2
	for _, v := range rh {
		r := strings.Split(v, "=")
		r[0] = strings.TrimSpace(r[0])
		if len(r) != want {
			continue
		}
		switch r[0] {
		case "filename*", "filename":
			return r[1]
		}
	}
	return ""
}

// Mutate applies fixes to known problematic URLs.
func Mutate(u *url.URL) *url.URL {
	if u == nil {
		s, err := url.Parse("")
		if err != nil {
			log.Print(fmt.Errorf("mutate url parse: %w", err))
		}
		return s
	}
	switch u.Hostname() {
	case "files.scene.org":
		p := strings.Split(u.Path, "/")
		// https://files.scene.org/view/.. >
		// https://files.scene.org/get:nl-http/..
		if p[1] == "view" {
			p[1] = "get:nl-http" // must include -http to avoid FTP links
			u.Path = strings.Join(p, "/")
		}
	default:
	}
	return u
}

// Parse takes a Pouet prod URL and extracts the ID.
func Parse(rawURL string) (id int, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, fmt.Errorf(" url parse: %w", err)
	}
	q := u.Query()
	w := q.Get("which")
	if w == "" {
		return 0, fmt.Errorf("parse pouet production &which=%v: %w", w, err)
	}
	id, err = strconv.Atoi(w)
	if err != nil {
		return 0, fmt.Errorf("parse pouet production &which=%v: %w", w, err)
	}
	if id < 0 {
		return 0, fmt.Errorf("parse pouet production &which=%v: %w", id, err)
	}
	return id, nil
}

// RandomName generates a random temporary filename.
func RandomName() (string, error) {
	tmp, err := ioutil.TempFile("", "df2-download")
	if err != nil {
		return "", fmt.Errorf("random name tempfile: %w", err)
	}
	defer tmp.Close()
	if err := os.Remove(tmp.Name()); err != nil {
		logs.Log(fmt.Errorf("random name remove tempfile %q: %w", tmp.Name(), err))
	}
	return tmp.Name(), nil
}

// SaveName gets a filename from the URL or generates a random filename.
func SaveName(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("save name %q: %w", rawURL, err)
	}
	name := filepath.Base(u.Path)
	if name == "." {
		name, err = RandomName()
		if err != nil {
			return "", fmt.Errorf("save name random: %w", err)
		}
	}
	return name, nil
}
