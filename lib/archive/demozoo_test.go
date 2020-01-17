package archive

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func TestDemozoo_String(t *testing.T) {
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
			d := Demozoo{
				DOSee: tt.fields.DOSee,
				NFO:   tt.fields.NFO,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("Demozoo.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractDemozoo(t *testing.T) {
	type args struct {
		name     string
		uuid     string
		varNames []string
	}
	tests := []struct {
		name    string
		args    args
		want    Demozoo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractDemozoo(tt.args.name, tt.args.uuid, &tt.args.varNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractDemozoo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractDemozoo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_content_filescan(t *testing.T) {
	type fields struct {
		name       string
		file       int
		ext        string
		path       string
		mime       *mimetype.MIME
		modtime    time.Time
		size       int64
		executable bool
		textfile   bool
	}
	type args struct {
		f os.FileInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &content{
				name:       tt.fields.name,
				file:       tt.fields.file,
				ext:        tt.fields.ext,
				path:       tt.fields.path,
				mime:       tt.fields.mime,
				modtime:    tt.fields.modtime,
				size:       tt.fields.size,
				executable: tt.fields.executable,
				textfile:   tt.fields.textfile,
			}
			c.filescan(tt.args.f)
		})
	}
}

func Test_moveText(t *testing.T) {
	type args struct {
		name string
		uuid string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := moveText(tt.args.name, tt.args.uuid); got != tt.want {
				t.Errorf("moveText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_finds_top(t *testing.T) {
	tests := []struct {
		name string
		f    finds
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.top(); got != tt.want {
				t.Errorf("finds.top() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findDOS(t *testing.T) {
	type args struct {
		name  string
		files contents
	}
	var empty contents
	var e []string
	f1 := content{
		ext:        ".com",
		name:       "hi.com",
		executable: true,
	}
	f2 := content{
		ext:        ".exe",
		name:       "hi.exe",
		executable: true,
	}
	f3 := content{
		ext:        ".exe",
		name:       "random.exe",
		executable: true,
	}
	ff1 := make(contents)
	ff1[0] = f1
	ff2 := make(contents)
	ff2[0] = f1
	ff2[1] = f2
	ff3 := make(contents)
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
			if got := findDOS(tt.args.name, tt.args.files, &e); got != tt.want {
				t.Errorf("findDOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findNFO(t *testing.T) {
	type args struct {
		name  string
		files contents
	}
	var empty contents
	var e []string
	f1 := content{
		ext:      ".diz",
		name:     "file_id.diz",
		textfile: true,
	}
	f2 := content{
		ext:      ".nfo",
		name:     "hi.nfo",
		textfile: true,
	}
	f3 := content{
		ext:      ".txt",
		name:     "random.txt",
		textfile: true,
	}
	ff1 := make(contents)
	ff1[0] = f1
	ff2 := make(contents)
	ff2[0] = f1
	ff2[1] = f2
	ff3 := make(contents)
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
			if got := findNFO(tt.args.name, tt.args.files, &e); got != tt.want {
				t.Errorf("findNFO() = %v, want %v", got, tt.want)
			}
		})
	}
}
