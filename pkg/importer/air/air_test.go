package air_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/air"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `                                PROUDLY PRESENTS

Some.Line.SomeSome.VST.v1.0.5 

	SUPPLIER ..: TEAM AiR   
	PROTECTION : SERIAL       
	SIZE ......: 02 * 2,78MB
	DATE ......: 01/2010 
	URL........: http://www.somesome.com `

const nfo2 = ` ┌─────────────────────────────────────┐
│::: i n f o :::::::::::::::::::::::: ├──────[aDDiCTiON.iN.rELEASiNG]──────┐
├═════════════════════════════════════┘                                    │
│                                                                          │
│  supplier....: dLLord                 released....: 19 March, 1999       │
│  cracker.....: dLLord                 # of disks..: 01 x 1.44 Mb         │
│  packer......: dLLord                 protection..: serial/nag           │
│  tester......: dLLord                 type........: music util           │
│                                                                          │
│  requires....: Pentium 133, Windows9xNT, Sound  card  with  mike, guitar │
│                would be  preferred, if you are in the mood of using this │
│                program =)                                                │
│                                                                          │`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := air.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = air.NfoDate(nfo1)
	assert.Equal(t, 2010, y)
	assert.Equal(t, time.Month(1), m)
	assert.Equal(t, 0, d)

	y, m, d = air.NfoDate(nfo2)
	assert.Equal(t, 1999, y)
	assert.Equal(t, time.Month(3), m)
	assert.Equal(t, 19, d)
}
