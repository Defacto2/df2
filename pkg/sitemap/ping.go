package sitemap

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/sitemap/internal/urlset"
	"github.com/gookit/color"
)

var (
	ErrTimeout = errors.New("request timed out")
)

type (
	IDs      []int
	PathElms [28]string
	Prints   int
)

const (
	LinkNotFound Prints = iota
	LinkSuccess
	NotFound
	Success
)

const (
	TitleSuffix = " | Defacto2"
)

func (p Prints) Range(urls []string) {
	const (
		pauseOnItem = 10
		pauseSecs   = 5
	)
	wg := &sync.WaitGroup{}
	for i, link := range urls {
		if link == "" {
			continue
		}
		if i > 9 && i%pauseOnItem == 0 {
			time.Sleep(pauseSecs * time.Second)
		}
		// todo, validate URL link?
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

func AbsPaths() (PathElms, error) {
	var (
		paths PathElms
		err   error
	)
	for i, path := range urlset.Paths() {
		paths[i], err = url.JoinPath(Base, path)
		if err != nil {
			return PathElms{}, err
		}
	}
	return paths, nil
}

func AbsHtml3s() ([]string, error) {
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
		path, err := url.JoinPath(Base, "html3", "platform", plat)
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
		path, err := url.JoinPath(Base, "html3", "category", sect)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
}

func ColorCode(i int) string {
	const (
		informational = 100
		successful    = 200
		redirectional = 300
		clientError   = 400
		serverError   = 500
	)
	switch {
	case i >= serverError,
		i >= clientError:
		return color.Danger.Sprintf("✗ %d", i)
	case i >= redirectional:
		return color.Warn.Sprintf("! %d", i)
	case i >= successful:
		return color.Success.Sprintf("✓ %d", i)
	case i >= informational:
		return color.Info.Sprintf("! %d", i)
	}
	return strconv.Itoa(i)
}

func Color404(i int) string {
	const (
		successful = 200
		notFound   = 404
	)
	switch i {
	case notFound:
		return color.Success.Sprintf("✓ %d", i)
	case successful:
		return color.Danger.Sprintf("✗ %d", i)
	}
	return color.Warn.Sprintf("! %d", i)
}

func RandDeleted(count int) (int, IDs, error) {
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetKeys(true)
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Random(count)
	return len(keys), res, err
}

func RandIDs(count int) (int, IDs, error) {
	if count < 1 {
		return 0, nil, nil
	}
	keys, err := GetKeys(false)
	if err != nil {
		return 0, nil, err
	}
	res, err := keys.Random(count)
	return len(keys), res, err
}

func (a IDs) URLs() []string {
	urls := make([]string, 0, len(a))
	for _, id := range a {
		obfus := database.ObfuscateParam(fmt.Sprint(id))
		urls = append(urls, fmt.Sprintf("https://defacto2.net/f/%s\n", obfus))
	}
	return urls
}

// Contains tells whether a contains x.
func (a *IDs) Contains(x int) bool {
	for _, n := range *a {
		if x == n {
			return true
		}
	}
	return false
}

func (a IDs) Random(count int) (IDs, error) {
	l := len(a)
	if l < 1 {
		return nil, errors.New("no ids to randomise")
	}
	if count > l {
		count = l
	}
	randoms := make(IDs, 0, count)
	var seeded int64
	for i := 1; i <= count; i++ {
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
		//fmt.Printf("index %d = %d\n", randomIndex, id)
	}
	return randoms, nil
}

func Platforms() ([]string, error) {
	return distinct("platform")
}

func Sections() ([]string, error) {
	return distinct("section")
}

func distinct(value string) ([]string, error) {
	db := database.Connect()
	defer db.Close()
	stmt := fmt.Sprintf("SELECT DISTINCT `%s` FROM `files`", value)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("distinct query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("distinct rows: %w", rows.Err())
	}
	defer rows.Close()
	res := []string{}
	dest := ""
	for rows.Next() {
		if err = rows.Scan(&dest); err != nil {
			return nil, fmt.Errorf("distinct scan: %w", err)
		}
		res = append(res, strings.ToLower(dest))
	}
	return res, nil
}

func GetKeys(deleted bool) (IDs, error) {
	db := database.Connect()
	defer db.Close()
	null := ""
	if deleted {
		null = "NOT "
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM `files` WHERE `deletedat` IS " + null + "NULL").Scan(&count); err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT `id` FROM `files` WHERE `deletedat` IS " + null + "NULL")
	if err != nil {
		return nil, fmt.Errorf("create db query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("create db rows: %w", rows.Err())
	}
	defer rows.Close()
	id, i := "", -1
	keys := make([]int, 0, count)
	for rows.Next() {
		i++
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("create rows next: %w", err)
		}
		val, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		keys = append(keys, val)
	}

	return keys, nil
}

var (
	ErrStatus = errors.New("response failed")
)

func GetTitle(trimSuffix bool, url string) (string, int, error) {
	const timeoutSecs = 60
	client := &http.Client{Timeout: timeoutSecs * time.Second}
	res, err := client.Get(strings.TrimSpace(url))
	if err != nil {
		if os.IsTimeout(err) {
			return "", 0, fmt.Errorf("%w after %v seconds: %s",
				ErrTimeout, timeoutSecs, url)
		}
		return "", 0, fmt.Errorf("%w: %s", err, url)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	// if res.StatusCode >= 500 {
	// 	return "", res.StatusCode, fmt.Errorf("%w: %s", ErrStatus, url)
	// }
	if err != nil {
		return "", 0, fmt.Errorf("%w: %s", err, url)
	}
	if !trimSuffix {
		return FindTitle(body), res.StatusCode, nil
	}
	s := FindTitle(body)
	return strings.TrimSuffix(s, TitleSuffix), res.StatusCode, nil

}

func FindTitle(b []byte) string {
	re := regexp.MustCompile(`<title>(\s*.*)<\/title>`)
	elm := re.Find(b)
	if string(elm) == "" {
		return ""
	}
	elm = re.ReplaceAll(elm, []byte(`$1`))
	return string(elm)
}
