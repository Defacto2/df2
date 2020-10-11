package download

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testTemp() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "tests", "download")
}

func TestRequest_Body(t *testing.T) {
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
			Timeout: 3 * time.Second,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Request{
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

func Test_checkTime(t *testing.T) {
	td := func(v int) time.Duration {
		sec, _ := time.ParseDuration(fmt.Sprintf("%ds", v))
		return sec
	}
	type args struct {
		t time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"0 sec", args{0}, td(5)},
		{"5 secs", args{5}, td(5)},
		{"300 secs", args{300}, td(300)},
		{"-99 secs", args{-99}, td(5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkTime(tt.args.t); got != tt.want {
				t.Errorf("checkTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteCounter_Write(t *testing.T) {
	type fields struct {
		Name    string
		Total   uint64
		Written uint64
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{"empty", fields{"", 0, 0}, args{[]byte("")}, 0, false},
		{"hi", fields{"hi", 2, 2}, args{[]byte("hi")}, 2, false},
		{"some filler text", fields{"x", 2, 6}, args{[]byte("some filler text")}, 16, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc := &WriteCounter{
				Name:    tt.fields.Name,
				Total:   tt.fields.Total,
				Written: tt.fields.Written,
			}
			got, err := wc.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteCounter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WriteCounter.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_percent(t *testing.T) {
	type args struct {
		count uint64
		total uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{"0", args{0, 0}, 0},
		{"1", args{1, 100}, 1},
		{"80", args{33, 41}, 80},
		{"100", args{100, 100}, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percent(tt.args.count, tt.args.total); got != tt.want {
				t.Errorf("percent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinkDownload(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LinkDownload(tt.args.name, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("LinkDownload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// cleanup
			if err := os.Remove(testTemp()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLinkPing(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{""}, true},
		{"fake", args{"https://example.com"}, false},
		{"fake", args{"https://thisisnotaurl-example.com"}, true},
		{"fake", args{"https://example.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := LinkPing(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("LinkPing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				p.Body.Close()
			}
		})
	}
}

func TestLinkDownloadQ(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LinkDownloadQ(tt.args.name, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("LinkDownloadQ() error = %v, wantErr %v", err, tt.wantErr)
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
		{"ok", args{200, "ok"}, "\u001b[1;32mok\u001b[0m"}, // 
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusColor(tt.args.code, tt.args.status); got != tt.want {
				t.Errorf("StatusColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
