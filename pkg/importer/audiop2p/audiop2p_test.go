package audiop2p_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/audiop2p"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `####################################################################################
#                           General Release Information                            #
####################################################################################
# Type.................: ADDON                                                     #
# Platform.............: ALL                                                       #
# Audio Format.........: N/A                                                       #
# Product WebSite......: www.example.com/?lang=en&page=products/nexus/skin         #
# Release..............: 62                                                        #
# Release Title........: ap-xxxx.rar                                               #
# Prepared.............: 07-16-2009                                                #
# CRC & File Counts....: ap-xxxx.sfv                                               #
####################################################################################`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := audiop2p.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = audiop2p.NfoDate(nfo1)
	assert.Equal(t, 2009, y)
	assert.Equal(t, time.Month(7), m)
	assert.Equal(t, 16, d)
}
