package assets_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/assets"
	"github.com/Defacto2/df2/pkg/assets/internal/scan"
	"github.com/Defacto2/df2/pkg/directories"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
)

const empty = "empty"

// createTempDir creates a temporary directory and copies some test images into it.
// dir is the absolute directory path while sum is the sum total of bytes copied.
func createTempDir() (sum int64, dir string, err error) {
	// make dir
	dir, err = os.MkdirTemp(os.TempDir(), "test-addtardir")
	if err != nil {
		return 0, "", err
	}
	// copy files
	src, err := filepath.Abs("../../tests/images")
	if err != nil {
		return 0, dir, err
	}
	imgs := []string{"test.gif", "test.png", "test.jpg"}
	done, sum := make(chan error), int64(0)
	for _, f := range imgs {
		go func(f string) {
			sum, err = archive.Copy(filepath.Join(src, f), filepath.Join(dir, f))
			if err != nil {
				done <- err
			}
			done <- nil
		}(f)
	}
	done1, done2, done3 := <-done, <-done, <-done
	if done1 != nil {
		return 0, dir, done1
	}
	if done2 != nil {
		return 0, dir, done2
	}
	if done3 != nil {
		return 0, dir, done3
	}
	return sum, dir, nil
}

func TestClean(t *testing.T) {
	type args struct {
		t      string
		delete bool
		human  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"bad", args{"invalid", false, false}, true},
		{empty, args{}, true},
		{"good", args{"DOWNLOAD", false, false}, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := assets.Clean(tt.args.t, tt.args.delete, tt.args.human); (err != nil) != tt.wantErr {
				t.Errorf("Clean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateUUIDMap(t *testing.T) {
	tests := []struct {
		name      string
		wantTotal bool
		wantUuids bool
		wantErr   bool
	}{
		{"", true, true, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, gotUuids, err := assets.CreateUUIDMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUUIDMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("CreateUUIDMap() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
			if (len(gotUuids) > 0) != tt.wantUuids {
				t.Errorf("CreateUUIDMap() gotUuids = %v, want %v", len(gotUuids), tt.wantUuids)
			}
		})
	}
}

func TestBackup(t *testing.T) {
	_, dir, err := createTempDir()
	if err != nil {
		t.Error(err)
	}
	_, uuids, err := assets.CreateUUIDMap()
	if err != nil {
		t.Error(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Error(err)
	}
	var list []fs.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			t.Error(err)
		}
		list = append(list, info)
	}
	s := scan.Scan{
		Path:   dir,
		Delete: false,
		Human:  true,
		M:      uuids,
	}
	type args struct {
		s    *scan.Scan
		list []os.FileInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{empty, args{}, true},
		{"ok", args{&s, list}, false},
	}
	d := directories.Init(false)
	d.Backup = os.TempDir() // overwrite /opt/assets/backups
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := scan.Backup(tt.args.s, &d, nil, tt.args.list); (err != nil) != tt.wantErr {
				t.Errorf("backup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Error(err)
	}
}

func TestCleaner(t *testing.T) {
	type args struct {
		t      assets.Target
		delete bool
		human  bool
	}
	d := directories.Init(false)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"bad", args{-1, false, false}, true},
		{empty, args{}, false},
		{"good", args{assets.Download, false, false}, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := assets.Cleaner(tt.args.t, &d, tt.args.delete, tt.args.human); (err != nil) != tt.wantErr {
				t.Errorf("Cleaner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIgnoreList(t *testing.T) {
	var want struct{}
	d := directories.Init(false)
	color.Enable = false
	if got := scan.IgnoreList("", &d)["blank.png"]; !reflect.DeepEqual(got, want) {
		t.Errorf("IgnoreList() = %v, want %v", got, want)
	}
}

func TestTargets(t *testing.T) {
	const allTargets = 5
	tests := []struct {
		name   string
		target assets.Target
		want   int
	}{
		{"", assets.All, allTargets},
		{"", assets.Image, 2},
		{"error", -1, 0},
	}
	d := directories.Init(false)
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assets.Targets(tt.target, &d); len(got) != tt.want {
				t.Errorf("Targets() = %v, want %v", got, tt.want)
			}
		})
	}
}
