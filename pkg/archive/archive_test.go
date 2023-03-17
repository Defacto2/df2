package archive_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/Defacto2/df2/pkg/archive/internal/demozoo"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/stretchr/testify/assert"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "tests", name)
}

func TestCompress(t *testing.T) {
	err := archive.Compress(nil, nil)
	assert.NotNil(t, err)

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	err = archive.Compress(gw, nil)
	assert.Nil(t, err)

	files := []string{"nothing", "nilfile"}
	err = archive.Compress(gw, files)
	assert.NotNil(t, err)
}

func TestDelete(t *testing.T) {
	err := archive.Delete(nil)
	assert.Nil(t, err)

	files := []string{"nothing", "nilfile"}
	err = archive.Delete(files)
	assert.NotNil(t, err)
}

func TestStore(t *testing.T) {
	err := archive.Store(nil, nil)
	assert.NotNil(t, err)

	var buf bytes.Buffer
	gw := tar.NewWriter(&buf)
	err = archive.Store(gw, nil)
	assert.Nil(t, err)

	files := []string{"nothing", "nilfile"}
	err = archive.Store(gw, files)
	assert.NotNil(t, err)
}

func TestMIME(t *testing.T) {
	f := content.File{}
	err := f.MIME()
	assert.NotNil(t, err)
	f = content.File{
		Ext:  ".txt",
		Path: testDir("text/test.txt"),
	}
	err = f.MIME()
	assert.Nil(t, err)
}

func TestRead(t *testing.T) {
	f, s, err := archive.Read(nil, "", "")
	assert.Len(t, f, 0)
	assert.Equal(t, "", s)
	assert.NotNil(t, err)

	src := testDir("demozoo/test.invalid.ext.rar")
	f, s, err = archive.Read(io.Discard, src, "test.invalid.ext.rar")
	assert.Len(t, f, 2)
	assert.Equal(t, "test.invalid.ext.zip", s)
	assert.Nil(t, err)

	src = testDir("demozoo/test.zip")
	f, s, err = archive.Read(io.Discard, src, "test.zip")
	assert.Len(t, f, 2)
	assert.Equal(t, "test.zip", s)
	assert.Nil(t, err)
}

func TestRestore(t *testing.T) {
	tmp := os.TempDir()
	type args struct {
		source      string
		filename    string
		destination string
	}
	gz := filepath.Join(testDir("demozoo"), "test.tar.gz")
	tar := filepath.Join(testDir("demozoo"), "test.tar")
	z7 := filepath.Join(testDir("demozoo"), "test.7z")
	zip := filepath.Join(testDir("demozoo"), "test.zip")
	res := []string{"test.png", "test.txt"}
	tests := []struct {
		name      string
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{"empty", args{}, nil, true},
		{"err", args{"/dev/null/fake.zip", "fake.zip", tmp}, nil, true},
		{"zip", args{zip, "test.zip", tmp}, res, false},
		{"tar", args{tar, "test.tar", tmp}, res, false},
		{"gz", args{gz, "test.tar.gz", tmp}, res, false},
		{"7z (unsupported)", args{z7, "test.7z", tmp}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := archive.Restore(nil, tt.args.source, tt.args.filename, tt.args.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("Restore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Restore() = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	const uuid = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
	type args struct {
		archive  string
		filename string
		uuid     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"error", args{testDir("demozoo/test.zip"), "test.zip", "xxxx"}, true},
		{"7z", args{testDir("demozoo/test.7z"), "test.7z", uuid}, true},
		{"tar", args{testDir("demozoo/test.tar"), "test.tar", uuid}, false},
		{"tar.gz", args{testDir("demozoo/test.tar.gz"), "test.tar.gz", uuid}, false},
		{"tar.bz2", args{testDir("demozoo/test.tar.bz2"), "test.tar.bz2", uuid}, false},
		{"sz (snappy)", args{testDir("demozoo/test.tar.sz"), "test.tar.sz", uuid}, false},
		{"xz", args{testDir("demozoo/test.tar.xz"), "test.tar.xz", uuid}, false},
		{"zip", args{testDir("demozoo/test.zip"), "test.zip", uuid}, false},
		{"zip.bz2", args{testDir("demozoo/test.bz2.zip"), "test.bz2.zip", uuid}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := archive.Proof(nil, nil, cfg, tt.args.archive, tt.args.filename, tt.args.uuid); (err != nil) != tt.wantErr {
			// 	t.Errorf("Proof(%s) error = %v, wantErr %v", tt.args.archive, err, tt.wantErr)
			// }
		})
	}
}

func TestNFO(t *testing.T) {
	const fileDiz = demozoo.FileDiz
	var empty []string
	const (
		ff2 = "hi.nfo"
		ff3 = "random.txt"
	)
	type args struct {
		name  string
		files []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", empty}, ""},
		{"empty zip", args{"hi.zip", empty}, ""},
		{"1 file", args{"hi.zip", []string{fileDiz}}, fileDiz},
		{"2 files", args{"hi.zip", []string{fileDiz, ff2}}, "hi.nfo"},
		{"3 files", args{"hi.zip", []string{fileDiz, ff2, ff3}}, "hi.nfo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := archive.NFO(tt.args.name, tt.args.files...); got != tt.want {
				t.Errorf("NFO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDemozoo_Decompress(t *testing.T) {
	dz := archive.Demozoo{}
	_, err := dz.Decompress(nil, nil)
	assert.NotNil(t, err)

	cfg := configger.Defaults()
	db, err := database.Connect(cfg)
	assert.Nil(t, err)
	defer db.Close()
	dz = archive.Demozoo{
		Source: testDir("demozoo/test.zip"),
		UUID:   database.TestID,
		Config: cfg,
	}
	_, err = dz.Decompress(db, nil)
	assert.Nil(t, err)
}

func TestExtractor(t *testing.T) {
	err := archive.Extractor("", "", "test.txt", os.TempDir())
	assert.NotNil(t, err)
	tmp, err := os.MkdirTemp(os.TempDir(), "df2-extractor-test")
	assert.Nil(t, err)
	defer os.Remove(tmp)
	err = archive.Extractor(
		"test.zip",
		testDir("demozoo/test.zip"), "", tmp)
	assert.Nil(t, err)

}
