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
	v1           = "https://demozoo.org/api/v1"
	releasersAPI = v1 + "/releasers"
	prod         = "productions"
)

// Filter the Production List using API fields.
type Filter uint

const (
	MsDos   Filter = iota // MsDos filters by platforms id 4.
	Windows               // Windows filters by platforms id 1.
)

func (f Filter) String() string {
	switch f {
	case MsDos:
		const before = "2000-01-01"
		return v1 + "/productions/?supertype=production&title=&platform=4&released_before=" +
			before + "&released_since=&added_before=&added_since=&updated_before=&updated_since=&author="
	case Windows:
		const before = ""
		return v1 + "/productions/?supertype=production&title=&platform=1&released_before=" +
			before + "&released_since=&added_before=&added_since=&updated_before=&updated_since=&author="
	}
	return ""
}

type Productions []ProductionV1

// Productions releasers productions API v1.
// This can be dynamically generated at https://mholt.github.io/json-to-go/
// Get the Demozoo JSON output from https://demozoo.org/api/v1/releasers/{{.ID}}/productions?format=json
type ProductionV1 struct {
	ExistsInDB bool
	// URL        string `json:"url"`
	// DemozooURL string `json:"demozoo_url"`
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
	Types []Type `json:"types"`
	// Types []struct {
	// 	URL  string `json:"url"`
	// 	ID   int    `json:"id"`
	// 	Name string `json:"name"`
	// } `json:"types"`
	Tags []string `json:"tags"`
}

// Type of production.
type Type struct {
	URL  string `json:"url"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func Tags(platforms, types, title string) (platform, section string) {
	const logo = "logo"
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
		section = logo
		platform = "dos"
	case "Artpack", "Pack", "ASCII Collection":
		section = "bbs"
		platform = "package"
	case "Graphics":
		section = logo
		platform = "image"
	case "ANSI":
		section = logo
		platform = "ansi"
	case "ASCII":
		section = logo
		platform = "text"
	case "Music", "Musicdisk", "Tracked Music":
		section = "demo"
		platform = "audio"
	default:
	}
	if strings.Contains(strings.ToLower(title), "application generator") {
		section = "groupapplication"
	}
	if p := tagPlatform(platforms); p != "" {
		platform = p
	}
	return platform, section
}

func tagPlatform(platforms string) string {
	switch platforms {
	case "Browser":
		return "html"
	case "Java":
		return "java"
	case "Linux":
		return "linux"
	case "MS-Dos":
		return "dos"
	case "Windows":
		return "windows"
	default:
		return ""
	}
}

// Released returns the production's release date as date_issued_ year, month, day values.
func (p ProductionV1) Released() (year, month, day int) {
	dates := strings.Split(p.ReleaseDate, "-")
	const (
		y    = 0
		m    = 1
		d    = 2
		ymd  = 3
		ym   = 2
		yyyy = 1
	)
	switch len(dates) {
	case ymd:
		year, _ = strconv.Atoi(dates[y])
		month, _ = strconv.Atoi(dates[m])
		day, _ = strconv.Atoi(dates[d])
	case ym:
		year, _ = strconv.Atoi(dates[y])
		month, _ = strconv.Atoi(dates[m])
	case yyyy:
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

// URL generates an API v1 URL used to fetch the productions filtered by a productions id.
// i.e. https://demozoo.org/api/v1/productions/?supertype=production&title=&platform=4
func URLFilter(f Filter) (string, error) {
	u, err := url.Parse(f.String()) // base URL
	if err != nil {
		return "", fmt.Errorf("releaser productions parse: %w", err)
	}
	q := u.Query()
	q.Set("format", "json") // append format=json
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// URL generates an API v1 URL used to fetch the productions of a releaser ID.
// i.e. https://demozoo.org/api/v1/releasers/1/productions/
func URLReleasers(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("releaser productions id %v: %w", id, ErrNegativeID)
	}
	u, err := url.Parse(releasersAPI) // base URL
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
