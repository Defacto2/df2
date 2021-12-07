// Package download handles the fetching of remote files.
package download

import (
	"context"
	"fmt"
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

// Request a HTTP download.
type Request struct {
	Link       string        // URL to request
	Timeout    time.Duration // Timeout in seconds
	Read       []byte        // received HTTP body data
	StatusCode int           // received HTTP statuscode
	Status     string        // received HTTP status
}

const (
	// UserAgent is the value of User-Agent request HEADER that that lets servers identify this application.
	UserAgent = "defacto2 cli"
	// RFC5322 is a HTTP-date value.
	RFC5322 = "Mon, 2 Jan 2006 15:04:05 MST"

	ua             = "User-Agent"
	infoRespCont   = 100
	infoRespEnd    = 199
	successOK      = 200
	successEnd     = 299
	redirectMulti  = 300
	redirectEnd    = 399
	clientBad      = 400
	clientEnd      = 499
	serverInternal = 500
	serverEnd      = 599
)

// Body fetches a HTTP link and returns its data and the status code.
func (r *Request) Body() error {
	_, err := url.Parse(r.Link)
	if err != nil {
		return err
	}
	timeout := checkTime(r.Timeout)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.Link, nil)
	defer cancel()
	if err != nil {
		return fmt.Errorf("request body new with context: %w", err)
	}
	req.Header.Set(ua, UserAgent)
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request body client do: %w", err)
	}
	defer res.Body.Close()
	r.Status = res.Status
	r.StatusCode = res.StatusCode
	r.Read, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("request body read body response: %w", err)
	}
	return nil
}

// checkTime creates a valid second duration for use with http.Client.Timeout.
func checkTime(t time.Duration) time.Duration {
	const timeout = 5
	if t < 1 {
		return time.Duration(time.Duration(timeout).Seconds())
	}
	return time.Duration(t.Seconds())
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
		logs.Printcrf("downloading %s (%d%%) from %s", humanize.Bytes(wc.Written), pct, wc.Name)
	} else {
		logs.Printcrf("downloading %s from %s", humanize.Bytes(wc.Written), wc.Name)
	}
}

func percent(count, total uint64) uint64 {
	if total == 0 {
		return 0
	}
	const max = 100
	return uint64(float64(count) / float64(total) * max)
}

// printProgress prints that the download progress is complete.
func progressDone(name string, written int64) {
	logs.Printcrf("%v download saved to: %v", humanize.Bytes(uint64(written)), name)
}

// LinkDownload downloads the link and saves it as the named file.
func LinkDownload(name, link string) (http.Header, error) {
	// open local target file
	out, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	// request remote file
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(ua, UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// download and write remote file to local target
	counter := &WriteCounter{Name: out.Name(), Total: uint64(resp.ContentLength)}
	i, err := io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return nil, err
	}
	progressDone(out.Name(), i)
	return resp.Header, nil
}

// LinkDownloadQ quietly downloads the URL and saves it to the named file.
// Not used.
func LinkDownloadQ(name, link string) (http.Header, error) {
	out, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(ua, UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// download and write remote file to the local target
	if _, err := io.Copy(out, resp.Body); err != nil {
		return nil, err
	}
	return resp.Header, nil
}

// LinkPing connects to a URL and returns its HTTP status code and status text.
func LinkPing(link string) (*http.Response, error) {
	const seconds = 5 * time.Second
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, seconds)
	req, err := http.NewRequestWithContext(ctx, "HEAD", link, nil)
	defer cancel()
	if err != nil {
		return nil, err
	}
	req.Header.Set(ua, UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// StatusColor colours the HTTP status based on its severity.
func StatusColor(code int, status string) string {
	if status == "" {
		return ""
	}
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
	c := code
	switch {
	case c >= infoRespCont && c <= infoRespEnd: // Informational responses
		return color.Info.Sprint(status)
	case c >= successOK && c <= successEnd: // Successful responses
		return color.Success.Sprint(status)
	case c >= redirectMulti && c <= redirectEnd: // Redirects
		return color.Notice.Sprint(status)
	case c >= clientBad && c <= clientEnd: // Client errors
		return color.Warn.Sprint(status)
	case c >= serverInternal && c <= serverEnd: // Server errors
		return color.Danger.Sprint(status)
	}
	return color.Question.Sprint(status) // unexpected
}
