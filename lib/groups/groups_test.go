package groups

import (
	"log"
	"os"
	"path"
	"reflect"
	"testing"
)

func BenchmarkGroupsToHTML(b *testing.B) {
	r := Request{"", true, true, true}
	for i := 0; i < b.N; i++ {
		if err := r.HTML(""); err != nil {
			log.Fatal(err)
		}
	}
}

func TestMakeSlug(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"Defacto2"}, "defacto2"},
		{"", args{"Defacto 2"}, "defacto-2"},
		{"", args{"Defacto 2 (DF2)"}, "defacto-2"},
		{"", args{"Defacto  2"}, "defacto-2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeSlug(tt.args.name); got != tt.want {
				t.Errorf("MakeSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_parse(t *testing.T) {
	type args struct {
		filename string
		templ    string
	}
	tests := []struct {
		name    string
		r       Request
		args    args
		wantErr bool
	}{
		{"empty", Request{}, args{"", ""}, false},
		{"empty", Request{}, args{os.TempDir(), ""}, true},
		{"empty", Request{}, args{"", "invalidTemplate"}, false},
		{"empty", Request{Filter: "bbs"}, args{path.Join(os.TempDir(), "dump.test"), ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.parse(tt.args.filename, tt.args.templ); (err != nil) != tt.wantErr {
				t.Errorf("Request.parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name      string
		wantCount bool
		wantErr   bool
	}{
		{"", false, false},
		{"abcdefgh", false, false},
		{"Defacto2", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, err := Count(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotCount > 0) != tt.wantCount {
				t.Errorf("Count() = %v, want %v", gotCount, tt.wantCount)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name      string
		r         Request
		wantTotal bool
		wantErr   bool
	}{
		{"", Request{}, true, false},
		{"bbs", Request{Filter: "bbs"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, err := Print(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Print() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("Print() = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestRequest_initialism(t *testing.T) {
	type args struct {
		group string
	}
	tests := []struct {
		name     string
		r        Request
		args     args
		wantName string
		wantErr  bool
	}{
		{"empty", Request{}, args{""}, "", false},
		{"none", Request{Initialisms: false}, args{"Defacto2"}, "", false},
		{"Defacto2", Request{Initialisms: true}, args{"Defacto2"}, "DF2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := tt.r.initialism(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.initialism() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("Request.initialism() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestRequest_files(t *testing.T) {
	type args struct {
		group string
	}
	tests := []struct {
		name      string
		r         Request
		args      args
		wantTotal bool
		wantErr   bool
	}{
		{"empty", Request{}, args{""}, false, false},
		{"none", Request{Counts: false}, args{"Defacto2"}, false, false},
		{"Defacto2", Request{Counts: true}, args{"Defacto2"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, err := tt.r.files(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.files() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("Request.files() = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestVariations(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name     string
		args     args
		wantVars []string
		wantErr  bool
	}{
		{"0", args{""}, []string{}, false},
		{"1", args{"hello"}, []string{"hello"}, false},
		{"2", args{"hello world"}, []string{
			"hello world",
			"helloworld", "hello-world", "hello_world", "hello.world",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVars, err := Variations(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Variations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVars, tt.wantVars) {
				t.Errorf("Variations() = %v, want %v", gotVars, tt.wantVars)
			}
		})
	}
}

func TestFix(t *testing.T) {
	tests := []struct {
		name     string
		simulate bool
		wantErr  bool
	}{
		{"sim", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Fix(tt.simulate); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
