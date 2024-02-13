// Package download fetches remote files used by this program and the website.
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
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

// Request a HTTP download.
type Request struct {
	Link    string        // URL to request.
	Timeout time.Duration // Timeout duration (5 * time.Second).
	Read    []byte        // HTTP body data received.
	Code    int           // HTTP statuscode received.
	Status  string        // HTTP status received.
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
	if _, err := url.Parse(r.Link); err != nil {
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
	r.Code = res.StatusCode
	r.Read, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("body read all: %w", err)
	}
	return nil
}

// CheckTime creates a valid time duration for use with http.Client.Timeout.
// The t value can be 0 or a number of seconds.
func CheckTime(t time.Duration) time.Duration {
	const maxTime = 10 * time.Second
	secs := time.Duration(t.Seconds())
	if secs < time.Second {
		return maxTime
	}
	return secs
}

// printProgress prints that the download progress is complete.
func progressDone(w io.Writer, name string, written int64) {
	if w == nil {
		w = io.Discard
	}
	logger.PrintfCR(w, "%v download saved to: %v", humanize.Bytes(uint64(written)), name)
}

// GetSave downloads the url and saves it as the named file.
func GetSave(w io.Writer, name, url string) (http.Header, error) {
	if w == nil {
		w = io.Discard
	}
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
	progressDone(w, out.Name(), i)
	return resp.Header, nil
}

func ping(url, method string, timeout time.Duration) (*http.Response, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
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
func Get(url string, timeout time.Duration) ([]byte, int, error) {
	if timeout == 0 {
		const httpTimeout = 15 * time.Second
		timeout = httpTimeout
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
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
func PingHead(url string, timeout time.Duration) (*http.Response, error) {
	if timeout == 0 {
		const httpTimeout = 10 * time.Second
		timeout = httpTimeout
	}
	return ping(url, http.MethodHead, timeout)
}

// PingFile connects to a URL file down and returns its status code, filename and file size.
func PingFile(link string, timeout time.Duration) ( //nolint:nonamedreturns
	code int, name string, size string, err error,
) {
	res, err := PingHead(link, timeout)
	if err != nil {
		if res != nil {
			return res.StatusCode, "", "", err
		}
		return 0, "", "", err
	}
	cd, cl := res.Header.Get(ContentDisposition), res.Header.Get(ContentLength)
	n := strings.TrimPrefix(cd, DownloadPrefix)
	res.Body.Close()

	b := ""
	i, err := strconv.Atoi(cl)
	if err == nil {
		b = humanize.Bytes(uint64(i))
	}
	return res.StatusCode, n, b, nil
}

// StatusColor colours the HTTP status based on its severity.
func StatusColor(code int, status string) string { //nolint:cyclop
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
	if c < infoRespCont || status == "" {
		return ""
	}
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
