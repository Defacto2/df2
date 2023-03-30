package zwt_test

import (
	"strings"
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
	nr := strings.NewReader(internal.RandStr)
	y, m, d := zwt.DizDate(nr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	nr = strings.NewReader(diz1)
	y, m, d = zwt.DizDate(nr)
	assert.Equal(t, 2005, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 14, d)
}

func TestDizTitle(t *testing.T) {
	nr := strings.NewReader(diz1)
	a, b := zwt.DizTitle(nr)
	assert.Equal(t, "Disk Director Suite v10.0.2077", a)
	assert.Equal(t, "Acronis", b)

	a, b = zwt.DizTitle(nil)
	assert.Equal(t, "", a)
	assert.Equal(t, "", b)
}
