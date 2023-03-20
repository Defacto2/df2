package rename_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/rename"
	"github.com/stretchr/testify/assert"
)

func TestClean(t *testing.T) {
	_, err := rename.Clean(nil, nil, "")
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	_, err = rename.Clean(db, io.Discard, "")
	assert.Nil(t, err)
	b, err := rename.Clean(db, io.Discard, "Defacto2")
	assert.Nil(t, err)
	assert.Equal(t, false, b)
	b, err = rename.Clean(db, io.Discard, "d-e-f-a-c-t-o-2")
	assert.Nil(t, err)
	assert.Equal(t, true, b)
}

func TestFmtSyntax(t *testing.T) {
	s := rename.FmtSyntax("")
	assert.Equal(t, "", s)
	s = rename.FmtSyntax("hello&&&&world")
	assert.Equal(t, "hello & world", s)
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
			if got := rename.CleanS(tt.args.s); got != tt.want {
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
			if got := rename.TrimThe(tt.args.g); got != tt.want {
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
			if got := rename.TrimDot(tt.args.s); got != tt.want {
				t.Errorf("TrimDot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFmtByName(t *testing.T) {
	s := rename.FmtByName("")
	assert.Equal(t, "", s)
	s = rename.FmtByName("abc")
	assert.Equal(t, "", s)
	s = rename.FmtByName("rzsoft ftp")
	assert.Equal(t, "RZSoft FTP", s)
	s = rename.FmtByName("Hashx")
	assert.Equal(t, "Hash X", s)
}

func TestFormat(t *testing.T) {
	s := rename.Format("")
	assert.Equal(t, "", s)
	s = rename.Format("mydox")
	assert.Equal(t, "MyDox", s)
	s = rename.Format("myfxp")
	assert.Equal(t, "MyFXP", s)
	s = rename.Format("myiso")
	assert.Equal(t, "MyISO", s)
	s = rename.Format("mynfo")
	assert.Equal(t, "MyNFO", s)
	s = rename.Format("pc-my")
	assert.Equal(t, "PC-My", s)
	s = rename.Format("lsdstuff")
	assert.Equal(t, "LSDStuff", s)
}

func TestFmtExact(t *testing.T) {
	s := rename.FmtExact("")
	assert.Equal(t, "", s)
	s = rename.FmtExact("tcsm bbs")
	assert.Equal(t, "TCSM BBS", s)
	s = rename.FmtExact("Scenet")
	assert.Equal(t, "scenet", s)
}
