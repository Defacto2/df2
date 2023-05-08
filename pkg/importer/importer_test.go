package importer_test

import (
	"testing"

	dizzer "github.com/Defacto2/df2/pkg/importer"
	"github.com/stretchr/testify/assert"
)

const (
	r1 = "Acronis.Disk.Director.Suite.v10.0.2077.Russian.Incl.Keymaker-ZWT"
	r2 = "Apollo-tech.No1.Video.Converter.v3.8.17.Incl.Keymaker-ZWT"
	r3 = "SiSoftware.Sandra.Pro.Business.XI.SP3.2007.6.11.40.Multilingual.Retail.Incl.Keymaker-ZWT"
)

func TestGroup(t *testing.T) {
	t.Parallel()
	s := dizzer.PathGroup("")
	assert.Equal(t, "", s)
	s = dizzer.PathGroup("HeLLo worLD! ")
	assert.Equal(t, "", s)
	s = dizzer.PathGroup(r1)
	assert.Equal(t, "ZWT", s)
	s = dizzer.PathGroup(r2)
	assert.Equal(t, "ZWT", s)
}
