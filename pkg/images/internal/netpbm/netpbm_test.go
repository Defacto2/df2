package netpbm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/images/internal/netpbm"
)

func TestConvert(t *testing.T) {
	t.Parallel()
	gif := filepath.Join("..", "..", "..", "..", "testdata", "images", "test.gif")
	iff := filepath.Join("..", "..", "..", "..", "testdata", "images", "test.iff")
	dest := filepath.Join(os.TempDir(), "test_netpbm.png")
	t.Cleanup(func() {
		os.Remove(dest)
	})

	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", ""}, true},
		{"not found", args{"abcde", "ghijk"}, true},
		{"gif", args{gif, dest}, false},
		{"iff", args{iff, dest}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := netpbm.Convert(nil, tt.args.src, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src     string
		want    netpbm.Program
		wantErr bool
	}{
		{"", "", true},
		{"somefile", "", true},
		{"image.png", "", true},
		{"image.iff", netpbm.Ilbm, false},
		{"some.image.pic", netpbm.Ilbm, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.src, func(t *testing.T) {
			t.Parallel()
			got, err := netpbm.ID(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("ID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ID() = %v, want %v", got, tt.want)
			}
		})
	}
}
