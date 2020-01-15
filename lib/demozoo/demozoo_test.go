package demozoo

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/df2/lib/database"
)

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
		{"scene.org", args{fso}, "https://files.scene.org/get/someplace"},
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

func Test_prodURL(t *testing.T) {
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
			got, err := prodURL(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("prodURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("prodURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randomName(t *testing.T) {
	rand := func() bool {
		r := randomName()
		println(r)
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

func TestRecord_sql(t *testing.T) {
	type fields struct {
		count          int
		AbsFile        string
		ID             string
		UUID           string
		WebIDDemozoo   string
		Filename       string
		Filesize       string
		FileZipContent string
		CreatedAt      string
		UpdatedAt      string
		SumMD5         string
		Sum384         string
		LastMod        time.Time
	}
	const where string = " WHERE id=?"
	var now = time.Now()
	var empty []interface{}
	tests := []struct {
		name   string
		fields fields
		want   string
		want1  []interface{}
	}{
		{"empty", fields{}, "", empty},
		{"filename", fields{ID: "1", Filename: "hi.txt"}, "UPDATE files SET filename=?" + where, []interface{}{"hi.txt", "1"}},
		{"filesize", fields{ID: "1", Filesize: "54321"}, "UPDATE files SET filesize=?" + where, []interface{}{"54321", "1"}},
		{"zip content", fields{ID: "1", FileZipContent: "HI.TXT\nHI.EXE"}, "UPDATE files SET file_zip_content=?" + where, []interface{}{"HI.TXT\nHI.EXE", "1"}},
		{"md5", fields{ID: "1", SumMD5: "md5placeholder"}, "UPDATE files SET file_integrity_weak=?" + where, []interface{}{"md5placeholder", "1"}},
		{"sha386", fields{ID: "1", Sum384: "shaplaceholder"}, "UPDATE files SET file_integrity_strong=?" + where, []interface{}{"shaplaceholder", "1"}},
		{"lastmod", fields{ID: "1", LastMod: now}, "UPDATE files SET file_last_modified=?" + where, []interface{}{now.Format(database.Datetime), "1"}},
		{"a file", fields{ID: "1", Filename: "some.gif", Filesize: "5012352"}, "UPDATE files SET filename=?,filesize=?" + where, []interface{}{"some.gif", "5012352", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Record{
				count:          tt.fields.count,
				AbsFile:        tt.fields.AbsFile,
				ID:             tt.fields.ID,
				UUID:           tt.fields.UUID,
				WebIDDemozoo:   tt.fields.WebIDDemozoo,
				Filename:       tt.fields.Filename,
				Filesize:       tt.fields.Filesize,
				FileZipContent: tt.fields.FileZipContent,
				CreatedAt:      tt.fields.CreatedAt,
				UpdatedAt:      tt.fields.UpdatedAt,
				SumMD5:         tt.fields.SumMD5,
				Sum384:         tt.fields.Sum384,
				LastMod:        tt.fields.LastMod,
			}
			got, got1 := r.sql()
			if got != tt.want {
				t.Errorf("Record.sql() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Record.sql() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
