package assets_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/assets"
	"github.com/Defacto2/df2/pkg/assets/internal/scan"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
)

func TestClean(t *testing.T) {
	t.Parallel()
	c := assets.Clean{}
	err := c.Walk(nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = c.Walk(db, io.Discard)
	assert.NotNil(t, err)

	c = assets.Clean{
		Name:   "invalid",
		Remove: false,
		Human:  false,
		Config: conf.Defaults(),
	}
	err = c.Walk(db, io.Discard)
	assert.NotNil(t, err)

	const ok = "DOWNLOAD"
	c = assets.Clean{
		Name:   ok,
		Remove: false,
		Human:  false,
		Config: conf.Defaults(),
	}
	err = c.Walk(db, io.Discard)
	assert.Nil(t, err)
}

func TestCreateUUIDMap(t *testing.T) {
	t.Parallel()
	i, ids, err := assets.CreateUUIDMap(nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	assert.Equal(t, database.IDs(nil), ids)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, ids, err = assets.CreateUUIDMap(db)
	assert.Nil(t, err)
	assert.Greater(t, i, 0)
	assert.NotEqual(t, database.IDs(nil), ids)
}

func TestWalker(t *testing.T) {
	t.Parallel()
	c := assets.Clean{}
	err := c.Walker(nil, nil, -1, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = c.Walker(db, io.Discard, -1, nil)
	assert.NotNil(t, err)

	d, err := directories.Init(conf.Defaults(), false)
	assert.Nil(t, err)
	err = c.Walker(db, io.Discard, -1, &d)
	assert.NotNil(t, err)

	c = assets.Clean{
		Remove: false,
		Human:  false,
		Config: conf.Defaults(),
	}
	err = c.Walker(db, io.Discard, assets.Download, &d)
	assert.Nil(t, err)
}

func TestSkip(t *testing.T) {
	t.Parallel()
	f, err := scan.Skip("", nil)
	assert.NotNil(t, err)
	assert.Equal(t, scan.Files{}, f)

	d, err := directories.Init(conf.Defaults(), false)
	assert.Nil(t, err)
	f, err = scan.Skip("", &d)
	assert.Nil(t, err)
	_, ok := f["blank.png"]
	assert.Equal(t, true, ok)
}

func TestTargets(t *testing.T) {
	t.Parallel()
	const allTargets = 5
	tests := []struct {
		name   string
		target assets.Target
		want   int
	}{
		{"", assets.All, allTargets},
		{"", assets.Image, 2},
		{"error", -1, 0},
	}
	cfg := conf.Defaults()
	d, err := directories.Init(cfg, false)
	if err != nil {
		t.Error(err)
	}
	color.Enable = false
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, _ := assets.Targets(cfg, tt.target, &d); len(got) != tt.want {
				t.Errorf("Targets() = %v, want %v", got, tt.want)
			}
		})
	}
}
