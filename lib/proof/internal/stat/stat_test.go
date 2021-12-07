package stat_test

import (
	"database/sql"
	"testing"

	"github.com/Defacto2/df2/lib/proof/internal/stat"
)

func TestSummary(t *testing.T) {
	type fields struct {
		base      string
		basePath  string
		columns   []string
		count     int
		missing   int
		overwrite bool
		total     int
		values    *[]sql.RawBytes
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"", fields{total: 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := stat.Stat{
				Base:      tt.fields.base,
				BasePath:  tt.fields.basePath,
				Columns:   tt.fields.columns,
				Count:     tt.fields.count,
				Missing:   tt.fields.missing,
				Overwrite: tt.fields.overwrite,
				Total:     tt.fields.total,
				Values:    tt.fields.values,
			}
			s.Summary("")
		})
	}
}
