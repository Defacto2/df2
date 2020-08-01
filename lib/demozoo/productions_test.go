package demozoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
)

const modDate = "Wed, 30 Apr 2012 16:29:51 -0500"

var example1, example2, example3 ProductionsAPIv1

func init() {
	c1 := make(chan ProductionsAPIv1)
	c2 := make(chan ProductionsAPIv1)
	c3 := make(chan ProductionsAPIv1)
	go load(1, c1)
	go load(2, c2)
	go load(3, c3)
	example1, example2, example3 = <-c1, <-c2, <-c3
}

func load(r int, c chan ProductionsAPIv1) {
	var name string
	switch r {
	case 1:
		name = "1"
	case 2:
		name = "188796"
	case 3:
		name = "267300"
	default:
		log.Fatal(fmt.Errorf("load r %d: %w", r, errors.New("unknown record value")))
	}
	path, err := filepath.Abs(fmt.Sprintf("../../tests/json/record_%s.json", name))
	if err != nil {
		log.Fatal(fmt.Errorf("path %q: %w", path, err))
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var dz ProductionsAPIv1
	if err := json.Unmarshal(data, &dz); err != nil {
		log.Fatal(fmt.Errorf("load json unmarshal: %w", err))
	}
	c <- dz
}

func Test_filename(t *testing.T) {
	type args struct {
		h http.Header
	}
	cd, err := mockHeader("cd")
	check(t, err)
	fn, err := mockHeader("fn")
	check(t, err)
	fn1, err := mockHeader("fn1")
	check(t, err)
	il, err := mockHeader("il")
	check(t, err)
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
			if gotFilename := filename(tt.args.h); gotFilename != tt.wantFilename {
				t.Errorf("filename() = %v, want %v", gotFilename, tt.wantFilename)
			}
		})
	}
}

func check(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func mockHeader(add string) (header http.Header, err error) {
	// source: https://blog.questionable.services/article/testing-http-handlers-go/
	req, err := http.NewRequest("GET", "/mock-header", nil)
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
		return header, fmt.Errorf("invalid add=%q argument", add)
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

func TestProductionsAPIv1_Authors(t *testing.T) {
	tests := []struct {
		name string
		p    ProductionsAPIv1
		want Authors
	}{
		{"empty", ProductionsAPIv1{}, Authors{}},
		{"record 1", example1, Authors{nil, []string{"Ile"}, []string{"Ile"}, nil}},
		{"record 2", example2, Authors{nil, []string{"Deep Freeze"}, []string{"The Cardinal"}, nil}},
		{"nick is_group", example3, Authors{nil, nil, nil, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Authors(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductionsAPIv1.Authors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductionsAPIv1_DownloadLink(t *testing.T) {
	tests := []struct {
		name     string
		p        ProductionsAPIv1
		wantName string
		wantLink string
	}{
		{"empty", ProductionsAPIv1{}, "", ""},
		{"record 1", example1, "feestje.zip", "https://files.scene.org/get:nl-http/parties/2000/ambience00/demo/feestje.zip"},
		{"record 2", example2, "the_untouchables_bbs7.zip", "http://www.sensenstahl.com/untergrund_mirror/bbs/the_untouchables_bbs7.zip"},
		{"record 3", example3, "x-wing_cracktro.zip", "http://www.sensenstahl.com/untergrund_mirror/cracktro/x-wing_cracktro.zip"},
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

func TestProductionsAPIv1_Downloads(t *testing.T) {
	prod := Production{
		ID: 1,
	}
	data, err := prod.data()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name string
		p    ProductionsAPIv1
	}{
		{"empty", ProductionsAPIv1{}},
		{"record 1", data},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.Downloads()
		})
	}
}

func TestProductionsAPIv1_Groups(t *testing.T) {
	prod := Production{
		ID: 1,
	}
	data, err := prod.data()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name string
		p    ProductionsAPIv1
		want [2]string
	}{
		{"empty", ProductionsAPIv1{}, [2]string{}},
		{"record 1", data, [2]string{"Aardbei", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Groups(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductionsAPIv1.Groups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductionsAPIv1_PouetID(t *testing.T) {
	prod := Production{
		ID: 1,
	}
	data, err := prod.data()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name           string
		p              ProductionsAPIv1
		ping           bool
		wantId         int
		wantStatusCode int
		wantErr        bool
	}{
		{"empty", ProductionsAPIv1{}, false, 0, 0, false},
		{"record 1", data, true, 7084, 200, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, gotStatusCode, err := tt.p.PouetID(tt.ping)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductionsAPIv1.PouetID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("ProductionsAPIv1.PouetID() gotId = %v, want %v", gotId, tt.wantId)
			}
			if gotStatusCode != tt.wantStatusCode {
				t.Errorf("ProductionsAPIv1.PouetID() gotStatusCode = %v, want %v", gotStatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestProductionsAPIv1_Print(t *testing.T) {
	prod := Production{
		ID: 1,
	}
	data, err := prod.data()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		p       ProductionsAPIv1
		wantErr bool
	}{
		{"empty", ProductionsAPIv1{}, false},
		{"record 1", data, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Print(); (err != nil) != tt.wantErr {
				t.Errorf("ProductionsAPIv1.Print() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
