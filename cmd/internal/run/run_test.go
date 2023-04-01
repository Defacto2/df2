package run_test

import (
	"database/sql"
	"io"
	"log"
	"testing"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals
var (
	cfg conf.Config
	db  *sql.DB
	l   *zap.SugaredLogger
)

func init() { //nolint:gochecknoinits
	var err error
	db, err = database.Connect(conf.Defaults())
	if err != nil {
		log.Fatal(err)
	}
	cfg = conf.Defaults()
	l = logger.Development().Sugar()
}

func TestRun(t *testing.T) {
	t.Parallel()
	err := run.Data(nil, nil, database.Flags{})
	assert.NotNil(t, err)
	err = run.Data(db, nil, database.Flags{})
	assert.NotNil(t, err)
	err = run.Data(db, nil, database.Flags{Type: "blah"})
	assert.NotNil(t, err)
	err = run.Data(db, nil, database.Flags{Type: "create"})
	assert.NotNil(t, err)
}

func TestAPIs(t *testing.T) {
	t.Parallel()
	err := run.APIs(nil, nil, arg.APIs{})
	assert.NotNil(t, err)
	err = run.APIs(db, io.Discard, arg.APIs{})
	assert.NotNil(t, err)
}

func TestDemozoo(t *testing.T) {
	t.Parallel()
	err := run.Demozoo(nil, nil, nil, conf.Config{}, arg.Demozoo{})
	assert.NotNil(t, err)
	err = run.Demozoo(db, nil, nil, conf.Config{}, arg.Demozoo{})
	assert.NotNil(t, err)
	err = run.Demozoo(db, io.Discard, l, conf.Config{}, arg.Demozoo{})
	assert.ErrorIs(t, err, run.ErrNothing)
}

func TestGroups(t *testing.T) {
	t.Parallel()
	err := run.Groups(nil, nil, nil, arg.Group{})
	assert.NotNil(t, err)
	err = run.Groups(db, io.Discard, io.Discard, arg.Group{})
	assert.NotNil(t, err)
	err = run.Groups(db, io.Discard, io.Discard, arg.Group{Filter: "bbs"})
	assert.Nil(t, err)
}

func TestNew(t *testing.T) {
	t.Parallel()
	err := run.New(nil, nil, nil, conf.Config{})
	assert.NotNil(t, err)
	err = run.New(db, io.Discard, l, conf.Config{})
	assert.NotNil(t, err)
}

func TestPeople(t *testing.T) {
	t.Parallel()
	err := run.People(nil, nil, "", arg.People{})
	assert.NotNil(t, err)
	err = run.People(db, io.Discard, "", arg.People{})
	assert.Nil(t, err)
}

func TestRename(t *testing.T) {
	t.Parallel()
	s := []string{}
	err := run.Rename(nil, nil, s...)
	assert.NotNil(t, err)
	err = run.Rename(db, io.Discard, s...)
	assert.NotNil(t, err)
}

func TestTestSite(t *testing.T) {
	t.Parallel()
	err := run.TestSite(nil, nil, "")
	assert.NotNil(t, err)
	err = run.TestSite(db, io.Discard, "")
	assert.Nil(t, err)
}
