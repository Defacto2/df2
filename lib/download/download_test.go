package download

import (
	"fmt"
	"testing"
	"time"
)

func Test_timeout(t *testing.T) {
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
			if got := timeout(tt.args.t); got != tt.want {
				t.Errorf("timeout() = %v, want %v", got, tt.want)
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
		{"empty", args{"temp", ""}, true},
		{"ftp", args{"temp", "ftp://example.com"}, true},
		{"fake", args{"temp", "https://thisisnotaurl-example.com"}, true},
		{"exp", args{"temp", "http://example.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LinkDownload(tt.args.name, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("LinkDownload() error = %v, wantErr %v", err, tt.wantErr)
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
			_, err := LinkPing(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("LinkPing() error = %v, wantErr %v", err, tt.wantErr)
				return
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
		{"empty", args{"temp", ""}, true},
		{"ftp", args{"temp", "ftp://example.com"}, true},
		{"fake", args{"temp", "https://thisisnotaurl-example.com"}, true},
		{"exp", args{"temp", "http://example.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LinkDownloadQ(tt.args.name, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("LinkDownloadQ() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
