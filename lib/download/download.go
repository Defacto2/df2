package download

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
)

// UserAgent is the value of User-Agent request HEADER that that lets servers identify this application.
const UserAgent string = "defacto2 cli"

// Request a download.
type Request struct {
	Link    string        // URL to request
	Timeout time.Duration // Timeout in seconds
}

// WriteCounter totals the number of bytes written.
type WriteCounter struct {
	Name    string // Filename
	Total   uint64 // Expected filesize
	Written uint64 // Bytes written
}

// Body fetches a HTTP link and returns its data.
func (r Request) Body() *[]byte {
	client := http.Client{
		Timeout: timeout(r.Timeout),
	}
	_, err := url.Parse(r.Link) // validate url
	logs.Check(err)
	logs.Printf("Requesting HTTP %v", r.Link)
	req, err := http.NewRequest(http.MethodGet, r.Link, nil)
	logs.Check(err)
	res, err := client.Do(req)
	logs.Check(err)
	logs.Printf(" %s\n", res.Status)
	body, err := ioutil.ReadAll(res.Body)
	logs.Check(err)
	return &body
}

// timeout creates a valid second duration for use with http.Client.Timeout.
func timeout(t time.Duration) time.Duration {
	if t < 1 {
		t = 5
	}
	return time.Second * t
}

// Write progress counter.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Written += uint64(n)
	wc.printProgress()
	return n, nil
}

// printProgress prints the current download progress.
func (wc WriteCounter) printProgress() {
	logs.Printf("\r%s", strings.Repeat(" ", 80)) // clear line
	pct := percent(wc.Written, wc.Total)
	if pct > 0 {
		logs.Printf("\rDownloading %s (%d%%) from %s", humanize.Bytes(wc.Written), pct, wc.Name)
	} else {
		logs.Printf("\rDownloading %s from %s", humanize.Bytes(wc.Written), wc.Name)
	}
}

func percent(count uint64, total uint64) uint64 {
	if total == 0 {
		return 0
	}
	return uint64(float64(count) / float64(total) * 100)
}

// printProgress prints that the download progress is complete.
func progressDone(name string, written int64) {
	logs.Printf("\r%s", strings.Repeat(" ", 35))
	logs.Printf("\rDownload %v and saved as: %v", humanize.Bytes(uint64(written)), name)
	logs.Println()
}

// LinkDownload downloads the URL and saves it to the named file.
func LinkDownload(name, url string) error {
	out, err := os.Create(name)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	counter := &WriteCounter{Name: out.Name(), Total: uint64(resp.ContentLength)}

	i, err := io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}
	progressDone(out.Name(), i)
	return nil
}

// LinkPing connects to a URL and returns its HTTP status code and status text.
func LinkPing(url string) (int, string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Status, nil
}

// LinkQuietGet quietly downloads the URL and saves it to the named file.
func LinkQuietGet(name, url string) error {
	out, err := os.Create(name)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
