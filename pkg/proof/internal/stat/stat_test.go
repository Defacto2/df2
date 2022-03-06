package stat_test

import (
	"database/sql"
	"testing"

	"github.com/Defacto2/df2/pkg/proof/internal/stat"
)

func TestProof_Summary(t *testing.T) {
	type fields struct {
		Base      string
		BasePath  string
		Columns   []string
		Count     int
		Missing   int
		Overwrite bool
		Total     int
		Values    *[]sql.RawBytes
	}
	tests := []struct {
		name   string
		fields fields
		id     string
		wantS  bool
	}{
		{"empty", fields{}, "", true},
		{"none", fields{}, "1", false},
		{"none", fields{Total: 5, Count: 4, Missing: 1}, "1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stat.Proof{
				Base:      tt.fields.Base,
				BasePath:  tt.fields.BasePath,
				Columns:   tt.fields.Columns,
				Count:     tt.fields.Count,
				Missing:   tt.fields.Missing,
				Overwrite: tt.fields.Overwrite,
				Total:     tt.fields.Total,
				Values:    tt.fields.Values,
			}
			if got := s.Summary(tt.id); len(got) > 0 != tt.wantS {
				t.Errorf("Proof.Summary() = %v, want %v", len(got) > 0, tt.wantS)
			}
		})
	}
}
