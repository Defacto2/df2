package scan_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/buffer"
)

func TestStats(t *testing.T) {
	t.Parallel()
	s, err := scan.Init(conf.Config{})
	assert.NotNil(t, err)
	assert.Empty(t, s)

	s, err = scan.Init(conf.Defaults())
	assert.Nil(t, err)
	assert.NotEmpty(t, s)

	b := &buffer.Buffer{}
	s.Summary(b)
	assert.Contains(t, b.String(), "Total archives scanned:")
}
