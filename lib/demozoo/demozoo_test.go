package demozoo_test

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/demozoo/internal/prods"
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
			f, err := demozoo.Fetch(tt.id)
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
