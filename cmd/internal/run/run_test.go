package run

import (
	"database/sql"
	"io"
	"log"
	"testing"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	cfg configger.Config
	db  *sql.DB
	l   *zap.SugaredLogger
)

func init() {
	var err error
	db, err = database.Connect(configger.Defaults())
	if err != nil {
		log.Fatal(err)
	}
	cfg = configger.Defaults()
	l = logger.Development().Sugar()
}

func TestRun(t *testing.T) {
	t.Parallel()
	err := Data(nil, nil, database.Flags{})
	assert.NotNil(t, err)
	err = Data(db, nil, database.Flags{})
	assert.NotNil(t, err)
	err = Data(db, nil, database.Flags{Type: "blah"})
	assert.NotNil(t, err)
	err = Data(db, nil, database.Flags{Type: "create"})
	assert.NotNil(t, err)
}

func TestAPIs(t *testing.T) {
	err := APIs(nil, nil, arg.APIs{})
	assert.NotNil(t, err)
	err = APIs(db, io.Discard, arg.APIs{})
	assert.Nil(t, err)
}

func TestDemozoo(t *testing.T) {
	err := Demozoo(nil, nil, nil, configger.Config{}, arg.Demozoo{})
	assert.NotNil(t, err)
	err = Demozoo(db, nil, nil, configger.Config{}, arg.Demozoo{})
	assert.NotNil(t, err)
	err = Demozoo(db, io.Discard, l, configger.Config{}, arg.Demozoo{})
	assert.ErrorIs(t, err, ErrNothing)
}

func TestGroups(t *testing.T) {
	err := Groups(nil, nil, nil, arg.Group{})
	assert.NotNil(t, err)
	err = Groups(db, io.Discard, io.Discard, arg.Group{})
	assert.NotNil(t, err)
	err = Groups(db, io.Discard, io.Discard, arg.Group{Filter: "bbs"})
	assert.Nil(t, err)
}

func TestNew(t *testing.T) {
	err := New(nil, nil, nil, configger.Config{})
	assert.NotNil(t, err)
	err = New(db, io.Discard, l, configger.Config{})
	assert.NotNil(t, err)
}

func TestPeople(t *testing.T) {
	err := People(nil, nil, "", arg.People{})
	assert.NotNil(t, err)
	err = People(db, io.Discard, "", arg.People{})
	assert.Nil(t, err)
}

func TestRename(t *testing.T) {
	s := []string{}
	err := Rename(nil, nil, s...)
	assert.NotNil(t, err)
	err = Rename(db, io.Discard, s...)
	assert.NotNil(t, err)
}

func TestTestSite(t *testing.T) {
	err := TestSite(nil, nil, "")
	assert.NotNil(t, err)
	err = TestSite(db, io.Discard, "")
	assert.Nil(t, err)
}
