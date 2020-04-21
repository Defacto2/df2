package text

import (
	"os"
	"path"
	"testing"
)

func TestTxt_check(t *testing.T) {
	type fields struct {
		UUID string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"empty", fields{""}, false},
		{"test", fields{"test"}, true},
		{"missingdir", fields{"test"}, false},
	}
	wd, _ := os.Getwd()
	dir.Img000 = path.Join(wd, "../../tests/text/")
	dir.Img150 = path.Join(wd, "../../tests/text/")
	dir.Img400 = path.Join(wd, "../../tests/text/")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "missingdir" {
				dir.Img400 = ""
			}
			x := Txt{
				UUID: tt.fields.UUID,
			}
			if got := x.check(); got != tt.want {
				t.Errorf("Txt.check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxt_ext(t *testing.T) {
	type fields struct {
		Filename string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"empty", fields{""}, false},
		{"no ext", fields{"hello"}, false},
		{"doc", fields{"hello.doc"}, true},
		{"two exts", fields{"hello.world.doc"}, true},
		{"zip", fields{"hello.zip"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := Txt{
				Filename: tt.fields.Filename,
			}
			if got := x.ext(); got != tt.want {
				t.Errorf("Txt.ext() = %v, want %v", got, tt.want)
			}
		})
	}
}