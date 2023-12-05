package alfiso_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/alfiso"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `  ││█                                                              █││  
││█             REL. DATE : 28-03-2003                           █││  
││█             SUPPLiER  : ALF                                  █││  
││█             RELEASE # : 260                                  █││  
││█             FORMAT    : BIN/CUE                              █││  
││█             DiSKS     : 31 * 14.3MB                          █││  `

const nfo2 = `ALFiSO  $     █▓                                                                
MMMMMMMHHHHHMMMMIMMH:::... 'IHI,...   $#                                        
        #$     0*                                                               
MMMMMMMMMMMMMMM H:MHIII:...:  ':.:.   $# Release date : 21.08.02                
        #$     █▓                                                               
MMMMMMMMMMMMMMM'H:MHIII:...:  ':.:.   $# Format Type : EMU                      
        #$     █▓                                                               
MMMMMHHMMHHMHI ,,MMMII:'::I:I.  :I::  $# Status : 28 * 14.3MB                   
        #$     █▓                                                               `

const nfo3 = `                                                                               
│▓█          ReL. DaTe : 2004-04-11                                │▓██  
│▓█          SuPPLieR  : ALFiSO                                    │▓██  
│▓█          ReLeASe # : 359                                       │▓██  
│▓█          FoRMaT    : BIN/CUE                                   │▓██  
│▓█          DiSKS     : 18 * 14.3mb                               │▓██
│▓█                                                                │▓██   `

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := alfiso.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = alfiso.NfoDate(nfo1)
	assert.Equal(t, 2003, y)
	assert.Equal(t, time.Month(3), m)
	assert.Equal(t, 28, d)

	y, m, d = alfiso.NfoDate(nfo2)
	assert.Equal(t, 2002, y)
	assert.Equal(t, time.Month(8), m)
	assert.Equal(t, 21, d)

	y, m, d = alfiso.NfoDate(nfo3)
	assert.Equal(t, 2004, y)
	assert.Equal(t, time.Month(4), m)
	assert.Equal(t, 11, d)
}
