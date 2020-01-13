package download

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

// UserAgent is the value of User-Agent request HEADER that that lets servers identify this application.
const UserAgent string = "defacto2 cli"

// RFC5322 is a HTTP-date value.
const RFC5322 string = "Mon, 2 Jan 2006 15:04:05 MST"

// Request a HTTP download.
type Request struct {
	Link       string        // URL to request
	Timeout    time.Duration // Timeout in seconds
	Read       []byte        // received HTTP body data
	StatusCode int           // received HTTP statuscode
	Status     string        // received HTTP status
}

// Body fetches a HTTP link and returns its data and the status code.
func (r *Request) Body() error {
	client := http.Client{
		Timeout: timeout(r.Timeout),
	}
	_, err := url.Parse(r.Link)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, r.Link, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	r.Status = res.Status
	r.StatusCode = res.StatusCode
	r.Read, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return nil
}

// timeout creates a valid second duration for use with http.Client.Timeout.
func timeout(t time.Duration) time.Duration {
	if t < 1 {
		t = 5
	}
	return time.Second * t
}

// WriteCounter totals the number of bytes written.
type WriteCounter struct {
	Name    string // Filename
	Total   uint64 // Expected filesize
	Written uint64 // Bytes written
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
	pct := percent(wc.Written, wc.Total)
	if pct > 0 {
		logs.Printf("\rdownloading %s (%d%%) from %s\033[0K", humanize.Bytes(wc.Written), pct, wc.Name)
	} else {
		logs.Printf("\rdownloading %s from %s\033[0K", humanize.Bytes(wc.Written), wc.Name)
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
	logs.Printf("\r%v download saved as: %v\033[0K", humanize.Bytes(uint64(written)), name)
	logs.Println()
}

// LinkDownload downloads the URL and saves it to the named file.
func LinkDownload(name, url string) (http.Header, error) {
	var h http.Header
	out, err := os.Create(name)
	if err != nil {
		return h, err
	}
	defer out.Close()

	res, err := http.Get(url)
	if err != nil {
		return h, err
	}
	defer res.Body.Close()

	counter := &WriteCounter{Name: out.Name(), Total: uint64(res.ContentLength)}
	i, err := io.Copy(out, io.TeeReader(res.Body, counter))
	if err != nil {
		return h, err
	}
	progressDone(out.Name(), i)

	return res.Header, nil
}

// LinkDownloadQ quietly downloads the URL and saves it to the named file.
func LinkDownloadQ(name, url string) (http.Header, error) {
	var h http.Header
	out, err := os.Create(name)
	if err != nil {
		return h, err
	}
	defer out.Close()

	res, err := http.Get(url)
	if err != nil {
		return h, err
	}
	defer res.Body.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return h, err
	}
	return res.Header, nil
}

// LinkPing connects to a URL and returns its HTTP status code and status text.
func LinkPing(url string) (*http.Response, error) {
	res, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return res, nil
}

// StatusColor colours the HTTP status based on its severity.
func StatusColor(code int, status string) string {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
	c := code
	switch {
	case c >= 100 && c <= 199: // Informational responses
		return color.Info.Sprint(status)
	case c >= 200 && c <= 299: // Successful responses
		return color.Success.Sprint(status)
	case c >= 300 && c <= 399: // Redirects
		return color.Notice.Sprint(status)
	case c >= 400 && c <= 499: // Client errors
		return color.Warn.Sprint(status)
	case c >= 500 && c <= 599: // Server errors
		return color.Danger.Sprint(status)
	}
	return color.Question.Sprint(status) // unexpected
}
