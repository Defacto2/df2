package spirit_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/spirit"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `?    SUPPLIER: TEAM SPiRiT                   DATE: 02/27/2007                ?
?     CRACKER: -                             REL#: 841                       ?
?  PROTECTION: GAS MASK                      SIZE: 1 CD (23) x 15Mb          ?`

const nfo2 = `  · -─■┴-░─┬─-───────────────── -    ·    - ───────░──────────-───┬-░─┴■-─ ·         
│   │                                                      │   │
:     ·└───┘·       Some.Interactive.DD.King.Style-SPiRiT        ·└───┘·     :
│                                                                            │
│    SUPPLIER: TEAM SPiRiT                   DATE: 13.08.2006                │
│     CRACKER: -                             REL#: 806                       │
│  PROTECTION: GAS MASK                      SIZE: 1 CD (10) x 15 MB         │
├─┐■                                                                      ■┌─┤
│▀┴┴──────────────────────────────────────────────────────────────────────┴┴▀│`

const nfo3 = `  ▀▀████▄░▄█▓██▀▀▀▀▀▀▀▀▀▀▀▀██▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀██▀▀▀▀▀▀▀▀▀▀▀▀████▄░▄█▓██▀▀
▀███▒█▓█▀         ▄▄▀▀      ▀  ▄▓  ▄▓  ▄▓      ▀▀▄▄         ▀███▒█▓█▀
  ▀█▓█▀         ▄▀ ▄        ▓ ▓ ▓ ▓ ▀ ▓ ▓        ▄ ▀▄         ▀█▓█▀
   ▐█▌         ▐▌  ▐▌       ▓ ▓ ▀ ▓▀  ▓▄▀       ▐▌  ▐▌         ▐█▌
	█           ▀▄▄▀                             ▀▄▄▀           █
	░                                                           ░
	  SUPPLiER.....: TEAM SPiRiT
	  CRACKER......: none	
	  PROTECTiON...: none
	  FORMAT.......: bin/cue
	  RLS DATE.....: 19.04.2010
	  RELEASE #....: 1128
	  RLS SiZE.....: big
	  RLS DISCS....: some
	░                                                           ░
	▓                   ▄▀▀▄            ▄▀▀▄                    ▓
   ▐█▌                 ▐▌  ▐▌    ▄▀▀▄  ▐▌  ▐▌                  ▐█▌
  ▄█▓█▄                 ▀▄ ▀    ▐▌  ▐▌  ▀▄ ▀                  ▄█▓█▄
▄▄███▀███▄▄                ▀▀▄▄   ▀▄ ▀     ▀▀▄▄             ▄▄███▀███▄▄
▀▀████▄░▄█▓██▀▀▀▀▀▀▀▀███▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀███▀▀▀▀▀▀▀▀████▄░▄█▓██▀▀ `

const nfo4 = `  · -─■┴-░─┬─-───────────────── -    ·    - ───────░──────────-───┬-░─┴■-─ ·         
│   │                                                      │   │
:     ·└───┘· Some.Interactive.Jam.Sessions.Some.Style.v.2-SPiRiT └───┘·     :
│                                                                            │
│    SUPPLIER: TEAM SPiRiT                   DATE: 2006-10-xx                │
│     CRACKER: -                             REL#: 819                       │
│  PROTECTION: GAS MASK                      SIZE: 1 CD (6) x 15 MB         │
├─┐■                                                                      ■┌─┤
│▀┴┴──────────────────────────────────────────────────────────────────────┴┴▀│`

const nfo5 = `           SUPPLiER.....: TEAM SPiRiT
CRACKER......: none	
PROTECTiON...: none
FORMAT.......: bin/cue
RLS DATE.....: 19.04.2010
RELEASE #....: 1130
RLS SiZE.....: big
RLS DISCS....: some`

const nfo6 = `        SUPPLiER.....: TEAM SPiRiT            RLS DATE....: 2007-07-13
CRACKER......: -                      RELEASE #...: 864
PROTECTiON...: -                      RLS SiZE....: 31 x 15Mb
FORMAT.......: CDDA                   RLS DISCS...: 1 CD`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := spirit.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = spirit.NfoDate(nfo1)
	assert.Equal(t, 2007, y)
	assert.Equal(t, time.Month(2), m)
	assert.Equal(t, 27, d)

	y, m, d = spirit.NfoDate(nfo2)
	assert.Equal(t, 2006, y)
	assert.Equal(t, time.Month(8), m)
	assert.Equal(t, 13, d)

	y, m, d = spirit.NfoDate(nfo3)
	assert.Equal(t, 2010, y)
	assert.Equal(t, time.Month(4), m)
	assert.Equal(t, 19, d)

	y, m, d = spirit.NfoDate(nfo4)
	assert.Equal(t, 2006, y)
	assert.Equal(t, time.Month(10), m)
	assert.Equal(t, 0, d)

	y, m, d = spirit.NfoDate(nfo5)
	assert.Equal(t, 2010, y)
	assert.Equal(t, time.Month(4), m)
	assert.Equal(t, 19, d)

	y, m, d = spirit.NfoDate(nfo6)
	assert.Equal(t, 2007, y)
	assert.Equal(t, time.Month(7), m)
	assert.Equal(t, 13, d)
}
