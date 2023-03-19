package prods_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/stretchr/testify/assert"
)

func TestProductionsAPIv1_Download(t *testing.T) {
	p := prods.ProductionsAPIv1{}
	err := p.Download(nil, prods.DownloadsAPIv1{})
	assert.NotNil(t, err)

	p = example1
	err = p.Download(io.Discard, prods.DownloadsAPIv1{})
	assert.NotNil(t, err)

	dl := prods.DownloadsAPIv1{
		LinkClass: "SceneOrgFile",
		URL:       "https://files.scene.org/view/parties/2000/ambience00/demo/feestje.zip",
	}
	p = example1
	err = p.Download(io.Discard, dl)
	assert.Nil(t, err)
}
