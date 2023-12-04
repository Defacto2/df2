package adsr_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/adsr"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `███████████████████████████████████████████████████████████████████████████████
██▀▀                                                                       ▀▀██
█                                                                             █
     RELEASE DATE................................................: 06-2018     
█                                                                             █
██▄▄                                                                       ▄▄██
███████████████████████████████████████████████████████████████████████████████`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := adsr.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = adsr.NfoDate(nfo1)
	assert.Equal(t, 2018, y)
	assert.Equal(t, time.Month(6), m)
	assert.Equal(t, 1, d)
}
