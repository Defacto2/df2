package groups

import (
	"log"
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_groupsWhere(t *testing.T) {
	type args struct {
		f              Filter
		incSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"mag-", args{Magazine, false}, "AND section = 'magazine' AND `deletedat` IS NULL"},
		{"bbs-", args{BBS, false}, "AND RIGHT(group_brand_for,4) = ' BBS' AND `deletedat` IS NULL"},
		{"ftp-", args{FTP, false}, "AND RIGHT(group_brand_for,4) = ' FTP' AND `deletedat` IS NULL"},
		{"grp-", args{Group, false}, "AND RIGHT(group_brand_for,4) != ' FTP' AND " +
			"RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND `deletedat` IS NULL"},
		{"mag+", args{Magazine, true}, "AND section = 'magazine'"},
		{"bbs+", args{BBS, true}, "AND RIGHT(group_brand_for,4) = ' BBS'"},
		{"ftp+", args{FTP, true}, "AND RIGHT(group_brand_for,4) = ' FTP'"},
		{"grp+", args{Group, true}, "AND RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := groupsWhere(tt.args.f, tt.args.incSoftDeletes); got != tt.want {
				t.Errorf("groupsWhere() = %q, want %q", got, tt.want)
			}
		})
	}
}

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

func Test_remInitialism(t *testing.T) { //nolint:dupl
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"Defacto (DF)"}, "Defacto"},
		{"", args{"Defacto 2 (DF2)"}, "Defacto 2"},
		{"", args{"Defacto 2"}, "Defacto 2"},
		{"", args{"Razor 1911 (RZR)"}, "Razor 1911"},
		{"", args{"Defacto (2) (DF2)"}, "Defacto (2)"},
		{"", args{"(Defacto 2) (DF2)"}, "(Defacto 2)"},
		{"", args{"Defacto(DF)"}, "Defacto(DF)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remInitialism(tt.args.s); got != tt.want {
				t.Errorf("remInitialism() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_list(t *testing.T) {
	tests := []struct {
		name      string
		wantTotal bool
		wantErr   bool
	}{
		{"bbs", true, false},
		{"ftp", true, false},
		{"magazine", true, false},
		{"group", true, false},
		{"", true, false},
		{"error", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGroups, gotTotal, err := list(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("list() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(gotGroups) > 0) != tt.wantTotal {
				t.Errorf("list() gotGroups count = %v, want >= %v", len(gotGroups) > 0, tt.wantTotal)
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("list() gotTotal = %v, want >= %v", gotTotal > 0, tt.wantTotal)
			}
		})
	}
}

func Test_groupsStmt(t *testing.T) {
	type args struct {
		f                  Filter
		includeSoftDeletes bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid soft", args{BBS, true}, false},
		{"valid", args{BBS, false}, false},
		{"invalid", args{filter("invalid filter"), false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := groupsStmt(tt.args.f, tt.args.includeSoftDeletes)
			if (err != nil) != tt.wantErr {
				t.Errorf("groupsStmt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_hrElement(t *testing.T) {
	type args struct {
		cap   string
		group string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{"empty", args{"", ""}, "", false},
		{"Defacto2", args{"", "Defacto2"}, "D", false},
		{"Defacto2", args{"D", "Defacto2"}, "D", false},
		{"Defacto2", args{"C", "Defacto2"}, "D", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := hrElement(tt.args.cap, tt.args.group)
			if got != tt.want {
				t.Errorf("hrElement() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("hrElement() got1 = %v, want %v", got1, tt.want1)
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
		{"0", args{""}, []string(nil), false},
		{"1", args{"hello"}, []string{"hello"}, false},
		{"2", args{"hello world"}, []string{"hello world", "helloworld", "hello-world", "hello_world", "hello.world"}, false},
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

func Test_initialism(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"Defacto2", "DF2", false},
		{"not found", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initialism(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("initialism() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("initialism() = %v, want %v", got, tt.want)
			}
		})
	}
}
