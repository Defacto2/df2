package img_test

import (
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/lib/text/internal/img"
)

func TestMakePng(t *testing.T) {
	dst, err := filepath.Abs("../../tests/text/test")
	if err != nil {
		t.Error(err)
	}
	src, err := filepath.Abs("../../tests/text/test.txt")
	if err != nil {
		t.Error(err)
	}
	type args struct {
		src   string
		dest  string
		amiga bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", "", false}, true},
		{"missing src", args{"", dst, false}, true},
		{"missing dst", args{src, "", false}, true},
		{"invalid src", args{src + "invalidate", dst, false}, true},
		{"text", args{src, dst, false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := img.MakePng(tt.args.src, tt.args.dest, tt.args.amiga)
			if (err != nil) != tt.wantErr {
				t.Errorf("makePng() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_generate(t *testing.T) {
	gif, err := filepath.Abs("../../tests/images/test.gif")
	if err != nil {
		t.Error(err)
	}
	txt, err := filepath.Abs("../../tests/text/test.txt")
	if err != nil {
		t.Error(err)
	}
	type args struct {
		name  string
		id    string
		amiga bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test", args{"", "", false}, true},
		{"missing", args{"abce", "1", false}, true},
		{"gif", args{gif, "1", false}, true},
		{"txt", args{txt, "1", false}, true},
		{"amiga", args{txt, "1", true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := img.Generate(tt.args.name, tt.args.id, tt.args.amiga)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
