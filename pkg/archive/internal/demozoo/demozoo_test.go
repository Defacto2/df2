package demozoo_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/Defacto2/df2/pkg/archive/internal/demozoo"
)

func TestData(t *testing.T) {
	type fields struct {
		DOSee string
		NFO   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"0", fields{"", ""}, "using \"\" for DOSee and \"\" as the NFO or text"},
		{"1a", fields{"hi.exe", ""}, "using \"hi.exe\" for DOSee and \"\" as the NFO or text"},
		{"1b", fields{"", "hi.txt"}, "using \"\" for DOSee and \"hi.txt\" as the NFO or text"},
		{"2", fields{"hi.exe", "hi.txt"}, "using \"hi.exe\" for DOSee and \"hi.txt\" as the NFO or text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := demozoo.Data{
				DOSee: tt.fields.DOSee,
				NFO:   tt.fields.NFO,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("Data.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTop(t *testing.T) {
	tests := []struct {
		name string
		f    demozoo.Finds
		want string
	}{
		{"empty", demozoo.Finds{}, ""},
		{"1", demozoo.Finds{"file.exe": 0}, "file.exe"},
		{"2", demozoo.Finds{"file.exe": 0, "file.bat": 9}, "file.exe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Top(); got != tt.want {
				t.Errorf("Top() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDOS(t *testing.T) {
	const (
		bat = ".bat"
		com = ".com"
		exe = ".exe"
	)
	type args struct {
		name  string
		files content.Contents
	}
	var empty content.Contents
	var e []string
	f1 := content.File{
		Ext:        com,
		Name:       "hi.com",
		Executable: true,
	}
	f2 := content.File{
		Ext:        exe,
		Name:       "hi.exe",
		Executable: true,
	}
	f3 := content.File{
		Ext:        exe,
		Name:       "random.exe",
		Executable: true,
	}
	ff1 := make(content.Contents)
	ff1[0] = f1
	ff2 := make(content.Contents)
	ff2[0] = f1
	ff2[1] = f2
	ff3 := make(content.Contents)
	ff3[0] = f1
	ff3[1] = f2
	ff3[2] = f3
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", empty}, ""},
		{"empty zip", args{"hi.zip", empty}, ""},
		{"1 file", args{"hi.zip", ff1}, "hi.com"},
		{"2 file", args{"hi.zip", ff2}, "hi.exe"},
		{"3 file", args{"hi.zip", ff2}, "hi.exe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := demozoo.DOS(tt.args.name, tt.args.files, &e); got != tt.want {
				t.Errorf("DOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NFO(t *testing.T) {
	type args struct {
		name  string
		files content.Contents
	}
	var empty content.Contents
	var e []string
	f1 := content.File{
		Ext:      ".diz",
		Name:     "file_id.diz",
		Textfile: true,
	}
	f2 := content.File{
		Ext:      ".nfo",
		Name:     "hi.nfo",
		Textfile: true,
	}
	f3 := content.File{
		Ext:      ".txt",
		Name:     "random.txt",
		Textfile: true,
	}
	ff1 := make(content.Contents)
	ff1[0] = f1
	ff2 := make(content.Contents)
	ff2[0] = f1
	ff2[1] = f2
	ff3 := make(content.Contents)
	ff3[0] = f1
	ff3[1] = f2
	ff3[2] = f3
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", empty}, ""},
		{"empty zip", args{"hi.zip", empty}, ""},
		{"1 file", args{"hi.zip", ff1}, "file_id.diz"},
		{"2 file", args{"hi.zip", ff2}, "hi.nfo"},
		{"3 file", args{"hi.zip", ff2}, "hi.nfo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := demozoo.NFO(tt.args.name, tt.args.files, &e); got != tt.want {
				t.Errorf("NFO() = %v, want %v", got, tt.want)
			}
		})
	}
}
