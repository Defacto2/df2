package groups

import (
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
	}{
		{"bbs", 3000},
		{"ftp", 400},
		{"magazine", 100},
		{"group", 2000},
		{"", 5000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGroups, gotTotal := list(tt.name)
			if len(gotGroups) <= tt.wantTotal {
				t.Errorf("list() gotGroups count = %v, want >= %v", len(gotGroups), tt.wantTotal)
			}
			if gotTotal <= tt.wantTotal {
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
