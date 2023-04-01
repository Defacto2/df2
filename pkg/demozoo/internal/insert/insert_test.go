package insert_test

import (
	"context"
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/insert"
	"github.com/Defacto2/df2/pkg/demozoo/internal/releases"
	models "github.com/Defacto2/df2/pkg/models/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func TestRecord_Insert(t *testing.T) {
	t.Parallel()
	r := insert.Record{}
	res, err := r.Insert(nil)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	r = insert.Record{}
	res, err = r.Insert(db)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	r = insert.Record{Title: "placeholder"}
	res, err = r.Insert(db)
	assert.Nil(t, err)
	id, err := res.LastInsertId()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, id)
	// remove newly inserted record
	i, err := models.Files(qm.Where("id=?", id)).DeleteAll(
		context.Background(), db, true)
	assert.Nil(t, err)
	assert.NotEqual(t, 1, i)
}

func TestProds(t *testing.T) {
	t.Parallel()
	err := insert.Prods(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = insert.Prods(db, io.Discard, nil)
	assert.NotNil(t, err)

	r := releases.Productions{}
	err = insert.Prods(db, io.Discard, &r)
	assert.Nil(t, err)
}

func TestProd(t *testing.T) {
	t.Parallel()
	r, err := insert.Prod(nil, nil, releases.ProductionV1{})
	assert.NotNil(t, err)
	assert.Empty(t, r)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	r, err = insert.Prod(db, io.Discard, releases.ProductionV1{})
	assert.NotNil(t, err)
	assert.Empty(t, r)

	const existsID = 294625
	prod := releases.ProductionV1{ID: existsID}
	r, err = insert.Prod(db, io.Discard, prod)
	assert.Nil(t, err)
	assert.Empty(t, r)

	const c64ID = 116376
	prod = releases.ProductionV1{ID: c64ID}
	r, err = insert.Prod(db, io.Discard, prod)
	assert.Nil(t, err)
	assert.Empty(t, r)
}
