package assets

import (
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_ignoreList(t *testing.T) {
	var want struct{}
	if got := ignoreList("")["blank.png"]; !reflect.DeepEqual(got, want) {
		t.Errorf("ignoreList() = %v, want %v", got, want)
	}
}

func Test_targets(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   int
	}{
		{"", "all", 6},
		{"", "image", 3},
		{"error", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := targets(tt.target); got != tt.want {
				t.Errorf("targets() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
func TestAddTarFile(t *testing.T) {
	// create dirs
	tempDir, err := ioutil.TempDir(os.TempDir(), "test-addtardir")
	if err != nil {
		t.Error(err)
	}
	tempFile := filepath.Join(os.TempDir(), "test-addtarfile.tar")
	// tempATF, err := ioutil.TempDir(os.TempDir(), "test-addtarfile")
	// if err != nil {
	// 	t.Error(err)
	// }
	// tempFile, err := ioutil.TempFile(tempATF, "test_addtar.tar")
	// if err != nil {
	// 	t.Error(err)
	// }
	// copy files
	src, err := filepath.Abs("../../tests/images")
	if err != nil {
		t.Error(err)
	}
	imgs := []string{"test.gif", "test.png", "test.jpg"}
	done := make(chan error)
	for _, f := range imgs {
		go func(f string) {
			if _, err := archive.FileCopy(filepath.Join(src, f), filepath.Join(tempDir, f)); err != nil {
				done <- err
			}
			done <- nil
		}(f)
	}
	done1, done2, done3 := <-done, <-done, <-done
	if done1 != nil {
		t.Error(done1)
	}
	if done2 != nil {
		t.Error(done2)
	}
	if done3 != nil {
		t.Error(done3)
	}
	// create tar archive
	newTar, err := os.Create(tempFile)
	if err != nil {
		t.Error(err)
	}
	tw := tar.NewWriter(newTar)
	defer tw.Close()
	fmt.Println("src:", src, "temp:", tempDir)
	type args struct {
		path string
		name string
		tw   *tar.Writer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"err path", args{"", tempFile, tw}, true},
		//{"err name", args{tempDir, "", tw}, true},
		//{"err tw", args{tempDir, tempFile.Name(), nil}, true},
		{"ok", args{tempDir, "/test.png", tw}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddTarFile(tt.args.path, tt.args.name, tt.args.tw); (err != nil) != tt.wantErr {
				t.Errorf("AddTarFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name, err := walkName(tempDir, path)
		if err != nil {
			return err
		}
		fmt.Println("---->", name)
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	const size = 4000
	if f, err := os.Stat(tempFile); err != nil {
		t.Error(err)
	} else if s := f.Size(); s != size {
		t.Errorf("AddTarFile() error = invalid file size, got %dB, want %dB for %s", s, size, tempFile)
	}
	// if err := os.RemoveAll(tempDir); err != nil {
	// 	t.Error(err)
	// }
	// if err := os.Remove(tempFile); err != nil {
	// 	t.Error(err)
	// }
}
*/

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
