package audioutopia_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/audioutopia"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `     d#   -------------------------------------------------------------------------------    @E
d#                                   Release Info                                       @E
d#   -------------------------------------------------------------------------------    @E
d#   Type.................: VST2                                                        @E
d#   Platform.............: Windows x86/x64                                             @E
d#   Homepage.............: http://www.example.com/                                     @E
d#   Date.................: 19.DEC.2015 / 02.JAN.2016                                   @E
d#                                                                                      @E`

const nfo2 = `Date.................: 27.SEPT.2015`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := audioutopia.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = audioutopia.NfoDate(nfo1)
	assert.Equal(t, 2015, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 19, d)

	y, m, d = audioutopia.NfoDate(nfo2)
	assert.Equal(t, 2015, y)
	assert.Equal(t, time.Month(9), m)
	assert.Equal(t, 27, d)
}
