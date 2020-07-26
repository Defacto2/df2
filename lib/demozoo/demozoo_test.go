package demozoo

import (
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestRequest_Query(t *testing.T) {
	r := Request{
		All:       false,
		Overwrite: false,
		Refresh:   false,
		Simulate:  true,
	}
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"empty", "", true},
		{"invalid", "abcde", true},
		{"not demozoo", "1", false},
		{"demozoo by id", "22884", false},
		{"demozoo by uuid", "0d4777a3-181a-4ce4-bcf2-2093b48be83b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.Query(tt.id); (err != nil) != tt.wantErr {
				t.Errorf("Request.Query() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mutateURL(t *testing.T) {
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
			if got := mutateURL(tt.args.u).String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mutateURL() = %v, want %v", got, tt.want)
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
			got, err := parsePouetProduction(tt.args.rawurl)
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

func Test_production(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"large", args{158411}, "https://demozoo.org/api/v1/productions/158411?format=json", false},
		{"small", args{1}, "https://demozoo.org/api/v1/productions/1?format=json", false},
		{"negative", args{-1}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := production(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("production() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("production() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randomName(t *testing.T) {
	rand := func() bool {
		r, err := randomName()
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
			got, err := saveName(tt.args.rawurl)
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

func TestFetch(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		wantCode   int
		wantStatus string
		wantTitle  string
	}{
		{"invalid", 0, 404, "404 Not Found", ""},
		{"record 1", 1, 200, "200 OK", "Rob Is Jarig"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := Fetch(tt.id)
			if err != nil {
				t.Error(err)
			}
			gotCode, gotStatus, gotAPI := f.Code, f.Status, f.API
			if gotCode != tt.wantCode {
				t.Errorf("Fetch() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("Fetch() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}
			if gotAPI.Title != tt.wantTitle {
				t.Errorf("Fetch() gotTitle = %v, want %v", gotAPI.Title, tt.wantTitle)
			}
		})
	}
}

func values() []sql.RawBytes {
	v := []sql.RawBytes{
		[]byte("1"), // id
		[]byte("41224f41-0262-4750-956a-893fd7f0f082"), // uuid
		[]byte(""),             // deletedat
		[]byte(""),             // createdat
		[]byte("somefile.zip"), // filename
		[]byte("123456789"),    // filesize
		[]byte(""),             // demozoo id
		[]byte("some.jpg\nsome.nfo\nfile_id.diz"), // file_zip_content
		[]byte(""),    // updatedat
		[]byte("dos"), // platform
		[]byte("6b447ced6d6f919a4b18a8b850442862908cd3eb35cfe1fc01c01b5aea6b25c53414fcbba989460b5423b6a29a429078"), // hash strong
		[]byte("3327792e5825386498ac00cd960a6b17"), // hash weak
		[]byte(""),                  // pouet id
		[]byte("Test Group"),        // group for
		[]byte("Fake Group"),        // group by
		[]byte("A test production"), // title
		[]byte("releaseadvert"),     // section
		[]byte("Jack,Jane,Jules"),   // art
		[]byte("Sam,Sock"),          // audio
		[]byte("Joe Blogs,Doe"),     // code
		[]byte("Lisa,Linus"),        // text
	}
	return v
}

func Test_newRecord(t *testing.T) {
	type args struct {
		c      int
		values []sql.RawBytes
	}
	short := values()
	short = short[:len(short)-1]
	pouet := values()
	pouet[12] = []byte("50")
	tests := []struct {
		name         string
		args         args
		wantID       string
		wantFilename string
		wantPlatform string
		wantText     []string
		wantPoeut    uint
		wantErr      bool
	}{
		{"empty", args{0, nil}, "", "", "", nil, 0, true},
		{"short", args{0, short}, "", "", "", nil, 0, true},
		{"ok", args{0, values()}, "1", "somefile.zip", "dos", []string{"Lisa", "Linus"}, 0, false},
		{"pouet", args{0, pouet}, "1", "somefile.zip", "dos", []string{"Lisa", "Linus"}, 50, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotErr := newRecord(tt.args.c, tt.args.values)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("newRecord() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if gotR.ID != tt.wantID {
				t.Errorf("newRecord().ID = %v, want %v", gotR.ID, tt.wantID)
			}
			if gotR.Filename != tt.wantFilename {
				t.Errorf("newRecord().Filename = %v, want %v", gotR.Filename, tt.wantFilename)
			}
			if gotR.Platform != tt.wantPlatform {
				t.Errorf("newRecord().Platform = %v, want %v", gotR.Platform, tt.wantPlatform)
			}
			if gotR.WebIDPouet != tt.wantPoeut {
				t.Errorf("newRecord().WebIDPouet = %v, want %v", gotR.WebIDPouet, tt.wantPoeut)
			}
			if !reflect.DeepEqual(gotR.CreditText, tt.wantText) {
				t.Errorf("newRecord().CreditText = %v, want %v", gotR.CreditText, tt.wantText)
			}
		})
	}
}

func TestRecord_download(t *testing.T) {
	type fields struct {
		UUID string
	}
	type args struct {
		overwrite bool
		api       ProductionsAPIv1
		st        stat
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantSkip bool
	}{
		{"empty", fields{}, args{}, true},
		{"okay", fields{UUID: "0d4777a3-181a-4ce4-bcf2-2093b48be83b"}, args{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Record{
				UUID: tt.fields.UUID,
			}
			if gotSkip := r.download(tt.args.overwrite, tt.args.api, tt.args.st); gotSkip != tt.wantSkip {
				t.Errorf("Record.download() = %v, want %v", gotSkip, tt.wantSkip)
			}
		})
	}
}

func TestRecord_doseeMeta(t *testing.T) {
	type fields struct {
		ID   string
		UUID string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty", fields{}, true},
		{"id", fields{ID: "22884"}, true},
		{"uuid", fields{UUID: "0d4777a3-181a-4ce4-bcf2-2093b48be83b"}, true}, // because physical files are missing
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
			}
			if err := r.doseeMeta(); (err != nil) != tt.wantErr {
				t.Errorf("Record.doseeMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestRecord_fileMeta(t *testing.T) {
	type fields struct {
		ID   string
		UUID string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty", fields{}, true},
		{"id", fields{ID: "22884"}, true},
		{"uuid", fields{UUID: "0d4777a3-181a-4ce4-bcf2-2093b48be83b"}, true}, // because physical files are missing
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
			}
			if err := r.fileMeta(); (err != nil) != tt.wantErr {
				t.Errorf("Record.fileMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
