package sitemap

import (
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
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/sitemap/internal/urlset"
	"github.com/gookit/color"
)

// Root URL element.
type Root int

const (
	File     Root = iota // File URL element.
	Download             // Download URL element.
)

func (r Root) String() string {
	switch r {
	case File:
		return "f"
	case Download:
		return "d"
	}
	return ""
}

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

// IDs are the primary keys of the file records.
type IDs []int

// Contains returns true whenever a contains x.
func (a *IDs) Contains(x int) bool {
	for _, n := range *a {
		if x == n {
			return true
		}
	}
	return false
}

// Randomize the IDs and returns the first x results.
func (a IDs) Randomize(x int) (IDs, error) {
	l := len(a)
	if l < 1 {
		return nil, errors.New("no ids to randomise")
	}
	if x > l {
		x = l
	}
	randoms := make(IDs, 0, x)
	var seeded int64
	for i := 1; i <= x; i++ {
		seed := time.Now().UnixNano()
		if seed == seeded {
			// avoid duplicates
			i -= 1
			continue
		}
		seeded = seed
		rand.Seed(seed)
		randomIndex := rand.Intn(l)
		id := a[randomIndex]
		if randoms.Contains(id) {
			i -= 1
			continue
		}
		randoms = append(randoms, id)
	}
	return randoms, nil
}

// JoinPaths return the URL strings of the IDs.
func (a IDs) JoinPaths(r Root) []string {
	urls := make([]string, 0, len(a))
	for _, id := range a {
		obfus := database.ObfuscateParam(fmt.Sprint(id))
		link, err := url.JoinPath(Base, r.String(), obfus)
		if err != nil {
			continue
		}
		urls = append(urls, link)
	}
	return urls
}

// Style the result of a link and its status code.
type Style int

// RangeFiles ranges over the file download URLs.
func (p Style) RangeFiles(urls []string) {
	wg := &sync.WaitGroup{}
	for _, link := range urls {
		if link == "" {
			continue
		}
		wg.Add(1)
		go func(link string) {
			link = strings.TrimSpace(link)
			code, name, size, err := download.PingFile(link)
			if err != nil {
				logs.Printf("%s\t%s\n", ColorCode(code), err)
				wg.Done()
				return
			}
			switch p {
			case NotFound:
				fmt.Printf("%s\t%s  ↳ %s - %s\n", link, Color404(code), size, name)
			case Success:
				fmt.Printf("%s\t%s  ↳ %s - %s\n", link, ColorCode(code), size, name)
			}
			wg.Done()
		}(link)
	}
	wg.Wait()
}

// Range over the file URLs.
func (p Style) Range(urls []string) {
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
		go func(link string) {
			link = strings.TrimSpace(link)
			s, code, err := GetTitle(true, link)
			if err != nil {
				logs.Printf("%s\t%s\n", ColorCode(code), err)
				wg.Done()
				return
			}
			switch p {
			case LinkNotFound:
				fmt.Printf("%s\t%s  ↳ %s\n", link, Color404(code), s)
			case LinkSuccess:
				fmt.Printf("%s\t%s  ↳ %s\n", link, ColorCode(code), s)
			case NotFound:
				fmt.Printf("%s\t%s  -  %s\n", Color404(code), link, s)
			case Success:
				fmt.Printf("%s\t%s  -  %s\n", ColorCode(code), link, s)
			}
			wg.Done()
		}(link)
	}
	wg.Wait()
}

// AbsPaths returns all the static URLs used by the sitemap.
func AbsPaths() ([28]string, error) {
	var paths [28]string
	var err error
	for i, path := range urlset.Paths() {
		paths[i], err = url.JoinPath(Base, path)
		if err != nil {
			return [28]string{}, err
		}
	}
	return paths, nil
}

// AbsPaths returns all the HTML3 static URLs used by the sitemap.
func AbsPathsH3() ([]string, error) {
	const root = "html3"
	urls := urlset.Html3Paths()
	paths := make([]string, 0, len(urls))
	for _, elem := range urls {
		path, err := url.JoinPath(Base, elem)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	// create links to platforms
	plats, err := Platforms()
	if err != nil {
		return nil, err
	}
	for _, plat := range plats {
		path, err := url.JoinPath(Base, root, "platform", plat)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	// create links to sections
	sects, err := Sections()
	if err != nil {
		return nil, err
	}
	for _, sect := range sects {
		path, err := url.JoinPath(Base, root, "category", sect)
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
func GetBlocked() (IDs, error) {
	ids, err := database.GetKeys(database.WhereDownloadBlock)
	if err != nil {
		return nil, fmt.Errorf("%w: blocked downloads", err)
	}
	return IDs(ids), nil
}

// GetKeys returns all the primary keys of the file records that are public.
func GetKeys() (IDs, error) {
	ids, err := database.GetKeys(database.WhereAvailable)
	if err != nil {
		return nil, fmt.Errorf("%w: keys", err)
	}
	return IDs(ids), nil
}

// GetSoftDeleteKeys returns all the primary keys of the file records that are not public and hidden.
func GetSoftDeleteKeys() (IDs, error) {
	ids, err := database.GetKeys(database.WhereHidden)
	if err != nil {
		return nil, fmt.Errorf("%w: hidden keys", err)
	}
	return IDs(ids), nil
}

// GetTitle returns the string value of the HTML <title> element and status code of a URL.
func GetTitle(trimSuffix bool, url string) (string, int, error) {
	res, err := download.Ping(url)
	if err != nil {
		return "", res.StatusCode, nil
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", 0, fmt.Errorf("%w: %s", err, url)
	}
	if !trimSuffix {
		return FindTitle(body), res.StatusCode, nil
	}
	s := FindTitle(body)
	return strings.TrimSuffix(s, TitleSuffix), res.StatusCode, nil
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
func RandBlocked(count int) (int, IDs, error) {
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetBlocked()
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Randomize(count)
	return len(keys), res, err
}

// RandDeleted returns a randomized count of primary keys for hidden file records.
func RandDeleted(count int) (int, IDs, error) {
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetSoftDeleteKeys()
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Randomize(count)
	return len(keys), res, err
}

// RandBlocked returns a randomized count of primary keys for public file records.
func RandIDs(count int) (int, IDs, error) {
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetKeys()
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Randomize(count)
	return len(keys), res, err
}

// Platforms lists the operating systems required by the files.
func Platforms() ([]string, error) {
	return database.Distinct("platform")
}

// Sections lists the categories of files.
func Sections() ([]string, error) {
	return database.Distinct("section")
}
