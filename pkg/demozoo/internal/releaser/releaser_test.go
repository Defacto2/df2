package releaser_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/demozoo/internal/releaser"
	"github.com/stretchr/testify/assert"
)

const (
	speccyGroup = 78489
	sprintGroup = 112416
)

func TestReleaserV1_Print(t *testing.T) {
	r := releaser.ReleaserV1{}
	err := r.Print(nil)
	assert.Nil(t, err)

	err = r.Print(io.Discard)
	assert.Nil(t, err)

	w := strings.Builder{}
	err = r.Print(&w)
	assert.Nil(t, err)
	assert.Contains(t, w.String(), "external_links")
}

func TestReleaser_URL(t *testing.T) {
	r := releaser.Releaser{}
	err := r.URL()
	assert.Nil(t, err)
	r = releaser.Releaser{ID: sprintGroup}
	err = r.URL()
	assert.Nil(t, err)
}

func TestReleaser_Get(t *testing.T) {
	r := releaser.Releaser{}
	rel, err := r.Get()
	if !errors.Is(err, context.DeadlineExceeded) {
		assert.NotNil(t, err)
	}
	assert.Empty(t, rel)

	r = releaser.Releaser{ID: sprintGroup}
	rel, err = r.Get()
	assert.Nil(t, err)
	assert.NotEmpty(t, rel)
}

func TestReleaser_Prods(t *testing.T) {
	r := releaser.Releaser{}
	rel, err := r.Prods()
	assert.NotNil(t, err)
	assert.Empty(t, rel)
	r = releaser.Releaser{ID: sprintGroup}
	rel, err = r.Prods()
	if !errors.Is(err, context.DeadlineExceeded) {
		assert.Nil(t, err)
	}
	assert.Empty(t, rel) // a valid group with no productions

	r = releaser.Releaser{ID: speccyGroup}
	rel, err = r.Prods()
	if !errors.Is(err, context.DeadlineExceeded) {
		assert.Nil(t, err)
		assert.NotEmpty(t, rel) // a valid group with 1 production
	}
}
