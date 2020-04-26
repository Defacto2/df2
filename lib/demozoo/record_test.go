package demozoo

import (
	"reflect"
	"testing"
	"time"
)

func TestRecord_sql(t *testing.T) {
	type fields struct {
		count          int
		FilePath       string
		ID             string
		UUID           string
		WebIDDemozoo   uint
		WebIDPouet     uint
		Filename       string
		Filesize       string
		FileZipContent string
		CreatedAt      string
		UpdatedAt      string
		SumMD5         string
		Sum384         string
		LastMod        time.Time
		Readme         string
		DOSeeBinary    string
		Platform       string
		GroupFor       string
		GroupBy        string
	}
	const where string = " WHERE id=?"
	var now = time.Now()
	tests := []struct {
		name   string
		fields fields
		want   string
		want1  int
	}{
		{name: "empty", fields: fields{}, want: "", want1: 0},
		{"filename", fields{ID: "1", Filename: "hi.txt"}, "UPDATE files SET filename=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5},
		{"filesize", fields{ID: "1", Filesize: "54321"}, "UPDATE files SET filesize=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5},
		{"zip content", fields{ID: "1", FileZipContent: "HI.TXT\nHI.EXE"}, "UPDATE files SET file_zip_content=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5},
		{"md5", fields{ID: "1", SumMD5: "md5placeholder"}, "UPDATE files SET file_integrity_weak=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5},
		{"sha386", fields{ID: "1", Sum384: "shaplaceholder"}, "UPDATE files SET file_integrity_strong=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5},
		{"lastmod", fields{ID: "1", LastMod: now}, "UPDATE files SET file_last_modified=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 5},
		{"a file", fields{ID: "1", Filename: "some.gif", Filesize: "5012352"}, "UPDATE files SET filename=?,filesize=?,web_id_demozoo=?,updatedat=?,updatedby=?" + where, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Record{
				count:          tt.fields.count,
				FilePath:       tt.fields.FilePath,
				ID:             tt.fields.ID,
				UUID:           tt.fields.UUID,
				WebIDDemozoo:   tt.fields.WebIDDemozoo,
				WebIDPouet:     tt.fields.WebIDPouet,
				Filename:       tt.fields.Filename,
				Filesize:       tt.fields.Filesize,
				FileZipContent: tt.fields.FileZipContent,
				CreatedAt:      tt.fields.CreatedAt,
				UpdatedAt:      tt.fields.UpdatedAt,
				SumMD5:         tt.fields.SumMD5,
				Sum384:         tt.fields.Sum384,
				LastMod:        tt.fields.LastMod,
				Readme:         tt.fields.Readme,
				DOSeeBinary:    tt.fields.DOSeeBinary,
				Platform:       tt.fields.Platform,
				GroupFor:       tt.fields.GroupFor,
				GroupBy:        tt.fields.GroupBy,
			}
			got, got1 := r.sql()
			if got != tt.want {
				t.Errorf("Record.sql() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(len(got1), tt.want1) {
				t.Errorf("Record.sql() got1 = %v, want %v", len(got1), tt.want1)
			}
		})
	}
}
