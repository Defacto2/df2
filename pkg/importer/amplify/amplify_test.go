package amplify_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/amplify"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const diz1 = `  :::      ..AMPLiFY 2008..    :::
._/\_______ ▄▄▄ ▄▄▄▄ ▄▄▄  _____/\_,
--))-+----▄( ▓▀█ ▓▓█▓ █▀▓ /-----+-/
 \\  ▀ ▀▓▀▀█▀▓▀▓▐▌█▀█▀▀.)______(--

 AlgoMusic.AMB.ElectraBass.v1.2.VS
			Ti-AMPLiFY            

 -[08.15.2008]===========[01/02]-
`

const diz2 = `     ▄ ▄▄█▄▄▄▄▄█▄▄▄▄▄▄█[P.R.O.U.D.L.Y  P.R.E.S.E.N.T.S]▄▄▄▄▄▄▄▄▄█▄▄▄▄▄█▄▄ ▄
▀                                            ▀
┌────────────[ NAME Some.Audio.Some.Some-Some.Some-1.v1.02.VST-AMPLiFY
├────────────[ DATE 05-16-2006
└────────────[ DISK xx/02
`

const diz3 = `  -[2009 Jan]===========[01/28]-`

const diz4 = `  -[XMAS2008]===========[01/01]-`

const diz5 = `-[2008]===========[01/01]-`

const nfo1 = `  ·─:■┌■--────────┐_|_|____┌─+┐_.              ._┌──┐__|_:___┌────────--■┐■:─·
┌┼┘└───·▌▓▒▒░+┤-│─└─┬──┘─·∙─[ rELEASE iNFO ]─∙· └∙+┘ └+┬─────░▒▒▓▐─·┘└┼┐
·∙┘·            └·┘   ·     └+∙─*·        ·*─∙+┘         └·┘            ·└∙·
								   ||
	   TEAM AMPLiFY  :...SUPPLIER  ││  REL-DATE..:  01.30.2008       
	   TEAM AMPLiFY  :...PACKAGER  ││  RELEASE#..:  507
				N/A  :....CRACKER  ││  DISKS.....:  32 x 15Mb        
				N/A  :.PROTECTION  ││  TYPE......:  SAMPLES          
			VipZone  :....COMPANY  ││  OS........:  WinAll           
								   ||

┌──────────────────────────────────────────────────────────────────────┐
└--─·■▓░■┌■─■┐_._             _._┌·┐_┌─·┐_|______|_┌─────────■┐■░▓■·─--┘`

func TestDizDate(t *testing.T) {
	t.Parallel()
	y, m, d := amplify.DizDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = amplify.DizDate(diz1)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(8), m)
	assert.Equal(t, 15, d)

	y, m, d = amplify.DizDate(diz2)
	assert.Equal(t, 2006, y)
	assert.Equal(t, time.Month(5), m)
	assert.Equal(t, 16, d)

	y, m, d = amplify.DizDate(diz3)
	assert.Equal(t, 2009, y)
	assert.Equal(t, time.Month(1), m)
	assert.Equal(t, 0, d)

	y, m, d = amplify.DizDate(diz4)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 0, d)

	y, m, d = amplify.DizDate(diz5)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)
}

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := amplify.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = amplify.NfoDate(nfo1)
	assert.Equal(t, 2008, y)
	assert.Equal(t, time.Month(1), m)
	assert.Equal(t, 30, d)
}
