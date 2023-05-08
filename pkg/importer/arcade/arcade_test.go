package arcade_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/arcade"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `    ░ ░░ ░▒▓█████████ ▄█▄ ░▒▓████▄▄ ▀▀▀███▀▀▀ ▄▄▄███▓▒░ ▄█▄ █████████▓▒░ ░░ ░
▀                                 ▀
Some.VSTi.AU.v4.12.MAC.OSX.INTEL 

░ ░▒▓█████████ ▄█▄ ░▒▓███████████████████████▓▒░ ▄█▄ █████████▓▒░ ░
▀                                 ▀
		Date.: 12/2008 
		Disks: 04 * 4,77MB
   
   ░▒▓███████████████████████▓▒░ `

const nfo2 = `                        P R O U D L Y   P R E S E N T S
                
Some.Some.Some.Some.VST.AU.RTAS.v1.5.1.MAC.OSX.INTEL 
		  
					 PROTECTION : RSA+CRC32+BASE32 
					 SIZE ......: 05 * 4,77MB 
					 DATE ......: 02/2011 
					 URL........: http://www.some.com `

const nfo3 = `         Name.: Some.Somefactory.v2.02.WinAll-ArCADE
Date.: 05.01.06
Type.: APP
OpSys: WinAll
Disks: 1 * 1.44mb`

const nfo4 = `
DATE:          08.31.04
`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := arcade.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = arcade.NfoDate(nfo1)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 0, d)

	y, m, d = arcade.NfoDate(nfo2)
	assert.Equal(t, 2011, y)
	assert.Equal(t, time.Month(2), m)
	assert.Equal(t, 0, d)

	y, m, d = arcade.NfoDate(nfo3)
	assert.Equal(t, 2006, y)
	assert.Equal(t, time.Month(5), m)
	assert.Equal(t, 1, d)

	y, m, d = arcade.NfoDate(nfo4)
	assert.Equal(t, 2004, y)
	assert.Equal(t, time.Month(8), m)
	assert.Equal(t, 31, d)
}
