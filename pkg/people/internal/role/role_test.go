package role_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/people/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestRole_String(t *testing.T) {
	t.Parallel()
	s := role.Writers
	assert.Equal(t, "writers", s.String())
}

func TestList(t *testing.T) {
	t.Parallel()
	s, i, err := role.List(nil, nil, 0)
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	assert.Len(t, s, 0)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, i, err = role.List(db, io.Discard, 9999)
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	assert.Len(t, s, 0)
	s, i, err = role.List(db, io.Discard, role.Everyone)
	assert.Nil(t, err)
	assert.Greater(t, i, 1)
	assert.Greater(t, len(s), 1)
	s, i, err = role.List(db, io.Discard, role.Musicians)
	assert.Nil(t, err)
	assert.Greater(t, i, 1)
	assert.Greater(t, len(s), 1)
}

func TestPeopleStmt(t *testing.T) {
	t.Parallel()
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

func TestRoles(t *testing.T) {
	t.Parallel()
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

func TestRename(t *testing.T) {
	t.Parallel()
	i, err := role.Rename(nil, "", "", 9999)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = role.Rename(db, "", "", 9999)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	i, err = role.Rename(db, "", "", role.Everyone)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)

	const (
		name    = "Rhythm Addiction"
		replace = "Testing 1 2 3 4 5"
		falseN  = "abcdef01234blahblah"
	)

	i, err = role.Rename(db, name, "", role.Everyone)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	i, err = role.Rename(db, name, falseN, role.Everyone)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)

	i, err = role.Rename(db, name, replace, role.Artists)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), i)
	i, err = role.Rename(db, replace, name, role.Artists)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), i)
}

func TestClean(t *testing.T) {
	t.Parallel()
	b, err := role.Clean(nil, nil, "", 9999)
	assert.NotNil(t, err)
	assert.Equal(t, false, b)
	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	b, err = role.Clean(db, io.Discard, "", 9999)
	assert.NotNil(t, err)
	assert.Equal(t, false, b)
	b, err = role.Clean(db, io.Discard, "%!somenick", 9999)
	assert.NotNil(t, err)
	assert.Equal(t, false, b)
	b, err = role.Clean(db, io.Discard, "%!somenick", role.Artists)
	assert.Nil(t, err)
	assert.Equal(t, true, b)
}

func TestCleanS(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
