package audiostrike_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/audiostrike"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `$&&&                      
&&&$       
$&&&  DATE....: 29/05/2015
&&&$       
$&&&  TYPE....: SAmples   
&&&$       
$&&&  FORMAT..: WAV       
&&&$       
$&&&  SiZE....: 02 * 50 mb
&&&$       `

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := audiostrike.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = audiostrike.NfoDate(nfo1)
	assert.Equal(t, 2015, y)
	assert.Equal(t, time.Month(5), m)
	assert.Equal(t, 29, d)

	// y, m, d = audioutopia.NfoDate(nfo2)
	// assert.Equal(t, 2015, y)
	// assert.Equal(t, time.Month(9), m)
	// assert.Equal(t, 27, d)
}
