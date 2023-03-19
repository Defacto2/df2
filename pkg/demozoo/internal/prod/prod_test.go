package prod_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prod"
	"github.com/stretchr/testify/assert"
)

func TestProduction_URL(t *testing.T) {
	p := prod.Production{}
	err := p.URL()
	assert.NotNil(t, err)

	p = prod.Production{ID: 1}
	err = p.URL()
	assert.Nil(t, err)
}

func TestURL(t *testing.T) {
	s, err := prod.URL(-1)
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
	s, err = prod.URL(1)
	assert.Nil(t, err)
	assert.Equal(t, "https://demozoo.org/api/v1/productions/1?format=json", s)
	s, err = prod.URL(158411)
	assert.Nil(t, err)
	assert.Equal(t, "https://demozoo.org/api/v1/productions/158411?format=json", s)
}

func TestProduction_Get(t *testing.T) {
	p := prod.Production{}
	res, err := p.Get()
	assert.NotNil(t, err)
	assert.Empty(t, res)

	p = prod.Production{ID: -50}
	res, err = p.Get()
	assert.NotNil(t, err)
	assert.Empty(t, res)

	p = prod.Production{ID: 1}
	res, err = p.Get()
	assert.Nil(t, err)
	assert.Equal(t, "Rob Is Jarig", res.Title)
}
