package zone_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/zone"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const diz1 = `Some.AutoTune.PlugIn.v1.1-ZONE
________________________  _______     
.:\   _  \_____ \______   |/       \:::.
:__\  /    /   |   \  |  .|   _,    \__:
/\__\/.   /    |    \_|  ||  (/     /_/\
\/__//   /\    |     /|  ||  /    //__\/
::://   /  \  _|   // |  :| /    //  \::
::/    /    \     //  |  .| \    /    \:
:/_________ /_____/___|   |\_\  /____ /:
--========\/=[Z O N E]|___|===\/====\/--
-==[Win9x/Me]=--=[12/20/00]=-=[01/01]==-`

const diz2 = `Some.Box.1.47-ZONE
[OS:WIN9X/NT] [DATE:17/06/00] [DiSK:o1/o1]`

const diz3 = `    Some AutoSome DX PlugIn v3.04
ÚÄ ÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄ Ä¿
³ ²ÛÛÛßßÛÛÛ°ÛÛÛÛßÛÛÛÛÛÛÛßßÛÛÛÜÛÛÛßßÛÛÛ ³
³  ßßß  ÛÛÛ°ÛÛÛÛ ÛÛÛ²ÛÛÛ  ÛÛÛ²ÛÛÛ  ßßß ³
³ ²ÛÛÛßßßßß²ÛÛÛÛ ÛÛÛ²ÛÛÛ  ÛÛÛ²ÛÛÛßß zk ³
³ °ÛÛÛ  ÛÛÛ²ÛÛÛÛ ÛÛÛ²ÛÛÛ  ÛÛÛ²ÛÛÛ  ÛÛÛ ³
³ °ÛÛÛ °ÛÛÛ²ÛÛÛÛ ÛÛÛ²ÛÛÛ  ÛÛÛ²ÛÛÛ  ÛÛÛ ³
³ °ÛÛÛ °ÛÛÛ²ÛÛÛÛ ÛÛÛ²ÛÛÛ  ÛÛÛ²ÛÛÛ °ÛÛÛ ³
³ °ÛÛÛÜÜÛÛÛ°ÛÛÛÛÜÛÛÛ°ÛÛÛ  ÛÛÛ°ÛÛÛÜÜÛÛÛ ³
ÀÄÄÄÄÄÄÄÄÄÄ ÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄÄ ÄÄÙ
[03.28.02]                         [1/1]`

const diz4 = `Some.Develop.Drums.Instrument.v2.0-ZONE
[oS:WIN9X/NT] [DATE:4/23/00] [DiSK:o1/o1]
Total Release size Of 704k, Completed By iNSOMNiA...`

const diz5 = `Spin.Audio.Room.Verb.DX-VST.v1.1.WORKING-ZONE
[oS:WIN9X/NT] [DATE:2/4/00] [DiSK:o1/o1]`

const diz6 = `Emagic.Logic.Audio.Platinum.v4.61-ZONE
________________________  _______     
.:\   _  \_____ \______   |/       \:::.
:__\  /    /   |   \  |  .|   _,    \__:
/\__\/.   /    |    \_|  ||  (/     /_/\
\/__//   /\    |     /|  ||  /    //__\/
::://   /  \  _|   // |  :| /    //  \::
::/    /    \     //  |  .| \    /    \:
:/_________ /_____/___|   |\_\  /____ /:
--========\/=[Z O N E]|___|===\/====\/--
-==[Win9x]==-==[11/17/00]==-==[03/03]==-`

func TestDizDate(t *testing.T) {
	t.Parallel()
	y, m, d := zone.DizDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = zone.DizDate(diz1)
	assert.Equal(t, 2000, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 20, d)

	y, m, d = zone.DizDate(diz2)
	assert.Equal(t, 2000, y)
	assert.Equal(t, time.Month(6), m)
	assert.Equal(t, 17, d)

	y, m, d = zone.DizDate(diz3)
	assert.Equal(t, 2002, y)
	assert.Equal(t, time.Month(3), m)
	assert.Equal(t, 28, d)

	y, m, d = zone.DizDate(diz4)
	assert.Equal(t, 2000, y)
	assert.Equal(t, time.Month(4), m)
	assert.Equal(t, 23, d)

	y, m, d = zone.DizDate(diz5)
	assert.Equal(t, 2000, y)
	assert.Equal(t, time.Month(2), m)
	assert.Equal(t, 4, d)

	y, m, d = zone.DizDate(diz6)
	assert.Equal(t, 2000, y)
	assert.Equal(t, time.Month(11), m)
	assert.Equal(t, 17, d)
}

func TestDizTitle(t *testing.T) {
	t.Parallel()
	s := zone.DizTitle("")
	assert.Equal(t, "", s)

	s = zone.DizTitle(diz1)
	assert.Equal(t, "Some AutoTune PlugIn v1.1", s)

	s = zone.DizTitle(diz2)
	assert.Equal(t, "Some Box 1 47", s)

	s = zone.DizTitle(diz3)
	assert.Equal(t, "Some AutoSome DX PlugIn v3.04", s)
}

const nfo1 = `  ┌────┐             ┌───-- - ∙ ·   PRESENTS   · ∙ - -─────┐           ┌────┐
:  ┌───────────────────────────────────────────────────────────────────┐  :
∙  : │                  Some.Waves.Audio.v2.4-ZONE                 │ :  ∙
:  └─────────────────────-- -  ∙  ·     ·  ∙  - -──────────────────────┘  :
└────┘┌─────────────────────────────┐  ┌────────────────────────────┐└────┘
	  : ┌───────────────────────────│──│────────────────────────────┐ :
	  └─│ [SUPPLIER: Zone Team]     :  : [CLASS: Sequencer]         │─┘
		│                           ∙  ∙                            │
		│ [CRACKER: Zone Team]           [Price: $60]               │
		│                                                           │
		│ [RELEASE No.: 170]             [z-mkwa24.zip = 1 Mb]      │ 
		│                           ∙  ∙                            │
	  ┌─│ [PACKAGER: Zone Team]     :  : [DATE: 02/05/01]           │─┐
	  : └───────────────────────────│──│────────────────────────────┘ :
┌────┐└─────────────────────────────┘  └────────────────────────────┘┌────┐
:  ┌───────────────────────────────────────────────────────────────────┐  :
∙  : │                         RELEASE NOTES                         │ :  ∙
:  └─────────────────────-- -  ∙  ·     ·  ∙  - -──────────────────────┘  :
└────┘                                                               └────┘`

const nfo2 = `  ?????????????????????? ???    ZONE PRESENTS    ??? ?????????????????????
??                                                                        ??
??                 VAZ 2010 v1.02 (c) Software-Technology                 ??
??                                                                        ??
??     DATE  : 05.29.02                     SUPPLIER   : ZONE TEAM        ??
??     DISKS : z-vz2010.zip x 1 = 4.6 Mb    CRACKER    : ZONE TEAM        ??
??     TYPE  : Software Synthesizer         PACKER     : ZONE TEAM        ??
??     PRICE : $183                         RELEASE #  : 332              ??
??                                                                        ??
????????????????????????????????????????????????????????????????????????????`

const nfo3 = ` .-=========================================================================-.
|Cracked By :  Dave Porno.                                                  |
|Disks      :  1x390Kb [zne-ab145.zip]                                      |
|Class      :  Synth.                                                       |
|URL        :  http://www.andyware.com/abox/index.html                      | 
|Relase Date:  03/07/2000                                                   |
-=========================================================================-'`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := zone.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = zone.NfoDate(nfo1)
	assert.Equal(t, 2001, y)
	assert.Equal(t, time.Month(5), m)
	assert.Equal(t, 2, d)

	y, m, d = zone.NfoDate(nfo2)
	assert.Equal(t, 2002, y)
	assert.Equal(t, time.Month(5), m)
	assert.Equal(t, 29, d)

	y, m, d = zone.NfoDate(nfo3)
	assert.Equal(t, 2000, y)
	assert.Equal(t, time.Month(7), m)
	assert.Equal(t, 3, d)
}
