package download_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/download"
	"github.com/gookit/color"
)

func testTemp() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "tests", "download")
}

func TestRequest_Body(t *testing.T) {
	const timeout = 3
	type fields struct {
		Link       string
		Timeout    time.Duration
		Read       []byte
		StatusCode int
		Status     string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty", fields{}, true},
		{"example", fields{
			Link:    "https://example.com",
			Timeout: timeout * time.Second,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &download.Request{
				Link:       tt.fields.Link,
				Timeout:    tt.fields.Timeout,
				Read:       tt.fields.Read,
				StatusCode: tt.fields.StatusCode,
				Status:     tt.fields.Status,
			}
			if err := r.Body(); (err != nil) != tt.wantErr {
				t.Errorf("Request.Body() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckTime(t *testing.T) {
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
	type args struct {
		name string
		url  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{testTemp(), ""}, true},
		{"ftp", args{testTemp(), "ftp://example.com"}, true},
		{"fake", args{testTemp(), "https://thisisnotaurl-example.com"}, true},
		{"exp", args{testTemp(), "http://example.com"}, false},
	}
	if _, err := os.Create(testTemp()); err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := download.Get(tt.args.name, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// cleanup
			if err := os.Remove(testTemp()); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := download.Ping(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				p.Body.Close()
			}
		})
	}
}

func TestSilent(t *testing.T) {
	type args struct {
		name string
		url  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{testTemp(), ""}, true},
		{"ftp", args{testTemp(), "ftp://example.com"}, true},
		{"fake", args{testTemp(), "https://thisisnotaurl-example.com"}, true},
		{"exp", args{testTemp(), "http://example.com"}, false},
	}
	if _, err := os.Create(testTemp()); err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := download.Silent(tt.args.name, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("Silent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// cleanup
			if err := os.Remove(testTemp()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestStatusColor(t *testing.T) {
	type args struct {
		code   int
		status string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"ok", args{200, "ok"}, "ok"},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := download.StatusColor(tt.args.code, tt.args.status); got != tt.want {
				t.Errorf("StatusColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
