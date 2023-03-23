package sitemap_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	t.Parallel()
	err := sitemap.Create(nil, nil, "")
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()

	err = sitemap.Create(db, io.Discard, "")
	assert.Nil(t, err)

	bb := bytes.Buffer{}
	err = sitemap.Create(db, &bb, "")
	assert.Nil(t, err)
	prefix := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` +
		`<url><loc>https://defacto2.net/welcome</loc><priority>1</priority></url>`
	exist := strings.HasPrefix(bb.String(), prefix)
	assert.Equal(t, true, exist)
}

func TestRoot(t *testing.T) {
	t.Parallel()
	s := sitemap.File.String()
	assert.Equal(t, "f", s)
	s = sitemap.Download.String()
	assert.Equal(t, "d", s)

	assert.Len(t, sitemap.Outputs(), 3)
	assert.Len(t, sitemap.Sorts(), 6)
}

func TestFileList(t *testing.T) {
	t.Parallel()
	s, err := sitemap.FileList(sitemap.Location)
	assert.Nil(t, err)
	assert.NotEqual(t, "", s)
	for _, x := range sitemap.Sorts() {
		assert.Contains(t, s, "https://defacto2.net/file/list/-?output=card&platform=-&section=-&sort="+x)
	}
}

func TestIDs(t *testing.T) {
	t.Parallel()
	ids := sitemap.IDs{1, 345, 543, 665}
	b := ids.Contains(0)
	assert.False(t, b)
	b = ids.Contains(543)
	assert.True(t, b)

	random, err := ids.Randomize(2)
	assert.Nil(t, err)
	assert.Len(t, random, 2)
	assert.NotEqual(t, random[0], random[1])

	s := ids.JoinPaths(sitemap.Location, sitemap.Download)
	assert.Contains(t, s, "https://defacto2.net/d/9b1c6")
}

func TestStyle_Range(t *testing.T) {
	t.Parallel()
	w := io.Discard
	sitemap.NotFound.RangeFiles(w, nil)
	sitemap.NotFound.RangeFiles(w, []string{sitemap.DockerLoc + "/d/9b1c6"})
	sitemap.NotFound.RangeFiles(w, []string{sitemap.Location + "/d/9b1c6"})
	sitemap.NotFound.Range(w, nil)
	sitemap.NotFound.Range(w, []string{sitemap.DockerLoc + "/f/9b1c6"})
	sitemap.NotFound.Range(w, []string{sitemap.Location + "/f/9b1c6"})
}

func TestAbsPaths(t *testing.T) {
	s, err := sitemap.AbsPaths(sitemap.Location)
	assert.Nil(t, err)
	assert.Contains(t, s, "https://defacto2.net/welcome")
}

func TestAbsPathsH3(t *testing.T) {
	t.Parallel()
	s, err := sitemap.AbsPathsH3(nil, nil, "")
	assert.NotNil(t, err)
	assert.Len(t, s, 0)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	s, err = sitemap.AbsPathsH3(db, io.Discard, "")
	assert.Nil(t, err)
	assert.Len(t, s, 9)
}

func TestColor(t *testing.T) {
	t.Parallel()
	s := sitemap.Color404(404)
	assert.Contains(t, s, "✓")
	s = sitemap.ColorCode(200)
	assert.Contains(t, s, "✓")
}

func TestGetBlocked(t *testing.T) {
	t.Parallel()
	i, err := sitemap.GetBlocked(nil)
	assert.NotNil(t, err)
	assert.Len(t, i, 0)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = sitemap.GetBlocked(db)
	assert.Nil(t, err)
	assert.Greater(t, len(i), 0)
}

func TestGetKeys(t *testing.T) {
	t.Parallel()
	i, err := sitemap.GetKeys(nil)
	assert.NotNil(t, err)
	assert.Len(t, i, 0)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = sitemap.GetKeys(db)
	assert.Nil(t, err)
	assert.Greater(t, len(i), 0)
}

func TestGetSoftDeleteKeys(t *testing.T) {
	t.Parallel()
	i, err := sitemap.GetSoftDeleteKeys(nil)
	assert.NotNil(t, err)
	assert.Len(t, i, 0)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = sitemap.GetSoftDeleteKeys(db)
	assert.Nil(t, err)
	assert.Greater(t, len(i), 0)
}

func TestRand(t *testing.T) {
	t.Parallel()
	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()

	i, ids, err := sitemap.RandBlocked(nil, 0)
	assert.NotNil(t, err)
	assert.Len(t, ids, 0)
	assert.Equal(t, 0, i)
	i, ids, err = sitemap.RandBlocked(db, 3)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Greater(t, i, 0)

	i, ids, err = sitemap.RandDeleted(nil, 0)
	assert.NotNil(t, err)
	assert.Len(t, ids, 0)
	assert.Equal(t, 0, i)
	i, ids, err = sitemap.RandDeleted(db, 3)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Greater(t, i, 0)

	i, ids, err = sitemap.RandIDs(nil, 0)
	assert.NotNil(t, err)
	assert.Len(t, ids, 0)
	assert.Equal(t, 0, i)
	i, ids, err = sitemap.RandIDs(db, 3)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Greater(t, i, 0)
}
