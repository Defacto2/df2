package fix_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo/internal/fix"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	err := fix.Configs(nil, nil)
	assert.NotNil(t, err)
	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = fix.Configs(db, io.Discard)
	assert.Nil(t, err)
}
