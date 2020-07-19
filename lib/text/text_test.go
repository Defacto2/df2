package text

import (
	"os"
	"path"
	"testing"
)

func Test_makePng(t *testing.T) {
	wd, _ := os.Getwd()
	src := path.Join(wd, "../../tests/text/test.txt")
	dst := path.Join(wd, "../../tests/text/test")
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
		{"missing src", args{"", dst}, true},
		{"missing dst", args{src, ""}, true},
		{"invalid src", args{src + "invalidate", dst}, true},
		{"text", args{src, dst}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := makePng(tt.args.src, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("makePng() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_generate(t *testing.T) {
	type args struct {
		name string
		id   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test", args{"", ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generate(tt.args.name, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
