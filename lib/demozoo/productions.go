package demozoo

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
)

// Authors contains Defacto2 people rolls.
type Authors struct {
	text  []string // credit_text, writer
	code  []string // credit_program, programmer/coder
	art   []string // credit_illustration, artist/graphics
	audio []string // credit_audio, musician/sound
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

// Authors parses Demozoo authors and reclassifies them into Defacto2 people rolls.
func (p *ProductionsAPIv1) Authors() Authors {
	var a Authors
	for _, n := range p.Credits {
		if n.Nick.Releaser.IsGroup {
			continue
		}
		switch category(n.Category) {
		case Text:
			a.text = append(a.text, n.Nick.Name)
		case Code:
			a.code = append(a.code, n.Nick.Name)
		case Graphics:
			a.art = append(a.art, n.Nick.Name)
		case Music:
			a.audio = append(a.audio, n.Nick.Name)
		case Magazine:
			// do nothing.
		}
	}
	return a
}

// DownloadLink parses the Demozoo DownloadLinks to return the filename and link of the first suitable download.
func (p *ProductionsAPIv1) DownloadLink() (name, link string) {
	const (
		found       = 200
		internalErr = 500
	)
	total := len(p.DownloadLinks)
	for _, l := range p.DownloadLinks {
		var l DownloadsAPIv1 = l // apply type so we can use it with methods
		if ok := l.parse(); !ok {
			continue
		}
		// skip defacto2 links if others are available
		if u, err := url.Parse(l.URL); total > 1 && u.Hostname() == df2 {
			if flag.Lookup("test.v") != nil {
				log.Printf("url.Parse(%s) error = %q\n", l.URL, err)
			}
			continue
		}
		ping, err := download.LinkPing(l.URL)
		if err != nil || ping.StatusCode != found {
			if flag.Lookup("test.v") != nil {
				if err != nil {
					log.Printf("download.LinkPing(%s) error = %q\n", l.URL, err)
				} else {
					log.Printf("download.LinkPing(%s) %v != %v\n", l.URL, ping.StatusCode, found)
				}
			}
			continue
		}
		defer ping.Body.Close()
		name = filename(ping.Header)
		if name == "" {
			name, err = saveName(l.URL)
			if err != nil {
				continue
			}
		}
		link = l.URL
		break
	}
	return name, link
}

func (p *ProductionsAPIv1) Download(l DownloadsAPIv1) error {
	const found = 200
	if ok := l.parse(); !ok {
		logs.Print(" not usable\n")
		return nil
	}
	ping, err := download.LinkPing(l.URL)
	if err != nil {
		return fmt.Errorf("download off demozoo ping: %w", err)
	}
	defer ping.Body.Close()
	if ping.StatusCode != found {
		logs.Printf(" %s", ping.Status) // print the HTTP status
		return nil
	}
	save, err := saveName(l.URL)
	if err != nil {
		return fmt.Errorf("download off demozoo: %w", err)
	}
	temp, err := ioutil.TempDir("", "demozoo-download")
	if err != nil {
		return fmt.Errorf("download off demozoo temp dir: %w", err)
	}
	dest, err := filepath.Abs(filepath.Join(temp, save))
	if err != nil {
		return fmt.Errorf("download off demozoo abs filepath: %w", err)
	}
	_, err = download.LinkDownload(dest, l.URL)
	if err != nil {
		return fmt.Errorf("download off demozoo download: %w", err)
	}
	return nil
}

// Downloads parses the Demozoo DownloadLinks and saves the first suitable download.
func (p *ProductionsAPIv1) Downloads() {
	for _, l := range p.DownloadLinks {
		if err := p.Download(l); err != nil {
			log.Printf(" %s", err)
		} else {
			break
		}
	}
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
		id, err := parsePouetProduction(l.URL)
		if err != nil {
			return 0, 0, fmt.Errorf("pouet id parse: %w", err)
		}
		if ping {
			resp, err := download.LinkPing(l.URL)
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

// Print displays the production API results as tabbed JSON.
func (p *ProductionsAPIv1) Print() error {
	js, err := json.MarshalIndent(&p, "", "  ")
	if err != nil {
		return fmt.Errorf("print json marshal indent: %w", err)
	}
	// ignore --quiet
	fmt.Println(string(js))
	return nil
}

// DownloadsAPIv1 are DownloadLinks for ProductionsAPIv1.
type DownloadsAPIv1 struct {
	LinkClass string `json:"link_class"`
	URL       string `json:"url"`
}

// parse corrects any known errors with a Downloads API link.
func (dl *DownloadsAPIv1) parse() (ok bool) {
	u, err := url.Parse(dl.URL) // validate url
	if err != nil {
		return false
	}
	u = mutateURL(u)
	dl.URL = u.String()
	return true
}

func filename(h http.Header) (filename string) {
	gh := h.Get(cd)
	if gh == "" {
		return filename
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
	return filename
}

// mutateURL applies fixes to known problematic URLs.
func mutateURL(u *url.URL) *url.URL {
	if u == nil {
		s, err := url.Parse("")
		if err != nil {
			log.Fatal(fmt.Errorf("mutate url parse: %w", err))
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

// parsePouetProduction takes a pouet prod URL and extracts the ID.
func parsePouetProduction(rawurl string) (id int, err error) {
	u, err := url.Parse(rawurl)
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

// randomName generates a random temporary filename.
func randomName() (name string, err error) {
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

// saveName gets a filename from the URL or generates a random filename.
func saveName(rawurl string) (name string, err error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", fmt.Errorf("save name %q: %w", rawurl, err)
	}
	name = filepath.Base(u.Path)
	if name == "." {
		name, err = randomName()
		if err != nil {
			return "", fmt.Errorf("save name random: %w", err)
		}
	}
	return name, nil
}
