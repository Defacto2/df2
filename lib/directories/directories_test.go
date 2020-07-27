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
		{"", args{0}, 0},
		{"", args{1}, 1},
		{"", args{10}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := randStringBytes(tt.args.n); len(got) != tt.want {
				t.Errorf("randStringBytes() = %v, want %v", len(got), tt.want)
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
}
