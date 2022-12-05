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
		{"this is the group,the group is this", true},
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
		{"empty string", args{""}, ""},
		{"leading the", args{"the blah"}, "The Blah"},
		{"common the", args{"in the blah"}, "In the Blah"},
		{"no spaces", args{"TheBlah"}, "Theblah"},
		{"elite fmt", args{"MiRROR now"}, "Mirror Now"},
		{"roman numbers", args{"In the row now ii"}, "In the Row Now II"},
		{"BBS", args{"MiRROR now bbS"}, "Mirror Now BBS"},
		{"slug", args{"this-is-a-slug-string"}, "This-is-a-Slug-String"},
		{
			"pair of groups",
			args{"Group inc.,RAZOR TO 1911"},
			"Group Inc,Razor to 1911",
		},
		{
			"2nd group with a leading the",
			args{"this is the group,the group is this"},
			"This is the Group,The Group is This",
		},
		{"ordinal", args{"4TH dimension"}, "4th Dimension"},
		{"ordinals", args{"4TH dimension, 5Th Dynasty"}, "4th Dimension, 5th Dynasty"},
		{"abbreviation", args{"2000 ad"}, "2000AD"},
		{"abbreviations", args{"2000ad, 500bc"}, "2000AD, 500BC"},
		{
			"mega-group",
			args{"Lightforce,Pact,TRSi,Venom,Razor 1911,the System"},
			"Lightforce,Pact,Trsi,Venom,Razor 1911,The System",
		},
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
		{"the cool group", "The Cool Group"},
		{"cooL THE GROUP", "Cool the Group"},
		{"anz ftp", "ANZ FTP"},
		{"1St ftp", "1st FTP"},
		{"2000ad", "2000AD"},
		{"pc-crew", "PC-Crew"},
		{"inc utils", "INC Utils"},
		{"inc", "INC"},
		{"razor 1911, inc", "Razor 1911, INC"},
		{"raZor 1911, the system", "Razor 1911, The System"},
		{"tristar&red sector inc", "Tristar & Red Sector Inc"},
		{"tristar & red sector inc", "Tristar & Red Sector Inc"},
		{"ab&c, xy&z", "Ab & C, Xy & Z"},
		{"&broken&&&name&&&&", "Broken & Name"},
		{"hello-world", "Hello-World"},
		{"2nd-hello-world", "2nd-Hello-World"},
		{"1st-the-hello-world", "1st-the-Hello-World"},
		{"4-am bbs", "4-AM BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := group.Format(tt.s); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			if got := group.FmtSyntax(tt.w); got != tt.want {
				t.Errorf("FmtSyntax() = %v, want %v", got, tt.want)
			}
		})
	}
}
