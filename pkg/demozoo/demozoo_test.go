package demozoo_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/demozoo"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/gookit/color"
)

func TestRequest_Query(t *testing.T) {
	r := demozoo.Request{
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

func TestGetProduct(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		wantCode   int
		wantStatus string
		wantTitle  string
	}{
		{"invalid", 0, 404, "404 Not Found", ""},
		{"deleted", 9609, 404, "404 Not Found", ""},
		{"record 1", 1, 200, "200 OK", "Rob Is Jarig"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f demozoo.Product
			err := f.Get(tt.id)
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

func TestGetReleaser(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		wantCode   int
		wantStatus string
		wantName   string
	}{
		{"invalid", 0, 404, "404 Not Found", ""},
		{"releaser #1", 1, 200, "200 OK", "Aardbei"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f demozoo.Releaser
			err := f.Get(tt.id)
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
			if gotAPI.Name != tt.wantName {
				t.Errorf("Fetch() gotTitle = %v, want %v", gotAPI.Name, tt.wantName)
			}
		})
	}
}

func TestGetReleases(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		wantProds bool
		wantErr   bool
	}{
		{"invalid", 0, false, true},
		{"releaser #1", 1, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f demozoo.ReleaserProducts
			gotErr := f.Get(tt.id)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ReleaserProducts() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if prods := (len(f.API) > 0); prods != tt.wantProds {
				t.Errorf("ReleaserProducts.Get() wantProds = %v", prods)
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
		[]byte("6b447ced6d6f919a4b18a8b850442862908cd3eb35cfe1fc01c01b5" +
			"aea6b25c53414fcbba989460b5423b6a29a429078"), // hash strong
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
			gotR, gotErr := demozoo.NewRecord(tt.args.c, tt.args.values)
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
		api       prods.ProductionsAPIv1
		st        demozoo.Stat
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
			r := &demozoo.Record{
				UUID: tt.fields.UUID,
			}
			if gotSkip := r.Download(tt.args.overwrite, &tt.args.api, tt.args.st); gotSkip != tt.wantSkip {
				t.Errorf("Record.download() = %v, want %v", gotSkip, tt.wantSkip)
			}
		})
	}
}

func TestRecord_doseeMeta_fileMeta(t *testing.T) {
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
			r := &demozoo.Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
			}
			if err := r.DoseeMeta(); (err != nil) != tt.wantErr {
				t.Errorf("Record.doseeMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &demozoo.Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
			}
			if err := r.FileMeta(); (err != nil) != tt.wantErr {
				t.Errorf("Record.fileMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type fields struct {
	count          int
	FilePath       string
	ID             string
	UUID           string
	WebIDDemozoo   uint
	WebIDPouet     uint
	Filename       string
	Filesize       string
	FileZipContent string
	CreatedAt      string
	UpdatedAt      string
	SumMD5         string
	Sum384         string
	LastMod        time.Time
	Readme         string
	DOSeeBinary    string
	Platform       string
	GroupFor       string
	GroupBy        string
	Title          string
	Section        string
	CreditText     []string
	CreditCode     []string
	CreditArt      []string
	CreditAudio    []string
}

func TestSQL(t *testing.T) { //nolint:funlen
	const where string = " WHERE id=?"
	now := time.Now()
	tests := []struct {
		name   string
		fields fields
		want   string
		want1  int
	}{
		{name: "empty", fields: fields{}, want: "", want1: 0},
		{
			"filename",
			fields{ID: "1", Filename: "hi.txt"},
			"UPDATE files SET filename=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5,
		},
		{
			"filesize",
			fields{ID: "1", Filesize: "54321"},
			"UPDATE files SET filesize=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5,
		},
		{
			"zip content",
			fields{ID: "1", FileZipContent: "HI.TXT\nHI.EXE"},
			"UPDATE files SET file_zip_content=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5,
		},
		{
			"md5",
			fields{ID: "1", SumMD5: "md5placeholder"},
			"UPDATE files SET file_integrity_weak=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5,
		},
		{
			"sha386",
			fields{ID: "1", Sum384: "shaplaceholder"},
			"UPDATE files SET file_integrity_strong=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5,
		},
		{
			"lastmod",
			fields{ID: "1", LastMod: now},
			"UPDATE files SET file_last_modified=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5,
		},
		{
			"a file",
			fields{ID: "1", Filename: "some.gif", Filesize: "5012352"},
			"UPDATE files SET filename=?,filesize=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := demozoo.Record{
				Count:          tt.fields.count,
				FilePath:       tt.fields.FilePath,
				ID:             tt.fields.ID,
				UUID:           tt.fields.UUID,
				WebIDDemozoo:   tt.fields.WebIDDemozoo,
				WebIDPouet:     tt.fields.WebIDPouet,
				Filename:       tt.fields.Filename,
				Filesize:       tt.fields.Filesize,
				FileZipContent: tt.fields.FileZipContent,
				CreatedAt:      tt.fields.CreatedAt,
				UpdatedAt:      tt.fields.UpdatedAt,
				SumMD5:         tt.fields.SumMD5,
				Sum384:         tt.fields.Sum384,
				LastMod:        tt.fields.LastMod,
				Readme:         tt.fields.Readme,
				DOSeeBinary:    tt.fields.DOSeeBinary,
				Platform:       tt.fields.Platform,
				GroupFor:       tt.fields.GroupFor,
				GroupBy:        tt.fields.GroupBy,
			}
			got, got1 := r.Stmt()
			if got != tt.want {
				t.Errorf("Stmt() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(len(got1), tt.want1) {
				t.Errorf("Stmt() got1 = %v, want %v", len(got1), tt.want1)
			}
		})
	}
}

func TestZipContent(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	pwd = filepath.Join(pwd, "..", "..")
	tests := []struct {
		name    string
		fields  fields
		wantOk  bool
		wantErr bool
	}{
		{"empty", fields{}, false, true},
		{"missing", fields{FilePath: "/dev/null"}, false, true},
		{"dir", fields{FilePath: "tests/demozoo"}, false, true},
		{"7zip", fields{
			FilePath: filepath.Join(pwd, "tests", "demozoo", "test.7z"),
			Filename: "test.7z",
		}, false, true},
		{"zip", fields{
			FilePath: filepath.Join(pwd, "tests", "demozoo", "test.zip"),
			Filename: "test.zip",
		}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &demozoo.Record{
				FilePath: tt.fields.FilePath,
				Filename: tt.fields.Filename,
			}
			gotOk, err := r.ZipContent()
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipContent()  error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ZipContent() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestFileExist(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	pwd = filepath.Join(pwd, "..", "..")
	type fields struct {
		count   int
		missing int
		total   int
	}
	r := demozoo.Record{}
	tests := []struct {
		name        string
		fields      fields
		path        string
		wantMissing bool
	}{
		{name: "empty", path: "", wantMissing: true},
		{name: "missing", path: "/this/dir/does/not/exist", wantMissing: true},
		{name: "7z", path: filepath.Join(pwd, "tests", "demozoo", "test.7z"), wantMissing: false},
		{name: "zip", path: filepath.Join(pwd, "tests", "demozoo", "test.zip"), wantMissing: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &demozoo.Stat{
				Count:   tt.fields.count,
				Missing: tt.fields.missing,
				Total:   tt.fields.total,
			}
			r.FilePath = tt.path
			if gotMissing := st.FileExist(&r); gotMissing != tt.wantMissing {
				t.Errorf("FileExist() = %v, want %v", gotMissing, tt.wantMissing)
			}
		})
	}
}

func TestRecord_String(t *testing.T) {
	color.Enable = false
	type fields struct {
		count        int
		ID           string
		WebIDDemozoo uint
		CreatedAt    string
	}
	type args struct {
		total int
	}
	f := fields{
		count:        5,
		ID:           "99",
		WebIDDemozoo: 77,
		CreatedAt:    "?",
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"default", f, args{}, "→ 0005. 99 (77) ?"},
		{"one", f, args{total: 1}, "→ 5. 99 (77) ?"},
		{"eight", f, args{total: 12345678}, "→ 00000005. 99 (77) ?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := demozoo.Record{
				Count:        tt.fields.count,
				ID:           tt.fields.ID,
				WebIDDemozoo: tt.fields.WebIDDemozoo,
				CreatedAt:    tt.fields.CreatedAt,
			}
			if got := r.String(tt.args.total); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
