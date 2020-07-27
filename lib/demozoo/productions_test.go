package demozoo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// TODO: go generate ProductionsAPIv1 as prefetched stored data using different DZ ids.

const modDate = "Wed, 30 Apr 2012 16:29:51 -0500"

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
		want Authors
	}{
		{"empty", ProductionsAPIv1{}, Authors{}},
		{"record 1", data, Authors{nil, []string{"Ile"}, []string{"Ile"}, nil}},
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
	prod := Production{
		ID: 1,
	}
	data, err := prod.data()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name     string
		p        ProductionsAPIv1
		wantName string
		wantLink string
	}{
		{"empty", ProductionsAPIv1{}, "", ""},
		{"record 1", data, "rob_s_birthday__o__by_random_dutch_scener__not_ile_.zip", "https://files.scene.org/get:nl-http/parties/2003/scene_event03/demo/rob_s_birthday__o__by_random_dutch_scener__not_ile_.zip"},
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
