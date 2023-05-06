package zwt_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/zwt"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const diz1 = `
	
Some Director Suite v10.0.2077

SomeCompany


 [2005-12-14]               [12/12]

TEAM Z.W.T ( ZERO WAiTiNG TiME ) 2005 
`

func TestDizDate(t *testing.T) {
	t.Parallel()
	y, m, d := zwt.DizDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = zwt.DizDate(diz1)
	assert.Equal(t, 2005, y)
	assert.Equal(t, time.Month(12), m)
	assert.Equal(t, 14, d)
}

func TestDizTitle(t *testing.T) {
	t.Parallel()
	a, b := zwt.DizTitle(diz1)
	assert.Equal(t, "Some Director Suite v10.0.2077", a)
	assert.Equal(t, "SomeCompany", b)

	a, b = zwt.DizTitle("")
	assert.Equal(t, "", a)
	assert.Equal(t, "", b)
}

const nfo1 = `
▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄
█▀ ▄▄▄▄▄▄ ▀ ▄▄▄▄▄▄ ▀ ▄▄▄▄▄▄ ▀ ▄▄▄▄▄▄ ▀▓
▄▓ ██▓  ██▓ ██▓  ██▓ ██▓  ██▓ ██▓  ██▓ █
█ ▄██▓  ██▓ ██▓  ██▓ ██▓  ██▓ ██▓  ██▓ █
█▄     ▄██▓ ██▓ ▄██▓ ██▓  ██▓ ██▓  ██▓ █
▓ ▄██▀     ██▓  ▄▄▄ ███▄▄▄▄  ██▓  ██▓ █
▒ ██▓  ██▓ ██▓  ██▓ ██▓  ██▓ ██▓  ██▓ █ x!FEAR
░ ██▓▄▄██▓ ▀█▓▄▄██▀ ██▓  ██▓ ▀█▓▄▄██▀ █
▀▄ ▄             ▀   ▄      ██▓▄         ▓▄▄▄█▀▀▀ ▀
░▓█▄          ▄▄▀  ▄▓██▄   ▐▀ ▀▌█▄▄        ▄▄▄▄▀▄
░███▀ ▄    ▄███░ ▄██▀░▀██▄ ▐█▄█▌▀█████▄██████▓▀ ░█▄        ░  ▄▄▀   ▄▄▄▄▄▄
░███░     ▀░███░█▓▀░░   ███ ███░░ ░▀▀███▀░ ▄▀ ▄  ▐█▓        ▄▓█▌ ▄▓██▀▀▀▓██▓▄
░███░      ░███░███░    ███░███░     ██▓░    ▀▓▀ ░███▄     ░▐██▌▐▓██░░   ████▌░
░███░      ░███░███░    ███░███░     ███░   ░▄▄▄  ████▓▄   ░▐██▌▐██▌░    ▐███▌░
░███░      ░███░███▄▄▄▄ ▀██░███░     ███░    ▓██░ ███░▀██▄ ░▐██▌▐██▌░    ░░░░░
░███░  ▄   ░███░███░░░░ ███░███░     ███░    ███░ ███░ ░▀█▓▄▐██▌▐██▌░   ▄▄▄▄▄
░▓██░▄█▀█▄ ░███░███░    ███░███░     ███░   ░███░ ███░  ░ ▀████▌▐██▌░    ▐▓██▌░
░▓▓██▀░░░▀▓▄▓██░███░    ███░███░     ███░   ░███░ █▓█░    ░░███▌▐██▌░    ░███▌░
░▓█▀░░   ░░▀█▓█░███░    ▓██░█▓█░     █▓█░   ░▓██░ ▓██░     ░▐▓█▌▐█▓▌░    ░███▌░
░▀░░       ░░▀█░▓██░   ▄▓▓█░▓██░    ▄▓██▄    █▓█░ █▓█░░    ░▐██▌▐▓██████████▓▌░
░ ▄▄▄▄▄▄▄▄▄▄▄  ▀██░ ▄▄▄▄▄▄▄▄▄▄▄▄▄▓ ▄▄▄     ▐▀ ▀▌ ▄▄▄▄▄ ▄▄▄▄▄   ▄▄▄▄▄▄  ▄▄▄▄▄▄
█              ▀▄              █ ██▓     ▐█▄▓▌██▓  ██▓  ██▓ ██▓  ██▓ █    █
▓                              █ ██▓      ██▓ ██▓  ██▓  ██▓ ██▓  ██▓ ▓    ▓
▒                              █ ██▓▀▀    ██▓ ██▓  ██▓  ██▓ ██▓ ▄██▓ █    ▒
░                              █ ██▓  ▄▄▄ ██▓ ██▓  ██▓  ██▓ ██▓  ▄▄▄ █    ░
▄                              █ ██▓  ██▓ ██▓ ██▓  ██▓  ██▓ ██▓  ██▓ █    ▄
							  ▓ ▀█▓▄▄██▓ ██▓ ██▓  ██▓  ███▄▀█▓▄▄██▀ █
							  ▀█▄▄▄▄▄▄▄▄▄██▓ ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▓▀


┌────────────────────────────────────────────────────────────────────────────┐ 
│■■                        » RELEASE iNFORMATiON «                         ■■│ 
└─┬──────────────────────────────────────────────────────────────────────────┘ 
│░│ SOFTWARE NAME : SomeSoftware.v1.0.0.1146                                    
│░├────────────────────────────────────┬─────────────────────────────────────■ 
│░│ PROTECTiON    : SN                 │  CRACKER       : TEAM Z.W.T         
│░├────────────────────────────────────┼─────────────────────────────────────■ 
│░│ RELEASE TYPE  : OOO000             │  SUPPLiER      : TEAM Z.W.T         
│░├────────────────────────────────────┼─────────────────────────────────────■ 
│░│ RELEASE DATE  : 2005-09-10         │  PACKER        : TEAM Z.W.T         
│░├────────────────────────────────────┼─────────────────────────────────────■ 
│░│ LANGUAGE      : ENGLiSH            │  SiZE          : 9  x 5.00MB
│░├────────────────────────────────────┼─────────────────────────────────────■ 
│░│ FORMAT        : ZIP/RAR            │  ZiP NAME      : zsd1146*.zip       
┌─┴────────────────────────────────────┴─────────────────────────────────────┐ 
│■■                          » ADDiTiONAL NOTES «                          ■■│ 
└─┬──────────────────────────────────────────────────────────────────────────┘ 
│░│ COMPANY       : SomeCompany                                                   
│░├──────────────────────────────────────────────────────────────────────────■ 
│░│ PLATFORM      : WiNALL                                                    
│░├──────────────────────────────────────────────────────────────────────────■ 
│░│ SOFTWARE TYPE : UTiLiTY                                                   
│░├──────────────────────────────────────────────────────────────────────────■ 
│░│ URL           : http://www.somecompany.com                                    
┌─┴──────────────────────────────────────────────────────────────────────────┐ 
│■■                           » RELEASE NOTES «                            ■■│ 
└────────────────────────────────────────────────────────────────────────────┘ 

Complete system deployment solution using disk imaging technology`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := zwt.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = zwt.NfoDate(nfo1)
	assert.Equal(t, 2005, y)
	assert.Equal(t, time.Month(9), m)
	assert.Equal(t, 10, d)
}
