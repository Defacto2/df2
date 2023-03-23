package file_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/recent/internal/file"
	"github.com/stretchr/testify/assert"
)

const uuid = "d37e5b5f-f5bf-4138-9078-891e41b10a12"

func TestThumb_Scan(t *testing.T) {
	t.Parallel()
	f := file.Thumb{}
	f.Scan(nil)
	assert.Equal(t, "", f.URLID)
	n := time.Now().Format(time.RFC3339)
	v := []sql.RawBytes{
		sql.RawBytes("1"),
		sql.RawBytes(uuid),
		sql.RawBytes("Placeholder title"),
		sql.RawBytes("For some group"),
		sql.RawBytes("By some group"),
		sql.RawBytes("file.txt"),
		sql.RawBytes("1990"),
		sql.RawBytes(n),
	}
	f.Scan(v)
	assert.NotEmpty(t, f)
}
