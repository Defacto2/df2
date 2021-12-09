package group_test

import (
	"testing"

	"github.com/Defacto2/df2/lib/groups/internal/group"
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
			if gotOk := group.Clean(tt.name, false); gotOk != tt.wantOk {
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
