package prods_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/df2/lib/demozoo/internal/prods"
)

const (
	api = "https://demozoo.org/api/v1/productions"
	cd  = "Content-Disposition"
	cls = "PouetProduction"
	df2 = "defacto2.net"
	dos = "dos"
	win = "windows"
)

const (
	modDate                      = "Wed, 30 Apr 2012 16:29:51 -0500"
	channel1, channel2, channel3 = 1, 2, 3
)

var example1, example2, example3 prods.ProductionsAPIv1 //nolint:gochecknoglobals

var (
	ErrAdd = errors.New("invalid add argument")
	ErrVal = errors.New("unknown record value")
)

func init() { //nolint:gochecknoinits
	c1 := make(chan prods.ProductionsAPIv1)
	c2 := make(chan prods.ProductionsAPIv1)
	c3 := make(chan prods.ProductionsAPIv1)
	go load(channel1, c1)
	go load(channel2, c2)
	go load(channel3, c3)
	example1, example2, example3 = <-c1, <-c2, <-c3
}

func load(r int, c chan prods.ProductionsAPIv1) {
	var name string
	switch r {
	case channel1:
		name = "1"
	case channel2:
		name = "188796"
	case channel3:
		name = "267300"
	default:
		log.Fatal(fmt.Errorf("load r %d: %w", r, ErrVal))
	}
	path, err := filepath.Abs(fmt.Sprintf("../../tests/json/record_%s.json", name))
	if err != nil {
		log.Fatal(fmt.Errorf("path %q: %w", path, err))
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var dz prods.ProductionsAPIv1
	if err := json.Unmarshal(data, &dz); err != nil {
		log.Fatal(fmt.Errorf("load json unmarshal: %w", err))
	}
	c <- dz
}

func Test_filename(t *testing.T) {
	check := func(err error) {
		if err != nil {
			t.Error(err)
		}
	}
	type args struct {
		h http.Header
	}
	cd, err := mockHeader("cd")
	check(err)
	fn, err := mockHeader("fn")
	check(err)
	fn1, err := mockHeader("fn1")
	check(err)
	il, err := mockHeader("il")
	check(err)
	tests := []struct {
		name         string
		args         args
		wantFilename string
	}{
		{"empty", args{}, ""},
		{"empty", args{cd}, ""},
		{"empty", args{il}, ""},
		{"empty", args{fn}, "example.zip"},
		{"empty", args{fn1}, "example.zip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFilename := prods.Filename(tt.args.h); gotFilename != tt.wantFilename {
				t.Errorf("filename() = %v, want %v", gotFilename, tt.wantFilename)
			}
		})
	}
}

func mockHeader(add string) (header http.Header, err error) {
	// source: https://blog.questionable.services/article/testing-http-handlers-go/
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", "/mock-header", nil)
	defer cancel()
	if err != nil {
		return header, err
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	var handler http.HandlerFunc
	switch add {
	case "cd":
		handler = http.HandlerFunc(mockContentDisposition)
	case "fn":
		handler = http.HandlerFunc(mockFilename)
	case "fn1":
		handler = http.HandlerFunc(mockFilename1)
	case "il":
		handler = http.HandlerFunc(mockInline)
	default:
		return header, fmt.Errorf("mock header %q: %w", add, ErrAdd)
	}
	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)
	return rr.Header(), err
}

func mockContentDisposition(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "attachment")
}

func mockFilename1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "attachment; filename*=example.zip;")
	w.Header().Add("modification-date", modDate)
}

func mockFilename(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "attachment; filename=example.zip;")
	w.Header().Add("modification-date", modDate)
}

func mockInline(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "inline")
}

func TestProductionsAPIv1_DownloadLink(t *testing.T) {
	tests := []struct {
		name     string
		p        prods.ProductionsAPIv1
		wantName string
		wantLink string
	}{
		{"empty", prods.ProductionsAPIv1{}, "", ""},
		{
			"record 1", example1, "feestje.zip",
			"https://files.scene.org/get:nl-http/parties/2000/ambience00/demo/feestje.zip",
		},
		{
			"record 2", example2, "the_untouchables_bbs7.zip",
			"http://www.sensenstahl.com/untergrund_mirror/bbs/the_untouchables_bbs7.zip",
		},
		{
			"record 3", example3, "x-wing_cracktro.zip",
			"http://www.sensenstahl.com/untergrund_mirror/cracktro/x-wing_cracktro.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotLink := tt.p.DownloadLink()
			if gotName != tt.wantName {
				t.Errorf("ProductionsAPIv1.DownloadLink() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotLink != tt.wantLink {
				t.Errorf("ProductionsAPIv1.DownloadLink() gotLink = %v, want %v", gotLink, tt.wantLink)
			}
		})
	}
}

func TestProductionsAPIv1_PouetID(t *testing.T) {
	tests := []struct {
		name           string
		p              prods.ProductionsAPIv1
		wantID         int
		wantStatusCode int
		ping           bool
		wantErr        bool
	}{
		{"empty", prods.ProductionsAPIv1{}, 0, 0, false, false},
		{"record 1", example1, 7084, 200, true, false},
		{"record 2", example2, 76652, 200, true, false},
		{"record 3 (no pouet)", example3, 0, 0, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotStatusCode, err := tt.p.PouetID(tt.ping)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductionsAPIv1.PouetID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("ProductionsAPIv1.PouetID() gotID = %v, want %v", gotID, tt.wantID)
			}
			if gotStatusCode != tt.wantStatusCode {
				t.Errorf("ProductionsAPIv1.PouetID() gotStatusCode = %v, want %v",
					gotStatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestProductionsAPIv1_Print(t *testing.T) {
	tests := []struct {
		name    string
		p       prods.ProductionsAPIv1
		wantErr bool
	}{
		{"empty", prods.ProductionsAPIv1{}, false},
		{"record 1", example1, false},
		{"record 2", example2, false},
		{"record 3", example3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Print(); (err != nil) != tt.wantErr {
				t.Errorf("ProductionsAPIv1.Print() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMutateURL(t *testing.T) {
	exp, _ := url.Parse("http://example.com")
	bro, _ := url.Parse("not-a-valid-url")
	fso, _ := url.Parse("https://files.scene.org/view/someplace")
	type args struct {
		u *url.URL
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"example", args{u: exp}, "http://example.com"},
		{"nil", args{nil}, ""},
		{"broken", args{bro}, "not-a-valid-url"},
		{"scene.org", args{fso}, "https://files.scene.org/get:nl-http/someplace"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prods.MutateURL(tt.args.u).String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MutateURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePouetProduction(t *testing.T) {
	type args struct {
		rawurl string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"valid", args{"https://www.pouet.net/prod.php?which=30352"}, 30352, false},
		{"valid", args{"https://www.pouet.net/prod.php?which=1"}, 1, false},

		{"letters", args{"https://www.pouet.net/prod.php?which=abc"}, 0, true},
		{"negative", args{"https://www.pouet.net/prod.php?which=-1"}, 0, true},
		{"malformed", args{"https://www.pouet.net/"}, 0, true},
		{"empty", args{""}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prods.ParsePouetProduction(tt.args.rawurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePouetProduction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePouetProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randomName(t *testing.T) {
	rand := func() bool {
		r, err := prods.RandomName()
		if err != nil {
			t.Error(err)
		}
		fmt.Println(r)
		return strings.Contains(r, "df2-download")
	}
	tests := []struct {
		name string
		want bool
	}{
		{"test random value", true},
		{"another random value", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rand(); got != tt.want {
				t.Errorf("randomName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_saveName(t *testing.T) {
	type args struct {
		rawurl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"file", args{"blob"}, "blob", false},
		{"file+ext", args{"blob.txt"}, "blob.txt", false},
		{"url", args{"https://example.com/myfile.txt"}, "myfile.txt", false},
		{"deep", args{"https://example.com/path/to/some/file/down/here/myfile.txt"}, "myfile.txt", false},
		{"no ext", args{"https://example.com/txt/myfile"}, "myfile", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prods.SaveName(tt.args.rawurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("saveName() = %v, want %v", got, tt.want)
			}
		})
	}
}
