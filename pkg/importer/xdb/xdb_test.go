package xdb_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/xdb"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `
                                 -=[PRESENTS]=-
╔════════════════════════════════════════════════════════════════════════════╗
│         SomeSome Audio Some v1.0.2 MacOSX (C) SomeSome Audio Company       │
╚════════════════════════════════════════════════════════════════════════════╝
│ Release Type ..: Cracked                                               │ 
│ Supplier.......: Team Xdb                                              │ 
│ Protection.....: Name/Serial                                           │ 
│ Release Date...: 19.09.2012                                            │ 
│ Format.........: AU/VST                                                │ 
╔════════════════════════════════════════════════════════════════════════════╗
│                       │ R e l e a s e   N o t e s │                        │
╚════════════════════════════════════════════════════════════════════════════╝`

const nfo2 = `                                 -=[PRESENTS]=-
╔════════════════════════════════════════════════════════════════════════════╗
│      SomeSomeLab AU VSTi v2.1.4 Update Only MAC OSX (C) SomeSome Co.       │
╚════════════════════════════════════════════════════════════════════════════╝
  │ Release Type .....: Cracked                                            │ 
  │ Cracker...........: Xdb                                                │ 
  │ Protection........: Serial                                             │ 
  │ Release Date......: 10.09.2012                                         │ 
  │ Format............: AU/VSTi/RTAS - x32/x64                             │ 
╔════════════════════════════════════════════════════════════════════════════╗
│                       │ R e l e a s e   N o t e s │                        │
╚════════════════════════════════════════════════════════════════════════════╝
  │                                                                        │ `

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := xdb.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = xdb.NfoDate(nfo1)
	assert.Equal(t, 2012, y)
	assert.Equal(t, time.Month(9), m)
	assert.Equal(t, 19, d)

	y, m, d = xdb.NfoDate(nfo2)
	assert.Equal(t, 2012, y)
	assert.Equal(t, time.Month(9), m)
	assert.Equal(t, 10, d)
}
