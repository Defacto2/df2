package beat_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/beat"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = ` 
░██▓▒░ ░▒▓▓░  │                                                   
░░░░░░░░░░░░\ │                                                   
░██▓▒░ ░▒▓▓░\\│                       ░░░░░░░░░░░░\               
░█SUPPLiER▓░ \   Team BEAT            ░██▓▒░ ░▒▓▓░\\              
░██▓▒░ ░▒▓▓░  │                       ░█  DATE  ▓░ \   1013 2004  
░░░░░░░░░░░░\ │                       ░██▓▒░ ░▒▓▓░  │             
░██▓▒░ ░▒▓▓░\\│                       ░░░░░░░░░░░░\ │             
░█  TYPE  ▓░ \   Live sequencer       ░██▓▒░ ░▒▓▓░\\│             
░██▓▒░ ░▒▓▓░  │                       ░█  DiSC  ▓░ \   XX/06      
░░░░░░░░░░░░\ │                       ░██▓▒░ ░▒▓▓░  │             
░██▓▒░ ░▒▓▓░\\│                       ░░░░░░░░░░░░  │             
░█   OS   ▓░ \   WinALL                \\        \\ │             
░██▓▒░ ░▒▓▓░  │                         \\________\\              `

const nfo2 = `       Date.: 0319 2007                   Supplier.: Team BEAT  
Disk.: [xx/02]                     Cracker..: Team BEAT  `

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := beat.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = beat.NfoDate(nfo1)
	assert.Equal(t, 2004, y)
	assert.Equal(t, time.Month(10), m)
	assert.Equal(t, 13, d)

	y, m, d = beat.NfoDate(nfo2)
	assert.Equal(t, 2007, y)
	assert.Equal(t, time.Month(3), m)
	assert.Equal(t, 19, d)
}
