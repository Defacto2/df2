package group_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/group"
	"github.com/Defacto2/df2/pkg/groups/internal/rename"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Parallel()
	f := group.BBS
	assert.Equal(t, "bbs", f.String())
	f = group.FTP
	assert.Equal(t, "ftp", f.String())
	f = group.Group
	assert.Equal(t, "group", f.String())
	f = group.Magazine
	assert.Equal(t, "magazine", f.String())
	f = group.None
	assert.Equal(t, "", f.String())
	f = 999
	assert.Equal(t, "", f.String())

	r := group.Get("BBS")
	assert.Equal(t, group.BBS, r)
}

func TestCount(t *testing.T) {
	t.Parallel()
	i, err := group.Count(nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = group.Count(db, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	i, err = group.Count(db, "qwertyuiopqwertyuiop")
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	i, err = group.Count(db, "Defacto2")
	assert.Nil(t, err)
	assert.Greater(t, i, 0)
}

func TestList(t *testing.T) {
	t.Parallel()
	s, i, err := group.List(nil, nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	assert.Len(t, s, 0)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, i, err = group.List(db, io.Discard, "")
	assert.Nil(t, err)
	assert.Greater(t, i, 0)
	assert.Greater(t, len(s), 0)

	s, i, err = group.List(db, io.Discard, "bbs")
	assert.Nil(t, err)
	assert.Greater(t, i, 0)
	assert.Greater(t, len(s), 0)
}

func TestSQLWhere(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestSlug(t *testing.T) {
	t.Parallel()
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

func Test_FmtSyntax(t *testing.T) {
	t.Parallel()
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
