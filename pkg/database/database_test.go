// Package database tests.
package database_test

import (
	"context"
	"database/sql"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/database/internal/templ"
	"github.com/Defacto2/df2/pkg/internal"
	models "github.com/Defacto2/df2/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestTable(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "files", database.Table(0).String())
	assert.Equal(t, "groupnames", database.Table(1).String())
	assert.Equal(t, "netresources", database.Table(2).String())
	assert.Equal(t, "", database.Table(3).String())
}

func TestConnect(t *testing.T) {
	t.Parallel()
	db, err := database.Connect(conf.Config{})
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = database.Connect(conf.Defaults())
	assert.Nil(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestConnInfo0(t *testing.T) {
	t.Parallel()
	s, err := database.ConnDebug(conf.Config{})
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
}

func TestConnInfo1(t *testing.T) {
	t.Parallel()
	s, err := database.ConnDebug(conf.Defaults())
	assert.Nil(t, err)
	assert.Equal(t, "", s)
}

func TestConnInfo2(t *testing.T) {
	t.Parallel()
	cfg := conf.Defaults()
	cfg.DBName = "foo"
	cfg.DBPass = "bar"
	s, err := database.ConnDebug(cfg)
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
}

func TestApprove(t *testing.T) {
	t.Parallel()
	err := database.Approve(nil, nil, conf.Config{}, false)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = database.Approve(db, io.Discard, conf.Defaults(), false)
	assert.Nil(t, err)
}

func TestCheckID(t *testing.T) {
	t.Parallel()
	err := database.CheckID("")
	assert.NotNil(t, err)
	err = database.CheckID("abcde")
	assert.NotNil(t, err)
	err = database.CheckID("00000")
	assert.NotNil(t, err)
	err = database.CheckID("00000876786")
	assert.NotNil(t, err)
	err = database.CheckID("0.0000876786")
	assert.NotNil(t, err)
	err = database.CheckID("-2")
	assert.NotNil(t, err)
	err = database.CheckID("2")
	assert.Nil(t, err)
}

func TestCheckUUID(t *testing.T) {
	t.Parallel()
	err := database.CheckUUID("")
	assert.NotNil(t, err)
	err = database.CheckUUID(internal.RandStr)
	assert.NotNil(t, err)
	err = database.CheckUUID("00000")
	assert.NotNil(t, err)
	err = database.CheckUUID("00000876786")
	assert.NotNil(t, err)
	err = database.CheckUUID("0.0000876786")
	assert.NotNil(t, err)
	err = database.CheckUUID("-2")
	assert.NotNil(t, err)
	err = database.CheckUUID("2")
	assert.NotNil(t, err)
	err = database.CheckUUID(internal.File01UUID)
	assert.Nil(t, err)
}

func TestColumns(t *testing.T) {
	t.Parallel()
	err := database.Columns(nil, nil, 0)
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = database.Columns(db, io.Discard, database.Files)
	assert.Nil(t, err)
	err = database.Columns(db, io.Discard, database.Groups)
	assert.Nil(t, err)
	err = database.Columns(db, io.Discard, database.Netresources)
	assert.Nil(t, err)
}

func TestDateTime(t *testing.T) {
	t.Parallel()
	color.Enable = false
	s, err := database.DateTime(nil)
	assert.Nil(t, err)
	assert.Equal(t, "", s)
	s, err = database.DateTime([]byte("hello world"))
	assert.NotNil(t, err)
	assert.Equal(t, "?", s)
	s, err = database.DateTime([]byte("01-01-2000 00:00:00"))
	assert.NotNil(t, err)
	assert.Equal(t, "?", s)
	s, err = database.DateTime([]byte("2000-01-01T00:00:00Z"))
	assert.Nil(t, err)
	assert.Equal(t, "01 Jan 2000", s)
}

func TestDistinct(t *testing.T) {
	t.Parallel()
	s, err := database.Distinct(nil, "")
	assert.NotNil(t, err)
	assert.Empty(t, s)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, err = database.Distinct(db, "")
	assert.Nil(t, err)
	assert.Empty(t, "", s)
	s, err = database.Distinct(db, "defacto")
	assert.Nil(t, err)
	assert.Len(t, s, 1)
}

func TestFileUpdate(t *testing.T) {
	t.Parallel()
	b, err := database.FileUpdate("", time.Time{})
	assert.NotNil(t, err)
	assert.False(t, b)

	dir := internal.Testdata(2)
	zip := internal.TestZip(2)

	b, err = database.FileUpdate(dir, time.Now())
	assert.Nil(t, err)
	assert.False(t, b)
	b, err = database.FileUpdate(zip, time.Now())
	assert.Nil(t, err)
	assert.False(t, b)
	b, err = database.FileUpdate(zip, time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestFix(t *testing.T) {
	t.Parallel()
	err := database.Fix(nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = database.Fix(db, io.Discard)
	assert.Nil(t, err)
}

func TestDemozooID(t *testing.T) {
	t.Parallel()
	id, err := database.DemozooID(nil, 0)
	assert.NotNil(t, err)
	assert.Equal(t, 0, id)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	id, err = database.DemozooID(db, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, id)
	id, err = database.DemozooID(db, internal.TestDemozooC64)
	assert.Nil(t, err)
	assert.Equal(t, 0, id)
	id, err = database.DemozooID(db, internal.TestDemozooMSDOS)
	assert.Nil(t, err)
	assert.Equal(t, 1047, id)
}

func TestGetID(t *testing.T) {
	t.Parallel()
	id, err := database.GetID(nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, id)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	id, err = database.GetID(db, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, id)
	id, err = database.GetID(db, "qwerty")
	assert.NotNil(t, err)
	assert.Equal(t, 0, id)

	id, err = database.GetID(db, "1")
	assert.Nil(t, err)
	assert.Equal(t, 1, id)
	id, err = database.GetID(db, internal.File01UUID)
	assert.Nil(t, err)
	assert.Equal(t, 1, id)
}

func TestGetKeys(t *testing.T) {
	t.Parallel()
	ids, err := database.GetKeys(nil, "")
	assert.NotNil(t, err)
	assert.Len(t, ids, 0)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()

	all, err := database.GetKeys(db, "")
	assert.Nil(t, err)
	assert.Greater(t, len(all), 1)

	ids, err = database.GetKeys(db, templ.WhereAvailable)
	assert.Nil(t, err)
	assert.Less(t, len(ids), len(all))
}

func TestGetFile(t *testing.T) {
	t.Parallel()
	s, err := database.GetFile(nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, err = database.GetFile(db, "")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
	s, err = database.GetFile(db, internal.RandStr)
	assert.NotNil(t, err)
	assert.Equal(t, "", s)

	ctx := context.Background()
	rec, err := models.FindFile(ctx, db, 1)
	assert.Nil(t, err)
	rec.Filename = null.NewString(internal.File01Name, true)
	_, err = rec.Update(ctx, db, boil.Infer())
	assert.Nil(t, err)

	s, err = database.GetFile(db, internal.File01UUID)
	assert.Nil(t, err)
	assert.Equal(t, internal.File01Name, s)
	s, err = database.GetFile(db, "1")
	assert.Nil(t, err)
	assert.Equal(t, internal.File01Name, s)
}

func TestIsID(t *testing.T) {
	t.Parallel()
	b := database.IsID("")
	assert.False(t, b)
	b = database.IsID(internal.File01Name)
	assert.False(t, b)
	b = database.IsID(internal.File01UUID)
	assert.False(t, b)
	b = database.IsID(internal.RandStr)
	assert.False(t, b)
	b = database.IsID("-1")
	assert.False(t, b)
	b = database.IsID("1.1")
	assert.False(t, b)
	b = database.IsID("1")
	assert.True(t, b)
}

func TestIsUnApproved(t *testing.T) {
	t.Parallel()
	b, err := database.IsUnApproved(nil)
	assert.Nil(t, err)
	assert.False(t, b)

	raw := make([]sql.RawBytes, 7)
	b, err = database.IsUnApproved(raw)
	assert.Nil(t, err)
	assert.False(t, b)

	now := time.Now().Format(time.RFC3339)
	raw[2] = []byte(now)
	b, err = database.IsUnApproved(raw)
	assert.Nil(t, err)
	assert.False(t, b)

	raw[6] = []byte(now)
	b, err = database.IsUnApproved(raw)
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestIsUUID(t *testing.T) {
	t.Parallel()
	b := database.IsUUID("")
	assert.False(t, b)
	b = database.IsUUID(internal.File01Name)
	assert.False(t, b)
	b = database.IsUUID("1")
	assert.False(t, b)
	b = database.IsUUID(internal.RandStr)
	assert.False(t, b)
	b = database.IsUUID("-1")
	assert.False(t, b)
	b = database.IsUUID("1.1")
	assert.False(t, b)
	b = database.IsUUID(internal.File01UUID)
	assert.True(t, b)
	b = database.IsUUID(strings.ToUpper(internal.File01UUID))
	assert.True(t, b)
}

func TestLastUpdate(t *testing.T) {
	t.Parallel()
	tt, err := database.LastUpdate(nil)
	assert.NotNil(t, err)
	assert.Empty(t, tt)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	tt, err = database.LastUpdate(db)
	assert.Nil(t, err)
	assert.Greater(t, tt.Year(), 2010)
}

func TestObfuscateParam(t *testing.T) {
	t.Parallel()
	s := database.ObfuscateParam("")
	assert.Equal(t, "", s)
	s = database.ObfuscateParam("000001")
	assert.Equal(t, "000001", s)
	s = database.ObfuscateParam("1")
	assert.Equal(t, "9b1c6", s)
	s = database.ObfuscateParam("999999999")
	assert.Equal(t, "eb77359232", s)
	s = database.ObfuscateParam("69247541")
	assert.Equal(t, "c06d44215", s)
}

func Test_StripChars(t *testing.T) {
	t.Parallel()
	s := database.StripChars("")
	assert.Equal(t, "", s)
	s = database.StripChars("ooÖØöøO")
	assert.Equal(t, "ooÖØöøO", s)
	s = database.StripChars("o.o|Ö+Ø=ö^ø#O")
	assert.Equal(t, "ooÖØöøO", s)
	s = database.StripChars("A Café!")
	assert.Equal(t, "A Café", s)
	s = database.StripChars("brunräven - över")
	assert.Equal(t, "brunräven - över", s)
}

func Test_StripStart(t *testing.T) {
	t.Parallel()
	s := database.StripStart("")
	assert.Equal(t, "", s)
	s = database.StripStart("hello world")
	assert.Equal(t, "hello world", s)
	s = database.StripStart("--argument")
	assert.Equal(t, "argument", s)
	s = database.StripStart("!!!OMG-WTF")
	assert.Equal(t, "OMG-WTF", s)
	s = database.StripStart("#ÖØöøO")
	assert.Equal(t, "ÖØöøO", s)
	s = database.StripStart("!@#$%^&A(+)ooÖØöøO")
	assert.Equal(t, "A(+)ooÖØöøO", s)
}

func TestTotal(t *testing.T) {
	t.Parallel()
	i, err := database.Total(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, -1, i)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = database.Total(db, io.Discard, nil)
	assert.NotNil(t, err)
	assert.Equal(t, -1, i)

	stmt := "SELECT `id` FROM `files`"
	i, err = database.Total(db, io.Discard, &stmt)
	assert.Nil(t, err)
	assert.Greater(t, i, 1)
}

func Test_TrimSP(t *testing.T) {
	t.Parallel()
	s := database.TrimSP("")
	assert.Equal(t, "", s)
	s = database.TrimSP("abc")
	assert.Equal(t, "abc", s)
	s = database.TrimSP("a b c")
	assert.Equal(t, "a b c", s)
	s = database.TrimSP("a  b  c")
	assert.Equal(t, "a b c", s)
}

func TestWaiting(t *testing.T) {
	t.Parallel()
	i, err := database.Waiting(nil)
	assert.NotNil(t, err)
	assert.Equal(t, -1, i)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = database.Waiting(db)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, i, 0)
}

func TestDeObfuscate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s    string
		want int
	}{
		{"", 0},
		{internal.RandStr, 0},
		{internal.File01Save, 1},
		{internal.File01Save + "?filename=KIWI.EXE", 1},
		{internal.File01URL, 1},
		{"defacto2.net/d/b84058", 9876},
		{"https://defacto2.net/file/detail/af3d95", 8445},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.s, func(t *testing.T) {
			t.Parallel()
			got := database.DeObfuscate(tt.s)
			if got != tt.want {
				t.Errorf("DeObfuscate() = %v, want %v", got, tt.want)
			}
		})
	}
}
