package imagemagick_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/images/internal/imagemagick"
	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	t.Parallel()
	gif := filepath.Join("testdata", "test.gif") + ` GIF 1280x32 1280x32+0+0 8-bit sRGB 2c 661B 0.000u 0:00.000`

	b, err := imagemagick.ID("")
	assert.NotNil(t, err)
	assert.Nil(t, b)
	b, err = imagemagick.ID(filepath.Join("dev", "null"))
	assert.NotNil(t, err)
	assert.Nil(t, b)
	b, err = imagemagick.ID("imagemagick_test.go")
	assert.NotNil(t, err, "not an image file")
	assert.Nil(t, b)
	b, err = imagemagick.ID(filepath.Join("testdata", "test.gif"))
	assert.Nil(t, err)
	assert.Equal(t, gif+"\n", string(b))
}

func TestConvert(t *testing.T) {
	t.Parallel()
	gif := filepath.Join("testdata", "test.gif")
	tmp := filepath.Join(t.TempDir(), "test.png")
	err := imagemagick.Convert(nil, "", "")
	assert.NotNil(t, err)
	err = imagemagick.Convert(io.Discard, gif, "")
	assert.Nil(t, err)
	err = imagemagick.Convert(io.Discard, "", tmp)
	assert.NotNil(t, err)
	err = imagemagick.Convert(io.Discard, gif, tmp)
	assert.Nil(t, err)
	defer os.Remove(tmp)
}
