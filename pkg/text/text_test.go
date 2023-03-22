package text_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestFix(t *testing.T) {
	t.Parallel()
	err := text.Fix(nil, nil, configger.Config{})
	assert.NotNil(t, err)

	cfg := configger.Defaults()
	db, err := database.Connect(cfg)
	assert.Nil(t, err)
	defer db.Close()
	err = text.Fix(db, io.Discard, configger.Config{})
	assert.NotNil(t, err)
	err = text.Fix(db, io.Discard, cfg)
	assert.Nil(t, err)
}
