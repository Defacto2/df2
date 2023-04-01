package export_test

import (
	"io"
	"os"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/database/internal/export"
	"github.com/stretchr/testify/assert"
)

func TestFlags_Run(t *testing.T) {
	t.Parallel()
	f := export.Flags{}
	err := f.Run(nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = f.Run(db, io.Discard)
	assert.NotNil(t, err)

	f = export.Flags{
		Type: "c",
	}
	err = f.Run(db, io.Discard)
	assert.NotNil(t, err)

	f = export.Flags{
		Type:     "c",
		Tables:   export.Netresources.String(),
		Parallel: true,
		Limit:    1,
	}
	err = f.Run(db, io.Discard)
	assert.Nil(t, err)

	f = export.Flags{
		Type:     "update",
		Tables:   export.Netresources.String(),
		Parallel: true,
		Limit:    1,
	}
	err = f.Run(db, io.Discard)
	assert.Nil(t, err)

	rm := []string{
		"d2-create_table.sql.bz2",
		"d2-update_table.sql.bz2",
	}
	for _, name := range rm {
		os.Remove(name)
	}
}

func TestFlags_DB(t *testing.T) {
	t.Parallel()
	f := export.Flags{}
	err := f.DB(nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = f.DB(db, io.Discard)
	assert.NotNil(t, err)

	f = export.Flags{
		Type:     "c",
		Tables:   export.Netresources.String(),
		Parallel: false,
		Limit:    1,
	}
	err = f.DB(db, io.Discard)
	assert.Nil(t, err)
	f = export.Flags{
		Type:     "c",
		Tables:   export.Netresources.String(),
		Parallel: true,
		Limit:    1,
	}
	err = f.DB(db, io.Discard)
	assert.Nil(t, err)
}
