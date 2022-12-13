// Package download handles the fetching of remote files.
package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/download/internal/cnter"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

// Request a HTTP download.
type Request struct {
	Link       string        // URL to request.
	Timeout    time.Duration // Timeout duration (5 * time.Second).
	Read       []byte        // HTTP body data received.
	StatusCode int           // HTTP statuscode received.
	Status     string        // HTTP status received.
}

const (
	// UserAgent is the value of User-Agent request HEADER
	// that lets servers identify this application.
	UserAgent = "defacto2 cli"
	// RFC5322 is a HTTP-date value.
	RFC5322 = "Mon, 2 Jan 2006 15:04:05 MST"
	// DownloadPrefix header filename attachment.
	DownloadPrefix = "attachment; filename="

	ContentDisposition = "Content-Disposition"
	ContentLength      = "Content-Length"

	ua = "User-Agent"
)

// Body fetches a HTTP link and returns its data and the status code.
func (r *Request) Body() error {
	_, err := url.Parse(r.Link)
	if err != nil {
		return err
	}
	timeout := CheckTime(r.Timeout)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.Link, nil)
	defer cancel()
	if err != nil {
		return fmt.Errorf("body new context: %w", err)
	}
	req.Header.Set(ua, UserAgent)
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("body client do: %w", err)
	}
	defer res.Body.Close()
	r.Status = res.Status
	r.StatusCode = res.StatusCode
	r.Read, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("body readall: %w", err)
	}
	return nil
}

// CheckTime creates a valid time duration for use with http.Client.Timeout.
// t can be 0 or a number of seconds.
func CheckTime(t time.Duration) time.Duration {
	const maxTime = 5 * time.Second
	secs := time.Duration(t.Seconds())
	if secs < time.Second {
		return maxTime
	}
	return secs
}

// printProgress prints that the download progress is complete.
func progressDone(name string, written int64) {
	logs.Printcrf("%v download saved to: %v", humanize.Bytes(uint64(written)), name)
}

// GetSave downloads the url and saves it as the named file.
func GetSave(name, url string) (http.Header, error) {
	// open local target file
	out, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	// request remote file
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	counter := &cnter.Writer{Name: out.Name(), Total: uint64(resp.ContentLength)}
	i, err := io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return nil, err
	}
	progressDone(out.Name(), i)
	return resp.Header, nil
}

// Silent quietly downloads the URL and saves it to the named file.
// Not in use.
func Silent(name, url string) (http.Header, error) {
	out, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func ping(url, method string, t time.Duration) (*http.Response, error) {
	seconds := t * time.Second
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, seconds)
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
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

// PingGet connects to a URL and returns its response.
func Get(url string) ([]byte, int, error) {
	//return ping(url, http.MethodGet, 60)
	seconds := 30 * time.Second
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, seconds)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	defer cancel()
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set(ua, UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("%w: %s", err, url)
	}
	return body, resp.StatusCode, nil
}

// PingHead connects to a URL and returns its HTTP status code and status text.
func PingHead(url string) (*http.Response, error) {
	return ping(url, http.MethodHead, 5)
}

// PingFile connects to a URL file down and returns its status code, filename and file size.
func PingFile(link string) (code int, name string, size string, err error) {
	res, err := PingHead(link)
	if err != nil {
		return res.StatusCode, "", "", err
	}
	cd, cl := res.Header.Get(ContentDisposition), res.Header.Get(ContentLength)
	name = strings.TrimPrefix(cd, DownloadPrefix)

	i, err := strconv.Atoi(cl)
	if err == nil {
		size = humanize.Bytes(uint64(i))
	}
	return res.StatusCode, name, size, nil
}

// StatusColor colours the HTTP status based on its severity.
func StatusColor(code int, status string) string {
	if status == "" {
		return ""
	}
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
	const (
		infoRespCont   = 100 // Informational responses
		infoRespEnd    = 199
		successOK      = 200 // Successful responses
		successEnd     = 299
		redirectMulti  = 300 // Redirects
		redirectEnd    = 399
		clientBad      = 400 // Client errors
		clientEnd      = 499
		serverInternal = 500 // Server errors
		serverEnd      = 599
	)
	c := code
	switch {
	case c >= infoRespCont && c <= infoRespEnd:
		return color.Info.Sprint(status)
	case c >= successOK && c <= successEnd:
		return color.Success.Sprint(status)
	case c >= redirectMulti && c <= redirectEnd:
		return color.Notice.Sprint(status)
	case c >= clientBad && c <= clientEnd:
		return color.Warn.Sprint(status)
	case c >= serverInternal && c <= serverEnd:
		return color.Danger.Sprint(status)
	}
	return color.Question.Sprint(status) // unexpected
}
