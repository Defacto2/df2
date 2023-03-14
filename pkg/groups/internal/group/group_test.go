package group_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/groups/internal/group"
	"github.com/Defacto2/df2/pkg/groups/internal/rename"
)

func TestSQLWhere(t *testing.T) {
	t.Helper()
	type args struct {
		f              group.Filter
		incSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"mag-", args{group.Magazine, false}, "AND section = 'magazine' AND `deletedat` IS NULL"},
		{"bbs-", args{group.BBS, false}, "AND RIGHT(group_brand_for,4) = ' BBS' AND `deletedat` IS NULL"},
		{"ftp-", args{group.FTP, false}, "AND RIGHT(group_brand_for,4) = ' FTP' AND `deletedat` IS NULL"},
		{"grp-", args{group.Group, false}, "AND RIGHT(group_brand_for,4) != ' FTP' AND " +
			"RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND `deletedat` IS NULL"},
		{"mag+", args{group.Magazine, true}, "AND section = 'magazine'"},
		{"bbs+", args{group.BBS, true}, "AND RIGHT(group_brand_for,4) = ' BBS'"},
		{"ftp+", args{group.FTP, true}, "AND RIGHT(group_brand_for,4) = ' FTP'"},
		{"grp+", args{group.Group, true}, "AND RIGHT(group_brand_for,4) != ' FTP' AND " +
			"RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := group.SQLWhere(nil, tt.args.f, tt.args.incSoftDeletes); got != tt.want {
				t.Errorf("SQLWhere() = %q, want %q", got, tt.want)
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
			got, got1 := group.UseHr(tt.args.cap, tt.args.group)
			if got != tt.want {
				t.Errorf("UseHr() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UseHr() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSQLSelect(t *testing.T) {
	type args struct {
		f                  group.Filter
		includeSoftDeletes bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid soft", args{group.BBS, true}, false},
		{"valid", args{group.BBS, false}, false},
		{"invalid", args{group.Get("invalid filter"), false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := group.SQLSelect(nil, tt.args.f, tt.args.includeSoftDeletes)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLSelect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestList(t *testing.T) {
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
			gotGroups, gotTotal, err := group.List(nil, tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(gotGroups) > 0) != tt.wantTotal {
				t.Errorf("List() gotGroups count = %v, want >= %v", len(gotGroups) > 0, tt.wantTotal)
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("List() gotTotal = %v, want >= %v", gotTotal > 0, tt.wantTotal)
			}
		})
	}
}

func TestSlug(t *testing.T) {
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
			if got := group.Slug(tt.args.name); got != tt.want {
				t.Errorf("Slug() = %v, want %v", got, tt.want)
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
			gotCount, err := group.Count(nil, tt.name)
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

func Test_FmtSyntax(t *testing.T) {
	const ok = "hello & world"
	tests := []struct {
		name string
		w    string
		want string
	}{
		{"word", "hello", "hello"},
		{"words", "hello world", "hello world"},
		{"amp", "hello&world", ok},
		{"multiple", "hello&world&example", "hello & world & example"},
		{"prefix", "&&hello&world", ok},
		{"suffix", "hello&world&&&&&", ok},
		{"malformed", "&&&hello&&&world&&&", ok},
		{"malformed 2", "hello&&world", ok},
		{"malformed 3", "hello&&&world", ok},
		{"ok", ok, ok},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rename.FmtSyntax(tt.w); got != tt.want {
				t.Errorf("FmtSyntax() = %v, want %v", got, tt.want)
			}
		})
	}
}
