package arctic_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/arctic"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `     ▄█▓                                                           ▓█▄
▓█   Supplier    : TEAM ArCTiC        System      : [■] W9X/ME   █▓
▓█   Cracker     : TEAM ArCTiC                      [■] Win2k    █▓
▓█   Packager    : TEAM ArCTiC                      [■] WinXP    █▓
▓█   Protection  : PACE                             [ ] Linux    █▓
▓█   Date        : 02-16-03                                      █▓
▓█   Size        : 08 * 2,78MB                                   █▓
▓█                                                               █▓`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := arctic.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = arctic.NfoDate(nfo1)
	assert.Equal(t, 2003, y)
	assert.Equal(t, time.Month(2), m)
	assert.Equal(t, 16, d)
}
