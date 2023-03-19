package filter_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/filter"
	"github.com/stretchr/testify/assert"
)

func TestProductions_Prods(t *testing.T) {
	p := filter.Productions{}
	r, err := p.Prods(nil, nil, 1)
	assert.NotNil(t, err)
	assert.Len(t, r, 0)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	r, err = p.Prods(db, io.Discard, 1)
	assert.Nil(t, err)
	assert.Len(t, r, 0)
}
