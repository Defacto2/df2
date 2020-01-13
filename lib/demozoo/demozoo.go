package demozoo

import (
	"crypto/md5"
	"crypto/sha512"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/download"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

const (
	prodAPI string = "https://demozoo.org/api/v1/productions"
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

// Production API production request.
type Production struct {
	ID         int64         // Demozoo production ID
	Timeout    time.Duration // HTTP request timeout in seconds (default 5)
	link       string        // URL link to sent the request // ??
	StatusCode int           // received HTTP statuscode
	Status     string        // received HTTP status
}

// Fetch a Demozoo production by its ID.
func Fetch(id uint) (int, string, ProductionsAPIv1) {
	var d = Production{
		ID: int64(id),
	}
	api := *d.data()
	return d.StatusCode, d.Status, api
}

// data gets a production API link and normalises the results.
func (p *Production) data() *ProductionsAPIv1 {
	var err error
	p.URL()
	var r = download.Request{
		Link: p.link,
	}
	err = r.Body()
	logs.Log(err)
	p.Status = r.Status
	p.StatusCode = r.StatusCode
	dz := ProductionsAPIv1{}
	if len(r.Read) > 0 {
		err = json.Unmarshal(r.Read, &dz)
	}
	if err != nil && logs.Panic {
		logs.Println(string(r.Read))
	}
	logs.Check(err)
	return &dz
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
		saveDest, err := filepath.Abs(filepath.Join("/home", "ben", save)) // TODO PATH arg instead of hardcoded
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
		ping, _ := download.LinkPing(l.URL)
		return id, ping.StatusCode
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

type row struct {
	base    string
	count   int
	missing int
}

// Request proofs.
type Request struct {
	Overwrite bool // overwrite existing files
	All       bool // parse all proofs
	HideMiss  bool // ignore missing uuid files
}

var prodID = ""

// Query parses a single Demozoo entry.
func (req Request) Query(id string) error {
	if !database.UUID(id) && !database.ID(id) {
		return fmt.Errorf("invalid id given %q it needs to be an auto-generated MySQL id or an uuid", id)
	}
	prodID = id
	return req.Queries()
}

// Record of a file item.
type Record struct {
	count          int
	AbsFile        string // absolute path to file
	ID             string // mysql auto increment id
	UUID           string // record unique id
	WebIDDemozoo   string // demozoo production id
	Filename       string
	Filesize       string
	FileZipContent string
	CreatedAt      string
	UpdatedAt      string
	SumMD5         string    // file download MD5 hash
	Sum384         string    // file download SHA384 hash
	LastMod        time.Time // file download last modified time
}

func (r Record) String() string {
	return fmt.Sprintf("%s item %04d (%v) %v DZ:%v %v",
		logs.Y(), r.count, r.ID,
		color.Primary.Sprint(r.UUID), color.Note.Sprint(r.WebIDDemozoo),
		r.CreatedAt)
}

func sqlSelect() string {
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`,`web_id_demozoo`,`file_zip_content`,`updatedat`,`platform`"
	w := " WHERE `web_id_demozoo` IS NOT NULL AND `platform` = 'dos'"
	if prodID != "" {
		switch {
		case database.UUID(prodID):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, prodID)
		case database.ID(prodID):
			w = fmt.Sprintf("%v AND `id`=%q", w, prodID)
		}
	}
	return s + " FROM `files`" + w
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (req Request) Queries() error {
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(sqlSelect())
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	dir := directories.Init(false)
	// fetch the rows
	rw := row{count: 0, missing: 0}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := database.IsNew(values); !new && !req.All {
			continue
		}
		rw.count++
		r := Record{
			count:          rw.count,
			ID:             string(values[0]),
			UUID:           string(values[1]),
			UpdatedAt:      database.DateTime(values[2]),
			CreatedAt:      database.DateTime(values[3]),
			Filename:       string(values[4]),
			Filesize:       string(values[5]),
			WebIDDemozoo:   string(values[6]),
			FileZipContent: string(values[7])}
		r.AbsFile = filepath.Join(dir.UUID, r.UUID)
		// confirm UUID is missing
		if !rw.skipRow(r) {
			continue // TODO maybe still remove this and instead handle filename = null
		}
		logs.Print(r)
		// confirm API request
		code, status, api := Fetch(273607)
		if code < 200 || code > 299 {
			logs.Printf("(%s)\n", download.StatusColor(code, status))
			continue
		}
		// handle download
		// ?look for existing file
		switch {
		case r.Filesize == "", r.Filename == "":
			name, link := api.DownloadLink()
			if len(link) == 0 {
				logs.Print(color.Note.Sprint("no suitable downloads found\n"))
				continue
			}
			logs.Printf("%s %s\n", color.Primary.Sprint(link), download.StatusColor(200, "200 OK"))
			head, err := download.LinkDownload(name, link)
			if err != nil {
				logs.Log(err)
				continue
			}
			logs.Print(" •")
			// last modified time passed via HTTP
			if lm := head.Get("Last-Modified"); len(lm) > 0 {
				if t, err := time.Parse(download.RFC5322, lm); err == nil {
					r.LastMod = t
				}
			}
			if err != nil {
				logs.Log(err)
				continue
			}
			stat, err := os.Stat(name)
			if err != nil {
				logs.Log(err)
				continue
			}
			r.AbsFile = stat.Name()
			r.Filesize = strconv.Itoa(int(stat.Size()))
			// Hashes (move to func) --->
			f, err := os.Open(stat.Name())
			if err != nil {
				logs.Log(err)
				continue
			}
			defer f.Close()
			h1 := md5.New()
			if _, err := io.Copy(h1, f); err != nil {
				logs.Log(err)
				continue
			}
			h2 := sha512.New384()
			if _, err := io.Copy(h2, f); err != nil {
				logs.Log(err)
				continue
			}
			r.SumMD5 = fmt.Sprintf("%x", h1.Sum(nil))
			r.Sum384 = fmt.Sprintf("%x", h2.Sum(nil))
			fallthrough
		case r.FileZipContent == "":
			if zip := r.fileZipContent(); zip {
				err := archive.ExtractDemozoo(r.AbsFile, r.UUID)
				logs.Log(err)
			}
		}
		logs.Println()
	}
	logs.Check(rows.Err())
	rw.summary()
	return nil
}

// fileZipContent reads an archive and saves its content to the database
func (r *Record) fileZipContent() bool {
	a, err := archive.Read(r.AbsFile)
	if err != nil {
		logs.Log(err)
		return false
	}
	r.FileZipContent = strings.Join(a, "\n")
	//updateZipContent(r.ID, strings.Join(a, "\n"))
	return true
}

func (rw row) skipRow(r Record) bool {
	if _, err := os.Stat(r.AbsFile); os.IsNotExist(err) {
		rw.missing++
		return true
	}
	return false
}

func (rw row) summary() {
	t := fmt.Sprintf("Total Demozoo items handled: %v", rw.count)
	logs.Println(strings.Repeat("─", len(t)))
	logs.Println(t)
	if rw.missing > 0 {
		logs.Println("UUID files not found:", rw.missing)
	}
}
