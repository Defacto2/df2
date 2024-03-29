package filter_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/filter"
	"github.com/Defacto2/df2/pkg/groups/internal/rename"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Parallel()
	f := filter.BBS
	assert.Equal(t, "bbs", f.String())
	f = filter.FTP
	assert.Equal(t, "ftp", f.String())
	f = filter.Group
	assert.Equal(t, "group", f.String())
	f = filter.Magazine
	assert.Equal(t, "magazine", f.String())
	f = filter.None
	assert.Equal(t, "", f.String())
	f = 999
	assert.Equal(t, "", f.String())

	r := filter.Get("BBS")
	assert.Equal(t, filter.BBS, r)
}

func TestCount(t *testing.T) {
	t.Parallel()
	i, err := filter.Count(nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = filter.Count(db, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	i, err = filter.Count(db, "qwertyuiopqwertyuiop")
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	i, err = filter.Count(db, "Defacto2")
	assert.Nil(t, err)
	assert.Greater(t, i, 0)
}

func TestList(t *testing.T) {
	t.Parallel()
	s, i, err := filter.List(nil, nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	assert.Len(t, s, 0)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, i, err = filter.List(db, io.Discard, "")
	assert.Nil(t, err)
	assert.Greater(t, i, 0)
	assert.Greater(t, len(s), 0)

	s, i, err = filter.List(db, io.Discard, "bbs")
	assert.Nil(t, err)
	assert.Greater(t, i, 0)
	assert.Greater(t, len(s), 0)
}

func TestSQLWhere(t *testing.T) {
	t.Parallel()
	t.Helper()
	type args struct {
		f              filter.Filter
		incSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"mag-", args{filter.Magazine, false}, "AND section = 'magazine' AND `deletedat` IS NULL"},
		{"bbs-", args{filter.BBS, false}, "AND RIGHT(group_brand_for,4) = ' BBS' AND `deletedat` IS NULL"},
		{"ftp-", args{filter.FTP, false}, "AND RIGHT(group_brand_for,4) = ' FTP' AND `deletedat` IS NULL"},
		{"grp-", args{filter.Group, false}, "AND RIGHT(group_brand_for,4) != ' FTP' AND " +
			"RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND `deletedat` IS NULL"},
		{"mag+", args{filter.Magazine, true}, "AND section = 'magazine'"},
		{"bbs+", args{filter.BBS, true}, "AND RIGHT(group_brand_for,4) = ' BBS'"},
		{"ftp+", args{filter.FTP, true}, "AND RIGHT(group_brand_for,4) = ' FTP'"},
		{"grp+", args{filter.Group, true}, "AND RIGHT(group_brand_for,4) != ' FTP' AND " +
			"RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine'"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, _ := filter.SQLWhere(nil, tt.args.f, tt.args.incSoftDeletes); got != tt.want {
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, got1 := filter.UseHr(tt.args.cap, tt.args.group)
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
		f                  filter.Filter
		includeSoftDeletes bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid soft", args{filter.BBS, true}, false},
		{"valid", args{filter.BBS, false}, false},
		{"invalid", args{filter.Get("invalid filter"), false}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := filter.SQLSelect(nil, tt.args.f, tt.args.includeSoftDeletes)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := filter.Slug(tt.args.name); got != tt.want {
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := rename.FmtSyntax(tt.w); got != tt.want {
				t.Errorf("FmtSyntax() = %v, want %v", got, tt.want)
			}
		})
	}
}
