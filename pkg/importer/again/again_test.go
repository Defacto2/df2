package again_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/again"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const diz1 = `    |    AccuRev Enterprise v4.7.0    |
|         (c) AccuRev, Inc        |
| .   Released on 14/09/2008      |
| ::.        [01/09]             _|_
_ __|________________________________\:`

func TestDizDate(t *testing.T) {
	t.Parallel()
	y, m, d := again.DizDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = again.DizDate(diz1)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(9), m)
	assert.Equal(t, 14, d)
}

const nfo1 = `▒ ▐██▌                          OS..........: Windows                 ▐▌███
░ ▐██▌                          Language....: English                  █ ███
   ███                          Protection..: License                  █ ▐██▌
▄   ▀██▄                        Size........: 09 x 4.77mb             ▐▌  ███
▄▓▄    ▀▀▀                      Date........: 14/09/2008             ▄▀   ███`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := again.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = again.NfoDate(nfo1)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(9), m)
	assert.Equal(t, 14, d)
}
