package assets

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/lib/archive"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
)

func init() {
	color.Enable = false
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
		{"empty", args{}, true},
		{"good", args{"DOWNLOAD", false, false}, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Clean(tt.args.t, tt.args.delete, tt.args.human); (err != nil) != tt.wantErr {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, gotUuids, err := CreateUUIDMap()
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

func Test_backup(t *testing.T) {
	_, dir, err := createTempDir()
	if err != nil {
		t.Error(err)
	}
	_, uuids, err := CreateUUIDMap()
	if err != nil {
		t.Error(err)
	}
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Error(err)
	}
	s := scan{
		dir,
		false,
		true,
		uuids,
	}
	type args struct {
		s    *scan
		list []os.FileInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"ok", args{&s, list}, false},
	}
	d.Backup = os.TempDir() // overwrite /opt/assets/backups
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := backup(tt.args.s, tt.args.list); (err != nil) != tt.wantErr {
				t.Errorf("backup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Error(err)
	}
}

func Test_clean(t *testing.T) {
	type args struct {
		t      Target
		delete bool
		human  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"bad", args{-1, false, false}, true},
		{"empty", args{}, false},
		{"good", args{Download, false, false}, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := clean(tt.args.t, tt.args.delete, tt.args.human); (err != nil) != tt.wantErr {
				t.Errorf("clean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// createTempDir creates a temporary directory and copies some test images into it.
// dir is the absolute directory path while sum is the sum total of bytes copied.
func createTempDir() (sum int64, dir string, err error) {
	// make dir
	dir, err = ioutil.TempDir(os.TempDir(), "test-addtardir")
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
			sum, err = archive.FileCopy(filepath.Join(src, f), filepath.Join(dir, f))
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

func Test_ignoreList(t *testing.T) {
	var want struct{}
	if got := ignoreList("")["blank.png"]; !reflect.DeepEqual(got, want) {
		t.Errorf("ignoreList() = %v, want %v", got, want)
	}
}

func Test_targets(t *testing.T) {
	tests := []struct {
		name   string
		target Target
		want   int
	}{
		{"", All, 6},
		{"", Image, 3},
		{"error", -1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := targets(tt.target); len(got) != tt.want {
				t.Errorf("targets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_walkName(t *testing.T) {
	const file = "file.temp"
	dir := os.TempDir()
	valid, err := walkName(dir, filepath.Join(dir, file))
	if err != nil {
		t.Error(err)
	}
	type args struct {
		basepath string
		path     string
	}
	tests := []struct {
		name     string
		args     args
		wantName string
		wantErr  bool
	}{
		{"empty", args{}, "", true},
		{"empty path", args{dir, ""}, "", true},
		{"empty dir", args{"", file}, file, false},
		{"dir", args{dir, file}, "", true},
		{"dir", args{dir, filepath.Join(dir, file)}, valid, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := walkName(tt.args.basepath, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("walkName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("walkName() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}
