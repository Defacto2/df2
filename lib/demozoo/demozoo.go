package demozoo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
)

const (
	prodAPI string = "https://demozoo.org/api/v1/productions"
)

// DownloadsAPIv1 are DownloadLinks for ProductionsAPIv1.
type DownloadsAPIv1 struct {
	LinkClass string `json:"link_class"`
	URL       string `json:"url"`
}

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
	AuthorAffiliationNicks []interface{} `json:"author_affiliation_nicks"`
	ReleaseDate            string        `json:"release_date"`
	Supertype              string        `json:"supertype"`
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
	ReleaseParties      []interface{} `json:"release_parties"`
	CompetitionPlacings []interface{} `json:"competition_placings"`
	InvitationParties   []interface{} `json:"invitation_parties"`
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

// Production API production request.
type Production struct {
	ID      int64         // Demozoo production ID
	Timeout time.Duration // HTTP request timeout in seconds (default 5)
	link    string        // URL link to sent the request // ??
}

// Fetch a Demozoo production by its ID.
func Fetch(id uint) ProductionsAPIv1 {
	var d = Production{
		ID: int64(id),
	}
	api := *d.data()
	return api
}

// Downloads parses the Demozoo DownloadLinks can saves the first suitable download.
func (p *ProductionsAPIv1) Downloads() {
	for _, l := range p.DownloadLinks {
		var l DownloadsAPIv1 = l // apply type so we can use it with methods
		if ok := l.parse(); !ok {
			logs.Print(" not usable\n")
			continue
		}
		c, s, err := download.LinkPing(l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		if c != 200 {
			logs.Printf(" %s", s) // print the HTTP status
			continue
		}
		logs.Print("\n")
		save, err := saveName(l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		saveDest, err := filepath.Abs(filepath.Join("/home", "ben", save)) // TODO PATH arg instead of hardcoded
		if err != nil {
			logs.Log(err)
			continue
		}
		err = download.LinkDownload(saveDest, l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		break
	}
}

// saveName gets a filename from the URL or generates a random filename.
func saveName(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	base := filepath.Base(u.Path)
	if base == "." {
		base = randomName()
	}
	return base, nil
}

// randomName generates a random temporary filename.
func randomName() string {
	tmpfile, err := ioutil.TempFile("", "df2-download")
	logs.Check(err)
	defer os.Remove(tmpfile.Name())
	name := tmpfile.Name()
	err = tmpfile.Close()
	logs.Check(err)
	return name
}

// parse corrects any known errors with a Downloads API link.
func (dl *DownloadsAPIv1) parse() bool {
	logs.Printf("Discovered %s link: %s", dl.LinkClass, dl.URL)
	u, err := url.Parse(dl.URL) // validate url
	if err != nil {
		return false
	}
	u = mutateURL(u)
	dl.URL = u.String()
	return true
}

// mutateURL applies fixes to known problematic URLs.
func mutateURL(u *url.URL) *url.URL {
	if u == nil {
		u, err := url.Parse("")
		logs.Check(err)
		return u
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
	}
	return u
}

// PouetID returns the ID value used by Pouet's which prod URL syntax
// and its HTTP status code.
// example: https://www.pouet.net/prod.php?which=30352
func (p *ProductionsAPIv1) PouetID() (int, int) {
	for _, l := range p.ExternalLinks {
		if l.LinkClass != "PouetProduction" {
			continue
		}
		id, err := parsePouetProduction(l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		code, _, _ := download.LinkPing(l.URL)
		return id, code
	}
	return 0, 0
}

// parsePouetProduction takes a pouet prod URL and extracts the ID
func parsePouetProduction(rawurl string) (int, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return 0, err
	}
	q := u.Query()
	w := q.Get("which")
	if w == "" {
		return 0, fmt.Errorf("unexpected PouetProduction url syntax: %s", rawurl)
	}
	id, err := strconv.Atoi(w)
	if err != nil {
		return 0, fmt.Errorf("unexpected PouetProduction which= query syntax: %s", w)
	}
	if id < 0 {
		return 0, fmt.Errorf("unexpected PouetProduction which= query value: %s", w)
	}
	return id, nil
}

// Print displays the production API results as tabbed JSON.
func (p *ProductionsAPIv1) Print() {
	js, err := json.MarshalIndent(&p, "", "  ")
	logs.Check(err)
	logs.Println(string(js))
}

// URL creates a productions API v.1 request link.
// example: https://demozoo.org/api/v1/productions/158411/?format=json
func (p *Production) URL() {
	rawurl, err := prodURL(p.ID)
	logs.Check(err)
	p.link = rawurl
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

// data gets a production API link and normalises the results.
func (p *Production) data() *ProductionsAPIv1 {
	p.URL()
	var r = download.Request{
		Link: p.link,
	}
	body := *r.Body()

	dz := ProductionsAPIv1{}
	err := json.Unmarshal(body, &dz)
	logs.Check(err)
	return &dz
}
