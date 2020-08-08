package directories

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_randStringBytes(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"zero", args{0}, 0},
		{"one", args{1}, 1},
		{"ten", args{10}, 10},
		{"sixty three", args{63}, 63},
		{"sixty four", args{64}, 64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := randStringBytes(tt.args.n)
			if len(got) != tt.want {
				t.Errorf("randStringBytes() = %v, want %v\n%s", len(got), tt.want, got)
			}
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_createDirectory(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "test-create-dir")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty", "", true},
		{"new", tempDir, false},
		{"temp", os.TempDir(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createDirectory(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("createDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	if err := os.RemoveAll(tempDir); err != nil {
		t.Error(err)
	}
}

func Test_createHolderFiles(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "test-create-holders")
	if err != nil {
		t.Error(err)
	}
	type args struct {
		dir    string
		size   int
		number uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, false},
		{"temp", args{tempDir, 50, 1}, false},
		{"too big", args{tempDir, 50, 100}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createHolderFiles(tt.args.dir, tt.args.size, tt.args.number); (err != nil) != tt.wantErr {
				t.Errorf("createHolderFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := os.RemoveAll(tempDir); err != nil {
		t.Error(err)
	}
	if err := os.Remove("00000000-0000-0000-0000-000000000000"); err != nil {
		t.Error(err)
	}
}
