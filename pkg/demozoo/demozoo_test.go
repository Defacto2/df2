package demozoo_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo"
	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestStat_NextRefresh(t *testing.T) {
	t.Parallel()
	s := demozoo.Stat{}
	err := s.NextRefresh(nil, nil, demozoo.Records{})
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = s.NextRefresh(db, io.Discard, demozoo.Records{})
	assert.NotNil(t, err)
}

func TestStat_NewPouet(t *testing.T) {
	t.Parallel()
	s := demozoo.Stat{}
	err := s.NextPouet(nil, nil, demozoo.Records{})
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = s.NextPouet(db, io.Discard, demozoo.Records{})
	assert.NotNil(t, err)
}

func TestCounter(t *testing.T) {
	t.Parallel()
	err := demozoo.Counter(nil, nil, 0)
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	w := strings.Builder{}
	err = demozoo.Counter(db, &w, 0)
	assert.Nil(t, err)
	assert.Contains(t, w.String(), "records with Demozoo links")
	err = demozoo.Counter(db, &w, 1)
	assert.Nil(t, err)
	assert.Contains(t, w.String(), "records with Pouet links")
	err = demozoo.Counter(db, &w, 1000)
	assert.NotNil(t, err)
}

func TestProduct_Get(t *testing.T) {
	t.Parallel()
	p := demozoo.Product{}
	err := p.Get(0)
	assert.NotNil(t, err)
	err = p.Get(1)
	if !errors.Is(err, context.DeadlineExceeded) {
		assert.Nil(t, err)
		assert.NotEmpty(t, p.API)
	}
}

func TestReleaser_Get(t *testing.T) {
	t.Parallel()
	p := demozoo.Releaser{}
	err := p.Get(0)
	assert.NotNil(t, err)
	err = p.Get(1)
	if !errors.Is(err, context.DeadlineExceeded) {
		assert.Nil(t, err)
		assert.NotEmpty(t, p.API)
	}
}

func TestReleaserProducts_Get(t *testing.T) {
	t.Parallel()
	p := demozoo.ReleaserProducts{}
	err := p.Get(0)
	assert.NotNil(t, err)
	err = p.Get(1)
	if !errors.Is(err, context.DeadlineExceeded) {
		assert.Nil(t, err)
		assert.NotEmpty(t, p.API)
	}
}

func TestMsDosProducts_Get(t *testing.T) {
	t.Parallel()
	p := demozoo.MsDosProducts{}
	err := p.Get(nil, nil)
	assert.NotNil(t, err)
}

func TestWindowsProducts_Get(t *testing.T) {
	t.Parallel()
	p := demozoo.WindowsProducts{}
	err := p.Get(nil, nil)
	assert.NotNil(t, err)
}

func TestFix(t *testing.T) {
	t.Parallel()
	err := demozoo.Fix(nil, nil)
	assert.NotNil(t, err)
}

func TestRequest_Query(t *testing.T) {
	t.Parallel()
	r := demozoo.Request{}
	err := r.Query(nil, nil, "")
	assert.NotNil(t, err)
	l, err := zap.NewProduction()
	assert.Nil(t, err)
	r = demozoo.Request{
		All:       false,
		Overwrite: false,
		Refresh:   false,
		Config:    conf.Defaults(),
		Logger:    l.Sugar(),
	}
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.Query(db, io.Discard, "")
	assert.NotNil(t, err)
	err = r.Query(db, io.Discard, "qwerty")
	assert.NotNil(t, err, "invalid id")
	err = r.Query(db, io.Discard, "1")
	assert.Nil(t, err, "record doesn't have a Demozoo association")
	err = r.Query(db, io.Discard, "22884")
	assert.Nil(t, err, "record has a Demozoo association")
	err = r.Query(db, io.Discard, "0d4777a3-181a-4ce4-bcf2-2093b48be83b")
	assert.Nil(t, err, "record by uuid has a Demozoo association")
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

func TestNewRecord(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

func TestRecord_Download(t *testing.T) {
	t.Parallel()
	r := demozoo.Record{}
	st := demozoo.Stat{}
	err := r.Download(nil, nil, st, false)
	assert.NotNil(t, err)

	api := prods.ProductionsAPIv1{}
	err = r.Download(io.Discard, &api, st, false)
	assert.NotNil(t, err)

	r = demozoo.Record{UUID: "0d4777a3-181a-4ce4-bcf2-2093b48be83b"}
	err = r.Download(io.Discard, &api, st, false)
	assert.NotNil(t, err)
}

func TestRecord_DoseeMeta(t *testing.T) { //nolint:tparallel
	t.Parallel()
	r := demozoo.Record{}
	c := conf.Defaults()
	err := r.DoseeMeta(nil, nil, c)
	assert.NotNil(t, err)
	db, err := database.Connect(c)
	assert.Nil(t, err)
	defer db.Close()
	err = r.DoseeMeta(nil, nil, c)
	assert.NotNil(t, err)

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
	cfg := conf.Defaults()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &demozoo.Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
			}
			if err := r.DoseeMeta(nil, nil, cfg); (err != nil) != tt.wantErr {
				t.Errorf("Record.doseeMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
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
		{"dir", fields{FilePath: "testdata/demozoo"}, false, true},
		{"7zip", fields{
			FilePath: filepath.Join(pwd, "testdata", "demozoo", "test.7z"),
			Filename: "test.7z",
		}, false, true},
		{"zip", fields{
			FilePath: filepath.Join(pwd, "testdata", "demozoo", "test.zip"),
			Filename: "test.zip",
		}, true, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &demozoo.Record{
				FilePath: tt.fields.FilePath,
				Filename: tt.fields.Filename,
			}
			gotOk, err := r.ZipContent(nil)
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
	t.Parallel()
	wd, err := os.Getwd()
	assert.Nil(t, err)
	pwd := filepath.Join(wd, "..", "..")
	st := demozoo.Stat{}
	b, err := st.FileExist(nil)
	assert.NotNil(t, err)
	assert.Equal(t, false, b)

	r := demozoo.Record{}
	b, err = st.FileExist(&r)
	assert.Nil(t, err)
	assert.Equal(t, false, b)
	assert.Equal(t, 1, st.Missing)

	r = demozoo.Record{
		FilePath: "/this/does/not/exist",
	}
	b, err = st.FileExist(&r)
	assert.Nil(t, err)
	assert.Equal(t, false, b)
	assert.Equal(t, 2, st.Missing)

	r = demozoo.Record{
		FilePath: pwd,
	}
	b, err = st.FileExist(&r)
	assert.NotNil(t, err)
	assert.Equal(t, false, b)
	assert.Equal(t, 2, st.Missing)

	r = demozoo.Record{
		FilePath: filepath.Join(pwd, "testdata", "demozoo", "test.7z"),
	}
	b, err = st.FileExist(&r)
	assert.Nil(t, err)
	assert.Equal(t, true, b)
	assert.Equal(t, 2, st.Missing)
}

func TestRecord_String(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := demozoo.Record{
				Count:        tt.fields.count,
				ID:           tt.fields.ID,
				WebIDDemozoo: tt.fields.WebIDDemozoo,
				CreatedAt:    tt.fields.CreatedAt,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
