package cmmt_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/zipcmmt/internal/cmmt"
)

const (
	path = "../../../../tests/uuid"
	uuid = "ef73b9dc-58b5-11ec-bf63-0242ac130002"
)

func TestZipfile_Checks(t *testing.T) {
	type fields struct {
		UUID    string
		ASCII   bool
		Unicode bool
	}
	tests := []struct {
		name   string
		fields fields
		path   string
		wantOk bool
	}{
		{"empty", fields{}, "", false},
		{"okay", fields{UUID: uuid, ASCII: true, Unicode: true}, path, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &cmmt.Zipfile{
				UUID:    tt.fields.UUID,
				ASCII:   tt.fields.ASCII,
				Unicode: tt.fields.Unicode,
			}
			if gotOk := z.CheckDownload(tt.path); gotOk != tt.wantOk {
				t.Errorf("Zipfile.CheckDownload() = %v, want %v", gotOk, tt.wantOk)
			}
			// the gotOk == wantOk is intentional.
			if gotOk := z.CheckCmmtFile(tt.path); gotOk == tt.wantOk {
				t.Errorf("Zipfile.CheckCmmtFile() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestZipfile_Save(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty", "", true},
		{"ok", filepath.Join(path, uuid), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &cmmt.Zipfile{}
			if err := z.Save(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("Zipfile.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestZipfile_Print(t *testing.T) {
	type fields struct {
		ID    uint
		Name  string
		Magic sql.NullString
	}
	tests := []struct {
		name   string
		fields fields
		cmmt   *string
		want   string
	}{
		{"empty", fields{}, nil, ""},
		{"ok", fields{
			ID:    1,
			Name:  "somefile.txt",
			Magic: sql.NullString{String: "text/plain", Valid: true},
		}, nil, "1. - somefile.txt [text/plain]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &cmmt.Zipfile{
				ID:    tt.fields.ID,
				Name:  tt.fields.Name,
				Magic: tt.fields.Magic,
			}
			if got := strings.TrimSpace(z.Print(tt.cmmt)); got != tt.want {
				t.Errorf("Zipfile.Print() = %q, want %q",
					got, tt.want)
			}
		})
	}
}
