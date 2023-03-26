package acronym_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/stretchr/testify/assert"
)

func TestGroup_Get(t *testing.T) {
	t.Parallel()
	g := acronym.Group{}
	err := g.Get(nil)
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = g.Get(db)
	assert.NotNil(t, err)

	g = acronym.Group{Name: "Defacto2"}
	err = g.Get(db)
	assert.Nil(t, err)
	assert.Equal(t, "DF2", g.Initialism)
}

func TestFix(t *testing.T) {
	t.Parallel()
	i, err := acronym.Fix(nil)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	_, err = acronym.Fix(db)
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	t.Parallel()
	s, err := acronym.Get(nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, err = acronym.Get(db, "")
	assert.Nil(t, err)
	assert.Equal(t, "", s)
	s, err = acronym.Get(db, "Defacto2")
	assert.Nil(t, err)
	assert.Equal(t, "DF2", s)
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
			if got := acronym.Trim(tt.args.s); got != tt.want {
				t.Errorf("Trim() = %v, want %v", got, tt.want)
			}
		})
	}
}
