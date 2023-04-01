package sitemap

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/download"
	"github.com/Defacto2/df2/pkg/sitemap/internal/urlset"
	"github.com/google/go-querystring/query"
	"github.com/gookit/color"
)

var ErrNoIDs = errors.New("no ids to randomise")

// Style the result of a link and its status code.
type Style int

const (
	LinkNotFound Style = iota // LinkNotFound first prints the link and expects 404 status codes.
	LinkSuccess               // LinkSuccess first prints the link and expects 200 status codes.
	NotFound                  // NotFound expects 404 status codes.
	Success                   // Success expects 200 status codes.
)

const (
	// TitleSuffix is the string normally appended to most browser tabs.
	TitleSuffix = " | Defacto2"
)

// Root URL element.
type Root int

const (
	File     Root = iota // File URL element.
	Download             // Download URL element.
)

func (r Root) String() string {
	return [...]string{"f", "d"}[r]
}

func Outputs() []string {
	return []string{
		"card", "text", "thumb-",
	}
}

func Sorts() []string {
	return []string{
		"date_asc", "date_desc",
		"posted_asc", "posted_desc",
		"size_asc", "size_desc",
	}
}

type Options struct {
	Output   string `url:"output"`
	Platform string `url:"platform"`
	Section  string `url:"section"`
	Sort     string `url:"sort"`
}

// FileList returns a complete list of URL query strings for the file lists.
func FileList(base string) ([]string, error) {
	const all = "-"
	urls := []string{}
	base, err := url.JoinPath(base, "file", "list", all)
	if err != nil {
		return nil, err
	}
	for _, output := range Outputs() {
		for _, sort := range Sorts() {
			opt := Options{output, all, all, sort}
			v, err := query.Values(opt)
			if err != nil {
				return nil, err
			}
			link := fmt.Sprintf("%s?%s", base, v.Encode())
			urls = append(urls, link)
		}
	}
	return urls, nil
}

// IDs are the primary keys of the file records.
type IDs []int

// Contains returns true whenever IDs contains x.
func (ids *IDs) Contains(x int) bool {
	for _, n := range *ids {
		if x == n {
			return true
		}
	}
	return false
}

// Randomize the IDs and returns the first x results.
func (ids IDs) Randomize(x int) (IDs, error) {
	l := len(ids)
	if l < 1 {
		return nil, ErrNoIDs
	}
	if x > l {
		x = l
	}
	randoms := make(IDs, 0, x)
	seeded := int64(0)
	for i := 1; i <= x; i++ {
		seed := time.Now().UnixNano()
		if seed == seeded {
			// avoid duplicates
			i--
			continue
		}
		seeded = seed
		r := rand.New(rand.NewSource(seed)) //nolint:gosec
		randIndex := r.Intn(l)
		id := ids[randIndex]
		if randoms.Contains(id) {
			i--
			continue
		}
		randoms = append(randoms, id)
	}
	return randoms, nil
}

// JoinPaths return the URL strings of the IDs.
func (ids IDs) JoinPaths(base string, r Root) []string {
	urls := make([]string, 0, len(ids))
	for _, id := range ids {
		obfus := database.ObfuscateParam(fmt.Sprint(id))
		link, err := url.JoinPath(base, r.String(), obfus)
		if err != nil {
			continue
		}
		urls = append(urls, link)
	}
	return urls
}

// RangeFiles ranges over the file download URLs.
func (p Style) RangeFiles(w io.Writer, urls []string) {
	if w == nil {
		w = io.Discard
	}
	var wg sync.WaitGroup
	for _, link := range urls {
		if link == "" {
			continue
		}
		wg.Add(1)
		go func(link string) {
			code, name, size, err := download.PingFile(strings.TrimSpace(link), 0)
			if err != nil {
				fmt.Fprintf(w, "%s\t%s\n", ColorCode(code), err)
				wg.Done()
				return
			}
			switch p {
			case NotFound:
				fmt.Fprintf(w, "%s\t%s  ↳ %s - %s\n", link, Color404(code), size, name)
			case Success:
				fmt.Fprintf(w, "%s\t%s  ↳ %s - %s\n", link, ColorCode(code), size, name)
			case LinkNotFound, LinkSuccess:
				fmt.Fprintf(w, "%q formatting is unused in RangeFiles", p)
			}
			wg.Done()
		}(link)
	}
	wg.Wait()
}

// Range over the file URLs.
func (p Style) Range(w io.Writer, urls []string) {
	const (
		pauseOnItem = 10
		pauseSecs   = 5
	)
	wg := &sync.WaitGroup{}
	for i, link := range urls {
		if link == "" {
			continue
		}
		if i > pauseOnItem-1 && i%pauseOnItem == 0 {
			time.Sleep(pauseSecs * time.Second)
		}
		wg.Add(1)
		go func(w io.Writer, link string) {
			link = strings.TrimSpace(link)
			s, code, err := GetTitle(true, link)
			if err != nil {
				fmt.Fprintf(w, "%s\t%s\n", ColorCode(code), err)
				wg.Done()
				return
			}
			switch p {
			case LinkNotFound:
				fmt.Fprintf(w, "%s\t%s  ↳ %s\n", link, Color404(code), s)
			case LinkSuccess:
				fmt.Fprintf(w, "%s\t%s  ↳ %s\n", link, ColorCode(code), s)
			case NotFound:
				fmt.Fprintf(w, "%s\t%s  -  %s\n", Color404(code), link, s)
			case Success:
				fmt.Fprintf(w, "%s\t%s  -  %s\n", ColorCode(code), link, s)
			}
			wg.Done()
		}(w, link)
	}
	wg.Wait()
}

// AbsPaths returns all the static URLs used by the sitemap.
func AbsPaths(base string) ([28]string, error) {
	var paths [28]string
	var err error
	for i, path := range urlset.Paths() {
		paths[i], err = url.JoinPath(base, path)
		if err != nil {
			return [28]string{}, err
		}
	}
	return paths, nil
}

// AbsPaths returns all the HTML3 static URLs used by the sitemap.
func AbsPathsH3(db *sql.DB, base string) ([]string, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	const root = "html3"
	urls := urlset.HTML3Path()
	paths := make([]string, 0, len(urls))
	for _, elem := range urls {
		path, err := url.JoinPath(base, elem)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	// create links to platforms
	plats, err := Platforms(db)
	if err != nil {
		return nil, err
	}
	for _, plat := range plats {
		path, err := url.JoinPath(base, root, "platform", plat)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	// create links to sections
	sects, err := Sections(db)
	if err != nil {
		return nil, err
	}
	for _, sect := range sects {
		path, err := url.JoinPath(base, root, "category", sect)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
}

// Color404 applies a success checkmark and color when i matches status code 404.
func Color404(i int) string {
	switch i {
	case http.StatusNotFound:
		return color.Success.Sprintf("✓ %d", i)
	case http.StatusOK:
		return color.Danger.Sprintf("✗ %d", i)
	}
	return color.Warn.Sprintf("! %d", i)
}

// ColorCode applies a success checkmark and color when i matches status code 200.
func ColorCode(i int) string {
	switch {
	case i == http.StatusOK:
		return color.Success.Sprintf("✓ %d", i)
	case i >= http.StatusBadRequest:
		return color.Danger.Sprintf("✗ %d", i)
	case i >= http.StatusMultipleChoices:
		return color.Warn.Sprintf("! %d", i)
	default:
		return color.Info.Sprintf("! %d", i)
	}
}

// GetBlocked returns all the primary keys of the records with blocked file downloads.
func GetBlocked(db *sql.DB) (IDs, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	ids, err := database.GetKeys(db, database.WhereDownloadBlock)
	if err != nil {
		return nil, fmt.Errorf("%w: blocked downloads", err)
	}
	return IDs(ids), nil
}

// GetKeys returns all the primary keys of the file records that are public.
func GetKeys(db *sql.DB) (IDs, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	ids, err := database.GetKeys(db, database.WhereAvailable)
	if err != nil {
		return nil, fmt.Errorf("%w: keys", err)
	}
	return IDs(ids), nil
}

// GetSoftDeleteKeys returns all the primary keys of the file records that are not public and hidden.
func GetSoftDeleteKeys(db *sql.DB) (IDs, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	ids, err := database.GetKeys(db, database.WhereHidden)
	if err != nil {
		return nil, fmt.Errorf("%w: hidden keys", err)
	}
	return IDs(ids), nil
}

// GetTitle returns the string value of the HTML <title> element and status code of a URL.
func GetTitle(trimSuffix bool, url string) (string, int, error) {
	b, status, err := download.Get(url, 0)
	if err != nil {
		return "", status, err
	}
	if !trimSuffix {
		return FindTitle(b), status, nil
	}
	s := FindTitle(b)
	return strings.TrimSuffix(s, TitleSuffix), status, nil
}

// FindTitle returns the string value of the HTML <title> element.
func FindTitle(b []byte) string {
	re := regexp.MustCompile(`<title>(\s*.*)<\/title>`)
	elm := re.Find(b)
	if string(elm) == "" {
		return ""
	}
	elm = re.ReplaceAll(elm, []byte(`$1`))
	return string(elm)
}

// RandBlocked returns a randomized count of primary keys for records with blocked file downloads.
func RandBlocked(db *sql.DB, count int) (int, IDs, error) {
	if db == nil {
		return 0, nil, database.ErrDB
	}
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetBlocked(db)
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Randomize(count)
	return len(keys), res, err
}

// RandDeleted returns a randomized count of primary keys for hidden file records.
func RandDeleted(db *sql.DB, count int) (int, IDs, error) {
	if db == nil {
		return 0, nil, database.ErrDB
	}
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetSoftDeleteKeys(db)
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Randomize(count)
	return len(keys), res, err
}

// RandBlocked returns a randomized count of primary keys for public file records.
func RandIDs(db *sql.DB, count int) (int, IDs, error) {
	if db == nil {
		return 0, nil, database.ErrDB
	}
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetKeys(db)
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Randomize(count)
	return len(keys), res, err
}

// Platforms lists the operating systems required by the files.
func Platforms(db *sql.DB) ([]string, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	return database.Distinct(db, "platform")
}

// Sections lists the categories of files.
func Sections(db *sql.DB) ([]string, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	return database.Distinct(db, "section")
}
