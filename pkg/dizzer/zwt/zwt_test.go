package zwt_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/dizzer/zwt"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const diz1 = `
	
Disk Director Suite v10.0.2077

Acronis


 [2005-12-14]               [12/12]

TEAM Z.W.T ( ZERO WAiTiNG TiME ) 2005 
`

func TestDizDate(t *testing.T) {
	y, m, d := zwt.DizDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = zwt.DizDate(diz1)
	assert.Equal(t, 2005, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 14, d)
}

func TestDizTitle(t *testing.T) {
	a, b := zwt.DizTitle(diz1)
	assert.Equal(t, "Disk Director Suite v10.0.2077", a)
	assert.Equal(t, "Acronis", b)

	a, b = zwt.DizTitle("")
	assert.Equal(t, "", a)
	assert.Equal(t, "", b)
}
