package releases

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

var ErrNegativeID = errors.New("demozoo production id cannot be a negative integer")

const (
	api  = "https://demozoo.org/api/v1/releasers"
	prod = "productions"
)

type Productions []ProductionV1

// Productions releasers productions API v1.
// This can be dynamically generated at https://mholt.github.io/json-to-go/
// Get the Demozoo JSON output from https://demozoo.org/api/v1/releasers/{{.ID}}/productions?format=json
type ProductionV1 struct {
	ExistsInDB bool
	//URL        string `json:"url"`
	//DemozooURL string `json:"demozoo_url"`
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
	// AuthorAffiliationNicks []interface{} `json:"author_affiliation_nicks"`
	ReleaseDate string `json:"release_date"`
	// Supertype              string        `json:"supertype"`
	Platforms []struct {
		URL  string `json:"url"`
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"platforms"`
	Types []struct {
		URL  string `json:"url"`
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"types"`
	// Tags []string `json:"tags"`
}

func Tags(platforms, types string) (platform, section string) {

	// TODO
	// read Title and find "application generator" and set groupapplication

	switch types {
	case "Diskmag", "Textmag":
		section = "magazine"
	case "Game", "Intro", "Demo",
		"96K Intro", "64K Intro", "40k Intro", "32K Intro", "16K Intro", "8K Intro", "4K Intro", "1K Intro", "256b Intro":
		section = "demo"
	case "BBStro":
		section = "bbs"
	case "Cracktro":
		section = "releaseadvert"
	case "Tool":
		section = "programmingtool"
	case "Executable Graphics":
		section = "logo"
		platform = "dos"
	case "Artpack", "Pack", "ASCII Collection":
		section = "bbs"
		platform = "package"
	case "Graphics":
		section = "logo"
		platform = "image"
	case "ANSI":
		section = "logo"
		platform = "ansi"
	case "ASCII":
		section = "logo"
		platform = "text"
	case "Music", "Musicdisk", "Tracked Music":
		section = "demo"
		platform = "audio"
	default:
	}

	switch platforms {
	case "Browser":
		platform = "html"
	case "Java":
		platform = "java"
	case "Linux":
		platform = "linux"
	case "MS-Dos":
		platform = "dos"
	case "Windows":
		platform = "windows"
	default:
	}

	return platform, section
}

// Released returns the production's release date as date_issued_ year, month, day values.
func (p ProductionV1) Released() (year, month, day int) {
	dates := strings.Split(p.ReleaseDate, "-")
	const (
		y = 0
		m = 1
		d = 2
	)
	switch len(dates) {
	case 3:
		year, _ = strconv.Atoi(dates[y])
		month, _ = strconv.Atoi(dates[m])
		day, _ = strconv.Atoi(dates[d])
	case 2:
		year, _ = strconv.Atoi(dates[y])
		month, _ = strconv.Atoi(dates[m])
	case 1:
		year, _ = strconv.Atoi(dates[y])
	default:
	}
	return year, month, day
}

// Groups returns the first two names in the production that have is_group as true.
// The one exception is if the production title contains a reference to a BBS or FTP site name.
// Then that title will be used as the first group returned.
func (p ProductionV1) Groups() (a string, b string) {
	// find any reference to BBS or FTP in the production title to
	// obtain a possible site name.
	if s := Site(p.Title); s != "" {
		a = s
	}
	// range through author nicks for any group matches
	for _, nick := range p.AuthorNicks {
		if !nick.Releaser.IsGroup {
			continue
		}
		if a == "" {
			a = nick.Releaser.Name
			continue
		}
		if b == "" {
			b = nick.Releaser.Name
			break
		}
	}
	return a, b
}

// Site parses a production title to see if it is suitable as a BBS or FTP site name,
// otherwise an empty string is returned.
func Site(title string) string {
	s := strings.Split(title, " ")
	if s[0] == "The" {
		s = s[1:]
	}
	for i, n := range s {
		if n == "BBS" {
			return strings.Join(s[0:i], " ") + " BBS"
		}
		if n == "FTP" {
			return strings.Join(s[0:i], " ") + " FTP"
		}
	}
	return ""
}

// Print to stdout the production API results as tabbed JSON.
func (p *Productions) Print() error {
	js, err := json.MarshalIndent(&p, "", "  ")
	if err != nil {
		return fmt.Errorf("print json marshal indent: %w", err)
	}
	// ignore --quiet
	fmt.Println(string(js))
	return nil
}

// URL generates an API v1 URL used to fetch the productions of a releaser ID.
// i.e. https://demozoo.org/api/v1/releasers/1/productions/
func URL(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("releaser productions id %v: %w", id, ErrNegativeID)
	}
	u, err := url.Parse(api) // base URL
	if err != nil {
		return "", fmt.Errorf("releaser productions parse: %w", err)
	}
	const decimal = 10
	u.Path = path.Join(u.Path, strconv.FormatInt(id, decimal), prod) // append ID
	q := u.Query()
	q.Set("format", "json") // append format=json
	u.RawQuery = q.Encode()
	return u.String(), nil
}
