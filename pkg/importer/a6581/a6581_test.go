package a6581_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/a6581"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `                                                                               
Release.Date....: 03-2023                                  
Release.Size....: 19 * 4,77MB                              
Release.Number..: 6581-02049                                
Supplier........: Team 6581                                
Operating.System: Windows    [■]                           
				  Macintosh  [■]                           `

const nfo2 = `                                                                                
Release.Date....: 12-2013                                   
Release.Format..: vstpreset                                 
Release.Size....: 02 * 4,77MB                               
Release.Number..: 6581-0296                                 `

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := a6581.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = a6581.NfoDate(nfo1)
	assert.Equal(t, 2023, y)
	assert.Equal(t, time.Month(3), m)
	assert.Equal(t, 1, d)

	y, m, d = a6581.NfoDate(nfo2)
	assert.Equal(t, 2013, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 1, d)
}
