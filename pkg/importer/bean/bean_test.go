package bean_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/bean"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `▒   Disks: 3 x 4,75mb                                   Date : May 01, 2008   ▒
░   OS   : Windows                                     			      ░`

const nfo2 = `                                                                               
Disks: 3 x 2,88mb                               Date : March 2,2009      ▒
																			
OS   : Windows                                                            ░ 
▄  ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄  ▄▄   ▄    `

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := bean.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = bean.NfoDate(nfo1)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(5), m)
	assert.Equal(t, 1, d)

	y, m, d = bean.NfoDate(nfo2)
	assert.Equal(t, 2009, y)
	assert.Equal(t, time.Month(3), m)
	assert.Equal(t, 2, d)
}
