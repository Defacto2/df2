package role_test

import (
	"testing"

	"github.com/Defacto2/df2/lib/people/internal/role"
)

func TestRoles(t *testing.T) {
	type args struct {
		r string
	}
	tests := []struct {
		name string
		args args
		want role.Role
	}{
		{"empty", args{""}, role.Everyone},
		{"artist", args{"artists"}, role.Artists},
		{"a", args{"a"}, role.Artists},
		{"all", args{"all"}, role.Everyone},
		{"error", args{"xxx"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := role.Roles(tt.args.r); got != tt.want {
				t.Errorf("Roles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeopleStmt(t *testing.T) {
	type args struct {
		role               string
		includeSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"error", args{"text", false}, false},
		{"empty f", args{"", false}, true},
		{"empty t", args{"", true}, true},
		{"writers", args{"writers", true}, true},
		{"writers", args{"w", true}, true},
		{"musicians", args{"m", true}, true},
		{"coders", args{"c", true}, true},
		{"artists", args{"a", true}, true},
		{"all", args{"", true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := role.PeopleStmt(role.Roles(tt.args.role), tt.args.includeSoftDeletes); len(got) > 0 != tt.want {
				t.Errorf("sqlPeople() = %v, want = %v", len(got) > 0, tt.want)
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
		{"", args{"the blah"}, "the blah"},
		{"", args{"a dude,blah"}, "a dude,blah"},
		{"", args{"name1,name2,!name3!"}, "name1,name2,name3!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := role.CleanS(tt.args.s); got != tt.want {
				t.Errorf("CleanS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"a nick"}, "a nick"},
		{"", args{"--a nick"}, "a nick"},
		{"", args{" ?!nick!! "}, "nick"},
		{"", args{"?!nick!!,someone else"}, "nick,someone else"},
		{"", args{"?!nick!!,--someone-else++"}, "nick,--someone-else"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := role.Trim(tt.args.s); got != tt.want {
				t.Errorf("Trim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRename(t *testing.T) {
	const (
		org     = "Rhythm Addiction"
		replace = "Testing 1 2 3 4 5"
		falseN  = "abcdef01234blahblah"
	)
	type args struct {
		replacement string
		name        string
		r           role.Role
	}
	tests := []struct {
		name      string
		args      args
		wantCount int64
		wantErr   bool
	}{
		{"empty", args{}, 0, true},
		{"missing replacement", args{name: org, r: role.Musicians}, 0, true},
		{"missing name", args{replacement: replace, r: role.Musicians}, 0, true},
		{"missing role", args{replacement: replace, name: org}, 0, true},
		{"404 name", args{replacement: replace, name: falseN, r: role.Musicians}, 0, false},
		{"okay", args{replacement: replace, name: org, r: role.Artists}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotCount, err := role.Rename(tt.args.replacement, tt.args.name, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCount != tt.wantCount {
				t.Errorf("Rename() = %v, want %v", gotCount, tt.wantCount)
				return
			}

			role.Rename(org, replace, role.Artists) // restore name
		})
	}
}
