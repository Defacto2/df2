package download_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/download"
	"github.com/stretchr/testify/assert"
)

const (
	lftp   = "ftp://example.com"
	lhttp  = "http://example.com"
	lhttps = "https://example.com"
)

func TestRequest_Body(t *testing.T) {
	t.Parallel()
	const timeout = 3
	r := download.Request{}
	err := r.Body()
	assert.NotNil(t, err)
	r = download.Request{
		Link:    lhttp,
		Timeout: timeout * time.Second,
	}
	err = r.Body()
	assert.Nil(t, err)
	r = download.Request{
		Link:    lhttps,
		Timeout: timeout * time.Second,
	}
	err = r.Body()
	assert.Nil(t, err)
	r = download.Request{
		Link:    lftp,
		Timeout: timeout * time.Second,
	}
	err = r.Body()
	assert.NotNil(t, err)
}

func TestCheckTime(t *testing.T) {
	t.Parallel()
	td := func(v int) time.Duration {
		sec, _ := time.ParseDuration(fmt.Sprintf("%ds", v))
		return sec
	}
	tests := []struct {
		name string
		t    time.Duration
		want time.Duration
	}{
		{"0 sec", 0, td(5)},
		{"5 secs", 5, td(5)},
		{"300 secs", 300, td(5)},
		{"-99 secs", -99, td(5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := download.CheckTime(tt.t); got != tt.want {
				t.Errorf("CheckTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPing(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp(os.TempDir(), "getping")
	assert.Nil(t, err)
	name := f.Name()
	defer os.Remove(name)

	h, err := download.GetSave(io.Discard, "", "")
	assert.NotNil(t, err)
	assert.Empty(t, h)
	h, err = download.GetSave(io.Discard, name, "")
	assert.NotNil(t, err)
	assert.Empty(t, h)
	h, err = download.GetSave(io.Discard, "", lhttps)
	assert.NotNil(t, err)
	assert.Empty(t, h)

	h, err = download.GetSave(io.Discard, name, lhttps)
	assert.Nil(t, err)
	assert.NotEmpty(t, h)
	h, err = download.GetSave(io.Discard, name, lhttp)
	assert.Nil(t, err)
	assert.NotEmpty(t, h)
	h, err = download.GetSave(io.Discard, name, lftp)
	assert.NotNil(t, err)
	assert.Empty(t, h)
}

func TestGet(t *testing.T) {
	t.Parallel()
	b, i, err := download.Get("", 0)
	assert.NotNil(t, err)
	assert.Empty(t, b)
	assert.Equal(t, 0, i)
	b, i, err = download.Get(lftp, 0)
	assert.NotNil(t, err)
	assert.Empty(t, b)
	assert.Equal(t, 0, i)
	b, i, err = download.Get(lhttp, 0)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
	assert.Greater(t, i, 1)
}

func TestPingHead(t *testing.T) {
	t.Parallel()
	r, err := download.PingHead("", 0)
	assert.NotNil(t, err)
	assert.Nil(t, r)
	r, err = download.PingHead(lhttp, 0)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	r, err = download.PingHead(lhttps, 0)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	r, err = download.PingHead(lftp, 0)
	assert.NotNil(t, err)
	assert.Nil(t, r)
	r, err = download.PingHead(lhttp, 1*time.Millisecond)
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestPingFile(t *testing.T) {
	t.Parallel()
	c, n, s, err := download.PingFile("", 0)
	assert.NotNil(t, err)
	assert.Equal(t, 0, c)
	assert.Equal(t, "", n)
	assert.Equal(t, "", s)
	c, n, s, err = download.PingFile(lhttp, 1*time.Millisecond)
	assert.NotNil(t, err)
	assert.Equal(t, 0, c)
	assert.Equal(t, "", n)
	assert.Equal(t, "", s)
	// don't test the returned name as it relies on the remote server content disposition header
	c, _, s, err = download.PingFile(lhttp, 0)
	assert.Nil(t, err)
	assert.Equal(t, 200, c)
	assert.NotEqual(t, "", s)
}

func TestStatusColor(t *testing.T) {
	t.Parallel()
	s := download.StatusColor(-1, "")
	assert.Equal(t, "", s)
	s = download.StatusColor(-1, "error")
	assert.Equal(t, "", s)
	s = download.StatusColor(9999, "")
	assert.Equal(t, "", s)
	s = download.StatusColor(200, "okay")
	assert.Contains(t, s, "ok")
}
