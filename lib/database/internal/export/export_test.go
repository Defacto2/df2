package export_test

import (
	"testing"

	"github.com/Defacto2/df2/lib/database/internal/export"
)

func TestFlags_Run(t *testing.T) {
	type fields struct {
		Parallel bool
		Tables   string
		Type     string
		Limit    uint
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty", fields{}, true},
		{"no table", fields{Type: "c"}, true},
		{"users", fields{
			Type:   "c",
			Tables: export.Users.String(),
			Limit:  1}, false},
		{"parallel", fields{
			Type:     "c",
			Tables:   export.Users.String(),
			Parallel: true,
			Limit:    1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &export.Flags{
				Parallel: tt.fields.Parallel,
				Tables:   tt.fields.Tables,
				Type:     tt.fields.Type,
				Limit:    tt.fields.Limit,
			}
			if err := f.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Flags.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlags_DB(t *testing.T) {
	type fields struct {
		Parallel bool
		Tables   string
		Type     string
		Limit    uint
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty", fields{}, true},
		{"export", fields{
			Type:     "c",
			Tables:   export.Users.String(),
			Parallel: false,
			Limit:    1,
		}, false},
		{"export", fields{
			Type:     "c",
			Tables:   export.Users.String(),
			Parallel: true,
			Limit:    1,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &export.Flags{
				Parallel: tt.fields.Parallel,
				Tables:   tt.fields.Tables,
				Type:     tt.fields.Type,
				Limit:    tt.fields.Limit,
			}
			if err := f.DB(); (err != nil) != tt.wantErr {
				t.Errorf("Flags.DB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
