package groups

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_sqlGroupsWhere(t *testing.T) {
	type args struct {
		name           string
		incSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"mag-", args{"magazine", false}, "AND section = 'magazine' AND `deletedat` IS NULL"},
		{"bbs-", args{"bbs", false}, "AND RIGHT(group_brand_for,4) = ' BBS' AND `deletedat` IS NULL"},
		{"ftp-", args{"ftp", false}, "AND RIGHT(group_brand_for,4) = ' FTP' AND `deletedat` IS NULL"},
		{"grp-", args{"group", false}, "AND RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND `deletedat` IS NULL"},
		{"mag+", args{"magazine", true}, "AND section = 'magazine'"},
		{"bbs+", args{"bbs", true}, "AND RIGHT(group_brand_for,4) = ' BBS'"},
		{"ftp+", args{"ftp", true}, "AND RIGHT(group_brand_for,4) = ' FTP'"},
		{"grp+", args{"group", true}, "AND RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := sqlGroupsWhere(tt.args.name, tt.args.incSoftDeletes); got != tt.want {
				t.Errorf("sqlGroupsWhere() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkGroupsToHTML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Request{"", true, true, true}.HTML("")
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

func Test_initialism(t *testing.T) {
	tests := []struct {
		name    string
		wantNew string
	}{
		{"", ""},
		{"Defacto2", "DF2"},
		{"not found", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNew := initialism(tt.name); gotNew != tt.wantNew {
				t.Errorf("initialism() = %v, want %v", gotNew, tt.wantNew)
			}
		})
	}
}

func Test_remInitialism(t *testing.T) {
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

func Test_remDupeSpaces(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"abc"}, "abc"},
		{"", args{"a b c"}, "a b c"},
		{"", args{"a  b  c"}, "a b c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remDupeSpaces(tt.args.s); got != tt.want {
				t.Errorf("remDupeSpaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dropThe(t *testing.T) {
	type args struct {
		g string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"The X BBS"}, "X BBS"},
		{"", args{"The X FTP"}, "X FTP"},
		{"", args{"the X BBS"}, "X BBS"},
		{"", args{"THE X BBS"}, "X BBS"},
		{"", args{"The"}, "The"},
		{"", args{"Hello BBS"}, "Hello BBS"},
		{"", args{"The High & Mighty Hello BBS"}, "High & Mighty Hello BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropThe(tt.args.g); got != tt.want {
				t.Errorf("dropThe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariations(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"0", args{""}, []string(nil)},
		{"1", args{"hello"}, []string{"hello"}},
		{"2", args{"hello world"}, []string{"hello world", "helloworld", "hello-world", "hello_world", "hello.world"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Variations(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Variations() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_dropDot(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, ""},
		{"no dots", args{"hello"}, "hello"},
		{"dots", args{"hello."}, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropDot(tt.args.s); got != tt.want {
				t.Errorf("dropDot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToClean(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"the blah"}, "The Blah"},
		{"", args{"in the blah"}, "In the Blah"},
		{"", args{"TheBlah"}, "Theblah"},
		{"", args{"MiRROR now"}, "Mirror Now"},
		{"", args{"In the row now ii"}, "In the Row Now II"},
		{"", args{"MiRROR now bbS"}, "Mirror Now BBS"},
		{"", args{"this-is-a-slug-string"}, "This-Is-A-Slug-String"},
		{"", args{"Group inc.,RAZOR TO 1911"}, "Group Inc,Razor to 1911"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToClean(tt.args.s); got != tt.want {
				t.Errorf("ToClean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_list(t *testing.T) {
	tests := []struct {
		name      string
		wantTotal int
		wantErr   bool
	}{
		{"bbs", 3000, false},
		{"ftp", 400, false},
		{"magazine", 100, false},
		{"group", 2000, false},
		{"", 5000, false},
		{"error", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGroups, gotTotal, err := list(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("list() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotGroups) < tt.wantTotal {
				t.Errorf("list() gotGroups count = %v, want >= %v", len(gotGroups), tt.wantTotal)
			}
			if gotTotal < tt.wantTotal {
				t.Errorf("list() gotTotal = %v, want >= %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func Test_sqlGroups(t *testing.T) {
	type args struct {
		filter             string
		includeSoftDeletes bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid soft", args{"bbs", true}, false},
		{"valid", args{"bbs", false}, false},
		{"invalid", args{"invalid filter", false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sqlGroups(tt.args.filter, tt.args.includeSoftDeletes)
			if (err != nil) != tt.wantErr {
				t.Errorf("sqlGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name      string
		wantCount int
	}{
		{"", 0},
		{"abcdefgh", 0},
		{"Defacto2", 28},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCount := Count(tt.name); gotCount != tt.wantCount {
				t.Errorf("Count() = %v, want %v", gotCount, tt.wantCount)
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

func TestRequest_files(t *testing.T) {
	type args struct {
		group string
	}
	tests := []struct {
		name      string
		r         Request
		args      args
		wantTotal int
	}{
		{"empty", Request{}, args{""}, 0},
		{"none", Request{Counts: false}, args{"Defacto2"}, 0},
		{"Defacto2", Request{Counts: true}, args{"Defacto2"}, 28},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTotal := tt.r.files(tt.args.group); gotTotal != tt.wantTotal {
				t.Errorf("Request.files() = %v, want %v", gotTotal, tt.wantTotal)
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
	}{
		{"empty", Request{}, args{""}, ""},
		{"none", Request{Initialisms: false}, args{"Defacto2"}, ""},
		{"Defacto2", Request{Initialisms: true}, args{"Defacto2"}, "DF2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotName := tt.r.initialism(tt.args.group); gotName != tt.wantName {
				t.Errorf("Request.initialism() = %v, want %v", gotName, tt.wantName)
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

func TestPrint(t *testing.T) {
	tests := []struct {
		name      string
		r         Request
		wantTotal int
	}{
		{"", Request{}, 6000},
		{"bbs", Request{Filter: "bbs"}, 3000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTotal := Print(tt.r); gotTotal < tt.wantTotal {
				t.Errorf("Print() = %v, want > %v", gotTotal, tt.wantTotal)
			}
		})
	}
}
