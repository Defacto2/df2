package prods_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/stretchr/testify/assert"
)

func TestProductionsAPIv1_Authors(t *testing.T) {
	p := prods.ProductionsAPIv1{}
	a := p.Authors()
	assert.Empty(t, a)

	p = example1
	a = p.Authors()
	assert.Nil(t, a.Text)
	assert.Contains(t, a.Code, "Ile")
	assert.Contains(t, a.Art, "Ile")
	assert.Nil(t, a.Audio)

	p = example3
	a = p.Authors()
	assert.Nil(t, a.Text)
	assert.Nil(t, a.Code)
	assert.Nil(t, a.Art)
	assert.Nil(t, a.Audio)
}
