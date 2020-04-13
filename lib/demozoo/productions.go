package demozoo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
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

// DownloadLink parses the Demozoo DownloadLinks to return the filename and link of the first suitable download.
func (p *ProductionsAPIv1) DownloadLink() (string, string) {
	var save, link string
	var err error
	for _, l := range p.DownloadLinks {
		var l DownloadsAPIv1 = l // apply type so we can use it with methods
		if ok := l.parse(); !ok {
			continue
		}
		if ping, err := download.LinkPing(l.URL); err != nil || ping.StatusCode != 200 {
			continue
		}
		link = l.URL
		save, err = saveName(l.URL)
		if err != nil {
			continue
		}
		break
	}
	return save, link
}

// DownloadsAPIv1 are DownloadLinks for ProductionsAPIv1.
type DownloadsAPIv1 struct {
	LinkClass string `json:"link_class"`
	URL       string `json:"url"`
}

// Downloads parses the Demozoo DownloadLinks and saves the first suitable download.
func (p *ProductionsAPIv1) Downloads() {
	for _, l := range p.DownloadLinks {
		var l DownloadsAPIv1 = l // apply type so we can use it with methods
		if ok := l.parse(); !ok {
			logs.Print(" not usable\n")
			continue
		}
		ping, err := download.LinkPing(l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		if ping.StatusCode != 200 {
			logs.Printf(" %s", ping.Status) // print the HTTP status
			continue
		}
		logs.Print("\n")
		save, err := saveName(l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		temp, err := ioutil.TempDir("", "dzdl")
		if err != nil {
			logs.Log(err)
			continue
		}
		saveDest, err := filepath.Abs(filepath.Join(temp, save))
		if err != nil {
			logs.Log(err)
			continue
		}
		_, err = download.LinkDownload(saveDest, l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		break
	}
}

// Groups returns the first two author_nicks that have is_group flagged.
func (p *ProductionsAPIv1) Groups() [2]string {
	g := [2]string{}
	for i, n := range p.AuthorNicks {
		if i > 1 || !n.Releaser.IsGroup {
			continue
		}
		g[i] = n.Name
	}
	return g
}

type Authors struct {
	text  []string // credit_text
	code  []string // credit_program
	art   []string // credit_illustration
	audio []string // credit_audio
}

func (p *ProductionsAPIv1) Authors() Authors {
	var a Authors
	for _, n := range p.Credits {
		if n.Nick.Releaser.IsGroup {
			continue
		}
		switch strings.ToLower(n.Category) {
		case "text":
			a.text = append(a.text, n.Nick.Name)
		case "code":
			a.code = append(a.code, n.Nick.Name)
		case "graphics":
			a.art = append(a.art, n.Nick.Name)
		case "music":
			a.audio = append(a.audio, n.Nick.Name)
		}
	}
	return a
}

// PouetID returns the ID value used by Pouet's which prod URL syntax
// and its HTTP status code.
// example: https://www.pouet.net/prod.php?which=30352
func (p *ProductionsAPIv1) PouetID(ping bool) (int, int) {
	for _, l := range p.ExternalLinks {
		if l.LinkClass != "PouetProduction" {
			continue
		}
		id, err := parsePouetProduction(l.URL)
		if err != nil {
			logs.Log(err)
			continue
		}
		if ping {
			ping, _ := download.LinkPing(l.URL)
			return id, ping.StatusCode
		}
		return id, 0
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
	const pfx = "productions parse pouet: unexpected"
	if w == "" {
		return 0, fmt.Errorf("%s url syntax %q", pfx, rawurl)
	}
	id, err := strconv.Atoi(w)
	if err != nil {
		return 0, fmt.Errorf("%s which= query syntax %q", pfx, w)
	}
	if id < 0 {
		return 0, fmt.Errorf("%s which= query value %q", pfx, w)
	}
	return id, nil
}

// Print displays the production API results as tabbed JSON.
func (p *ProductionsAPIv1) Print() {
	js, err := json.MarshalIndent(&p, "", "  ")
	logs.Check(err)
	logs.Println(string(js))
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

// parse corrects any known errors with a Downloads API link.
func (dl *DownloadsAPIv1) parse() bool {
	u, err := url.Parse(dl.URL) // validate url
	if err != nil {
		return false
	}
	u = mutateURL(u)
	dl.URL = u.String()
	return true
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
