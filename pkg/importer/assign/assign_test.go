package assign_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/importer/assign"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const nfo1 = `
                        SUPPLiER : TEAM ASSiGN     
PACKAGER : TEAM ASSiGN (QUADRA, GoGe-MonSooN)     
												 
   DATE : 01 OCTORBER 2010                       
 NUMBER : ASGN-0669                              
   TYPE : DEMO                          
   DiSK : 04 x 1.5MB                             
												 
PLATFORM : WiNDOWS                                
LANGUAGE : ENGLiSH                                
`

const nfo2 = `
DATE : 10 APRiL 2010  
                                
NUMBER : ASGN-0286      
						
  TYPE : XACKED        
						
  DiSK : 08 x 1.5MB     
`

func TestNfoDate(t *testing.T) {
	t.Parallel()
	y, m, d := assign.NfoDate(internal.RandStr)
	assert.Equal(t, 0, y)
	assert.Equal(t, time.Month(0), m)
	assert.Equal(t, 0, d)

	y, m, d = assign.NfoDate(nfo1)
	assert.Equal(t, 2010, y)
	assert.Equal(t, time.Month(10), m)
	assert.Equal(t, 1, d)

	y, m, d = assign.NfoDate(nfo2)
	assert.Equal(t, 2010, y)
	assert.Equal(t, time.Month(4), m)
	assert.Equal(t, 10, d)
}
