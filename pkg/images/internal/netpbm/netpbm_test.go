package netpbm_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/images/internal/netpbm"
)

func TestConvert(t *testing.T) {
	const gif = "../../../../tests/images/test.gif"
	const iff = "../../../../tests/images/test.iff"
	dest := filepath.Join(os.TempDir(), "test_netpbm.png")
	fmt.Fprintln(os.Stdout, dest)
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
		t.Run(tt.name, func(t *testing.T) {
			if err := netpbm.Convert(nil, tt.args.src, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer os.Remove(dest)
		})
	}
}

func TestID(t *testing.T) {
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
		t.Run(tt.src, func(t *testing.T) {
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
