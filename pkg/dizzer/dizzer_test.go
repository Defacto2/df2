package dizzer_test

import (
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/dizzer"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

const (
	r1 = "Acronis.Disk.Director.Suite.v10.0.2077.Russian.Incl.Keymaker-ZWT"
	r2 = "Apollo-tech.No1.Video.Converter.v3.8.17.Incl.Keymaker-ZWT"
	r3 = "SiSoftware.Sandra.Pro.Business.XI.SP3.2007.6.11.40.Multilingual.Retail.Incl.Keymaker-ZWT"
)

func dir() string {
	return filepath.Join(internal.Testdata(2), "rar")
}

func rar() string {
	return filepath.Join(dir(), "dizzer.rar")
}

func TestRun(t *testing.T) {
	t.Parallel()
	err := dizzer.Run(nil, nil, "")
	assert.NotNil(t, err)
	err = dizzer.Run(nil, nil, rar())
	assert.Nil(t, err)
}

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

func TestTitle(t *testing.T) {
	t.Parallel()
	s := dizzer.PathTitle("")
	assert.Equal(t, "", s)
	s = dizzer.PathTitle("HeLLo worLD! ")
	assert.Equal(t, "HeLLo worLD!", s)
	s = dizzer.PathTitle(r1)
	assert.Equal(t, "Acronis Disk Director Suite v10.0.2077 Russian including keymaker", s)
	s = dizzer.PathTitle("Acronis.Disk.Director.Suite.v10.1.Russian.Incl.Keymaker-ZWT")
	assert.Equal(t, "Acronis Disk Director Suite v10.1 Russian including keymaker", s)
	s = dizzer.PathTitle("Acronis.Disk.Director.Suite.v10.Russian.Incl.Keymaker-ZWT")
	assert.Equal(t, "Acronis Disk Director Suite v10 Russian including keymaker", s)
	s = dizzer.PathTitle(r2)
	assert.Equal(t, "Apollo-tech No1 Video Converter v3.8.17 including keymaker", s)
	s = dizzer.PathTitle(r3)
	assert.Equal(t, "SiSoftware Sandra Pro Business XI SP3 2007 6 11 40 Multilingual Retail including keymaker", s)
}
