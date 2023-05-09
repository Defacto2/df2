package hexwars_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/arcade"
	"github.com/Defacto2/df2/pkg/importer/hexwars"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `    @BfB                         -+-+-+-+-+-+-+-                         @;@B     
@BfB                                                                 @;@B     
@BfB                                                                 @iB@     
@BfB                       SUPPLiER : HEXWARS                        @i@@     
@BfB                SOMEER/KEYGENER : TEAM HEXWARS:fraktaL           @iB@     
@BfB                       PACKAGER : HEXWARS                        @iB@     
@BfB                                                                 @iB@     
@BfB                           DATE : 29.07.2018                     @i@@     
@BfB                         NUMBER : HW-350                         @iB@     
@BfB                           TYPE : RETAiL/PATCHED/SOMEGEN         @iB@     
@BfB                           DiSK : 21 x 25 MB                     @iB@     
@BfB                                                                 @iB@     
@BfB                       PLATFORM : SomeOS AU/VSTi                 @iB@     
@BfB                     PROTECTiON : WEBAUTH,SHA1,RSA2048,BASE64    @iB@  `

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := arcade.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = hexwars.NfoDate(nfo1)
	assert.Equal(t, 2018, y)
	assert.Equal(t, time.Month(7), m)
	assert.Equal(t, 29, d)
}
