package sitemap

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
)

type IDs []int

func RandDeleted(count int) (IDs, error) {
	if count < 1 {
		return nil, nil
	}
	keys, err := GetKeys(true)
	if err != nil {
		return nil, err
	}
	return keys.Random(count)
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
	// https://defacto2.net/f/b22af71?name=-&platform=-&section=releaseadvert&sort=date_desc
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

func GetTitle(url string) (string, int, error) {
	res, err := http.Get(strings.TrimSpace(url))
	if err != nil {
		return "", 0, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		return "", res.StatusCode, ErrStatus
	}
	if err != nil {
		return "", 0, err
	}
	return FindTitle(body), res.StatusCode, nil
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
