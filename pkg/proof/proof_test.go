package proof_test

import (
	"database/sql"
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/proof"
	"github.com/stretchr/testify/assert"
)

const uuid = "10000000-0000-0000-0000-000000000000"

func TestQuery(t *testing.T) {
	t.Parallel()
	r := proof.Request{}
	err := r.Query(nil, nil, configger.Config{}, "")
	assert.NotNil(t, err)
	cfg := configger.Defaults()
	db, err := database.Connect(cfg)
	assert.Nil(t, err)
	defer db.Close()
	err = r.Query(db, io.Discard, configger.Config{}, "")
	assert.NotNil(t, err)
	err = r.Query(db, io.Discard, configger.Config{}, "1")
	assert.NotNil(t, err)
	err = r.Query(db, io.Discard, cfg, "1")
	assert.Nil(t, err)
	err = r.Query(db, io.Discard, cfg, uuid)
	assert.Nil(t, err)
}

func Test_Select(t *testing.T) {
	t.Parallel()
	s := proof.Select("")
	assert.Contains(t, s, "FROM `files` WHERE `section` = 'releaseproof'")
	s = proof.Select("1")
	assert.Contains(t, s, `="1"`)
	s = proof.Select(uuid)
	assert.Contains(t, s, `uuid`)
}

func TestRequest_Skip(t *testing.T) {
	t.Parallel()
	type fields struct {
		Overwrite   bool
		All         bool
		HideMissing bool
		ByID        string
	}
	tests := []struct {
		name   string
		fields fields
		values []sql.RawBytes
		want   bool
	}{
		{"empty", fields{}, nil, true},
		{"false", fields{ByID: "1", Overwrite: true}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := proof.Request{
				Overwrite:   tt.fields.Overwrite,
				All:         tt.fields.All,
				HideMissing: tt.fields.HideMissing,
				ByID:        tt.fields.ByID,
			}
			if got, _ := request.Skip(nil, tt.values); got != tt.want {
				t.Errorf("Request.Skip() = %v, want %v", got, tt.want)
			}
		})
	}
}
