package zipcmmt_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/zipcmmt"
	"github.com/stretchr/testify/assert"
)

func TestFix(t *testing.T) {
	t.Parallel()
	const (
		ascii     = false
		unicode   = false
		overwrite = false
		summary   = false
	)
	err := zipcmmt.Fix(nil, nil, configger.Defaults(),
		unicode, overwrite, summary)
	assert.NotNil(t, err)
	cfg := configger.Defaults()
	db, err := database.Connect(cfg)
	assert.Nil(t, err)
	defer db.Close()
	err = zipcmmt.Fix(db, io.Discard, cfg,
		unicode, overwrite, summary)
	assert.Nil(t, err)
}
