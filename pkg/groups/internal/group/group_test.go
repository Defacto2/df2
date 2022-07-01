package group_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/groups/internal/group"
)

func TestClean(t *testing.T) {
	tests := []struct {
		name   string
		wantOk bool
	}{
		{"", false},
		{"Defacto2", false},
		{"defacto2", true},
		{"d-e-f-a-c-t-o-2", true},
		{"d_f", true},
		{"D2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := group.Clean(tt.name); gotOk != tt.wantOk {
				t.Errorf("Clean() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestCleanS(t *testing.T) {
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
			if got := group.CleanS(tt.args.s); got != tt.want {
				t.Errorf("CleanS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimThe(t *testing.T) {
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
			if got := group.TrimThe(tt.args.g); got != tt.want {
				t.Errorf("TrimThe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimDot(t *testing.T) {
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
			if got := group.TrimDot(tt.args.s); got != tt.want {
				t.Errorf("TrimDot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"", ""},
		{"hello", "Hello"},
		{"hello  world", "Hello  World"},
		{"By THE Way", "By the Way"},
		{"BENS ftp", "Bens FTP"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := group.Format(tt.s); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TesSQLWhere(t *testing.T) {
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
			if got, _ := group.SQLWhere(tt.args.f, tt.args.incSoftDeletes); got != tt.want {
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
			_, err := group.SQLSelect(tt.args.f, tt.args.includeSoftDeletes)
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
			gotGroups, gotTotal, err := group.List(tt.name)
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
			gotCount, err := group.Count(tt.name)
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
