package record_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/record"
	"github.com/Defacto2/df2/pkg/importer/zwt"
	"github.com/stretchr/testify/assert"
)

const (
	grp    = "Some Group"
	dizMd5 = "9565262e62cbc52cb2f53518e466663d"
	dizSha = "9d177bdc30a76ca800f6e69e05b2d3146fd63c17dec2653f7f6cf9b104620ce5ce5a8b84c96e0898f2fe3f45de6fb171"
	uuid   = "461d2850-5f18-403f-9937-e207279338ff"
)

const diz1 = `
	
Disk Director Suite v10.0.2077

Acronis


 [2005-12-14]               [12/12]

TEAM Z.W.T ( ZERO WAiTiNG TiME ) 2005 
`

func diz() string {
	return filepath.Join(dir(), "text", "file_id.diz")
}

func TestRecord_New(t *testing.T) {
	t.Parallel()
	r, err := record.New("", "", "")
	assert.NotNil(t, err)
	assert.Empty(t, r)

	tmp := os.TempDir()
	r, err = record.New("", tmp, grp)
	assert.Nil(t, err)
	assert.NotEqual(t, "", r.UUID)
	assert.Equal(t, grp, r.Group)
	assert.Equal(t, record.Section, r.Section)
	assert.Equal(t, record.Platform, r.Platform)
	assert.Equal(t, "release directory: "+tmp, r.Comment)
}

func TestDownload_Create(t *testing.T) {
	t.Parallel()
	dl := record.Download{}
	err := dl.Create("", "")
	assert.ErrorIs(t, err, record.ErrGroup)

	err = dl.Create("", grp)
	assert.NotNil(t, err)

	err = dl.Create(dir(), grp)
	assert.ErrorIs(t, err, record.ErrDir)

	err = dl.Create(rar(), grp)
	assert.Nil(t, err)
	assert.Equal(t, rar(), dl.Path)
	assert.Equal(t, "dizzer.rar", dl.Name)
	assert.Equal(t, int64(14058), dl.Bytes)
	assert.Equal(t, sha384, dl.HashStrong)
	assert.Equal(t, summd5, dl.HashWeak)
	assert.Equal(t, magic, dl.Magic)
	assert.Equal(t, true, dl.ReadDate.IsZero())
	assert.Equal(t, "", dl.ReadTitle)

	err = dl.Create(diz(), zwt.Name)
	assert.Nil(t, err)
	assert.Equal(t, diz(), dl.Path)
	assert.Equal(t, "file_id.diz", dl.Name)
	assert.Equal(t, int64(148), dl.Bytes)
	assert.Equal(t, dizSha, dl.HashStrong)
	assert.Equal(t, dizMd5, dl.HashWeak)
	assert.Equal(t, "ASCII text, with CRLF line terminators", dl.Magic)
	assert.Equal(t, 2005, dl.ReadDate.Year())
	assert.Equal(t, time.Month(9), dl.ReadDate.Month())
	assert.Equal(t, 5, dl.ReadDate.Day())
	assert.Equal(t, "XP Recovery CD Maker v1.01.05 by Super Win Software, Inc.", dl.ReadTitle)
}

func TestDownload_ReadDIZ(t *testing.T) {
	t.Parallel()
	dl := record.Download{}
	err := dl.ReadDIZ("", "")
	assert.ErrorIs(t, err, record.ErrGroup)

	err = dl.ReadDIZ("", grp)
	assert.Nil(t, err)

	err = dl.ReadDIZ("", zwt.Name)
	assert.Nil(t, err)
	assert.Equal(t, true, dl.ReadDate.IsZero())
	assert.Equal(t, "", dl.ReadTitle)

	err = dl.ReadDIZ(diz1, zwt.Name)
	assert.Nil(t, err)
	assert.Equal(t, 2005, dl.ReadDate.Year())
	assert.Equal(t, time.Month(12), dl.ReadDate.Month())
	assert.Equal(t, 14, dl.ReadDate.Day())
	assert.Equal(t, "Disk Director Suite v10.0.2077 by Acronis", dl.ReadTitle)
}
