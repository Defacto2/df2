package releases

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
)

var (
	ErrNegativeID = errors.New("demozoo production id cannot be a negative integer")
)

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
	ID    int    `json:"id"`
	Title string `json:"title"`
	// AuthorNicks []struct {
	// 	Name         string `json:"name"`
	// 	Abbreviation string `json:"abbreviation"`
	// 	Releaser     struct {
	// 		URL     string `json:"url"`
	// 		ID      int    `json:"id"`
	// 		Name    string `json:"name"`
	// 		IsGroup bool   `json:"is_group"`
	// 	} `json:"releaser"`
	// } `json:"author_nicks"`
	// AuthorAffiliationNicks []interface{} `json:"author_affiliation_nicks"`
	// ReleaseDate            string        `json:"release_date"`
	// Supertype              string        `json:"supertype"`
	// Platforms              []struct {
	// 	URL  string `json:"url"`
	// 	ID   int    `json:"id"`
	// 	Name string `json:"name"`
	// } `json:"platforms"`
	// Types []struct {
	// 	URL  string `json:"url"`
	// 	ID   int    `json:"id"`
	// 	Name string `json:"name"`
	// } `json:"types"`
	// Tags []string `json:"tags"`
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
