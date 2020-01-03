package groups

import (
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
			if got := sqlGroupsWhere(tt.args.name, tt.args.incSoftDeletes); got != tt.want {
				t.Errorf("sqlGroupsWhere() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_sqlGroups(t *testing.T) {
	type args struct {
		name               string
		includeSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"all-", args{"all", false}, "(SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 AND `deletedat` IS NULL) UNION (SELECT DISTINCT group_brand_by AS pubCombined FROM files WHERE Length(group_brand_by) <> 0 AND `deletedat` IS NULL) ORDER BY pubCombined"},
		{"all+", args{"all", true}, "(SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 ) UNION (SELECT DISTINCT group_brand_by AS pubCombined FROM files WHERE Length(group_brand_by) <> 0 ) ORDER BY pubCombined"},
		{"ftp-", args{"ftp", false}, "SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 AND RIGHT(group_brand_for,4) = ' FTP' AND `deletedat` IS NULL) ORDER BY pubCombined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqlGroups(tt.args.name, tt.args.includeSoftDeletes); got != tt.want {
				t.Errorf("sqlGroups() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkGroupsToHTML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HTML("", Request{"", true, true, true})
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

func Test_removeInitialism(t *testing.T) {
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
			if got := removeInitialism(tt.args.s); got != tt.want {
				t.Errorf("removeInitialism() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFixSpaces(t *testing.T) {
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
			if got := FixSpaces(tt.args.s); got != tt.want {
				t.Errorf("FixSpaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_progressPct(t *testing.T) {
	type args struct {
		name  string
		count int
		total int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"", args{"", 1, 10}, float64(10)},
		{"", args{"", 10, 10}, float64(100)},
		{"", args{"", 0, 10}, float64(0)},
		{"", args{"", -1, 10}, float64(-10)},
		{"", args{"", 1, 99999}, float64(0.001000010000100001)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := progressPct(tt.args.name, tt.args.count, tt.args.total); got != tt.want {
				t.Errorf("progressPct() = %v, want %v", got, tt.want)
			}
		})
	}
}
