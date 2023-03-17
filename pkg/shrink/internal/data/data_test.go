package data_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/shrink/internal/data"
)

const (
	path = "../../../../tests/empty"
	imgs = "../../../../tests/images"
	uuid = "29e0ca1f-c0a6-4b1a-b019-94a54243c093"
)

func TestMonth(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want data.Months
	}{
		{"empty", "", 0},
		{"two letters", "ja", 0},
		{"three letters", "jan", 1},
		{"last", "dec", 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := data.Month(tt.s); got != tt.want {
				t.Errorf("Month() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	ds := string(filepath.Separator)
	d, err := filepath.Abs(ds)
	if err != nil {
		t.Error(err)
	}
	if err := data.Init(nil, ""); err == nil {
		t.Errorf("Init() should have an error: %v", err)
	}
	if err := data.Init(nil, d); err != nil {
		t.Errorf("Init() should have an error: %v", err)
	}
}

func TestSaveDir(t *testing.T) {
	if s := data.SaveDir(); s == "" {
		t.Errorf("SaveDir() is empty")
	}
}

func TestApprovals_Approve(t *testing.T) {
	tests := []struct {
		name    string
		cmd     data.Approvals
		wantErr bool
	}{
		{"empty", "", true},
		{"bad", "invalid", true},
		{"okay", data.Incoming, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cmd.Approve(nil); !errors.Is(err, data.ErrApprove) != tt.wantErr {
				t.Errorf("Approvals.Approve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func touch() error {
	file, err := os.Create(filepath.Join(path, uuid))
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func TestApprovals_Store(t *testing.T) {
	d, err := filepath.Abs(path)
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Error(err)
		}
		defer os.Remove(d)
	}
	if err := touch(); err != nil {
		t.Error(err)
	}
	type args struct {
		path    string
		partial string
	}
	tests := []struct {
		name    string
		cmd     data.Approvals
		args    args
		wantErr bool
	}{
		{"empty", "", args{}, true},
		{"no args", data.Incoming, args{}, true},
		{"", data.Incoming, args{
			path: d, partial: "",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cmd.Store(nil, tt.args.path, tt.args.partial); (err != nil) != tt.wantErr {
				t.Errorf("Approvals.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompress(t *testing.T) {
	// archive file
	d, err := filepath.Abs(path)
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Error(err)
		}
		defer os.Remove(d)
	}
	tgz := filepath.Join(d, uuid)
	// use images as the content of the archive
	img, err := filepath.Abs(imgs)
	if err != nil {
		t.Error(err)
	}
	i1 := filepath.Join(img, "test_0x.png")
	i2 := filepath.Join(img, "test-clone.png")
	type args struct {
		name  string
		files []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"nofiles", args{
			name: tgz,
		}, false},
		{"okay", args{
			name:  tgz,
			files: []string{i1, i2},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := data.Compress(io.Discard, tt.args.files, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer os.Remove(tgz)
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		wantErr bool
	}{
		{"no files", nil, false},
		{"bad dir", []string{"/invalid"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := data.Remove(nil, tt.files); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
