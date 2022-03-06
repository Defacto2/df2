package export_test

import (
	"os"
	"testing"

	"github.com/Defacto2/df2/pkg/database/internal/export"
)

func TestFlags_Run(t *testing.T) { //nolint:funlen
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
			Limit:  1,
		}, false},
		{"users parallel", fields{
			Type:     "c",
			Tables:   export.Users.String(),
			Parallel: true,
			Limit:    1,
		}, false},
		{"net parallel", fields{
			Type:     "c",
			Tables:   export.Netresources.String(),
			Parallel: true,
			Limit:    1,
		}, false},
		{"groups parallel", fields{
			Type:     "c",
			Tables:   export.Groups.String(),
			Parallel: true,
			Limit:    1,
		}, false},
		{"update groups parallel", fields{
			Type:     "update",
			Tables:   export.Groups.String(),
			Parallel: true,
			Limit:    1,
		}, false},
	}
	rm := []string{
		"d2-create_files.sql.bz2",
		"d2-create_groups.sql.bz2",
		"d2-create_netresources.sql.bz2",
		"d2-create_table.sql.bz2",
		"d2-update_groups.sql.bz2",
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
			for _, name := range rm {
				os.Remove(name)
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
