package record_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/dizzer/record"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

var (
	dir = filepath.Join(internal.Testdata(3))
	rar = filepath.Join(dir, "rar", "dizzer.rar")
	zip = filepath.Join(dir, "pkzip", "PKZ80A1.ZIP")
)

const (
	sha384 = "749b328f5284e5c196f932a07194894e0fc50c1d9c414457883bc2d79a5ee8a94ac2981e5168ad4df1f4a6405dce99c7"
	summd5 = "7c7d17c6faec74918f4a7047e1c50412"
	magic  = "RAR archive data, v4, os"
)

func TestSum386(t *testing.T) {
	sum, err := record.Sum386(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "", sum)

	f, err := os.Open(rar)
	assert.Nil(t, err)
	defer f.Close()

	sum, err = record.Sum386(f)
	assert.Nil(t, err)
	assert.Equal(t, sha384, sum)
}

func TestSumMD5(t *testing.T) {
	sum, err := record.SumMD5(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "", sum)

	f, err := os.Open(rar)
	assert.Nil(t, err)
	defer f.Close()

	sum, err = record.SumMD5(f)
	assert.Nil(t, err)
	assert.Equal(t, summd5, sum)
}

func TestDetermine(t *testing.T) {
	s, err := record.Determine("")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
	s, err = record.Determine(rar)
	assert.Nil(t, err)
	assert.Equal(t, magic, s)
	s, err = record.Determine(zip)
	assert.Nil(t, err)
	assert.Equal(t, "Zip archive data, at least v1.0 to extract, compression method=Shrinking", s)
}
