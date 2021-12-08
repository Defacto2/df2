package record_test

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/lib/zipcontent/internal/record"
	"github.com/Defacto2/df2/lib/zipcontent/internal/scan"
)

func TestNew(t *testing.T) {
	const id, uuid, filename, readme = 0, 1, 4, 6
	var empty record.Record
	var mock = make([]sql.RawBytes, 7)
	const ids, uuids = "345674", "b4ef0174-57b4-11ec-bf63-0242ac130002"
	mock[id] = sql.RawBytes(ids)
	mock[uuid] = sql.RawBytes(uuids)
	mock[filename] = sql.RawBytes("somefile.zip")
	mock[readme] = sql.RawBytes("readme.txt")
	want := record.Record{
		ID:    ids,
		UUID:  uuids,
		File:  "dir/" + uuids,
		Name:  "somefile.zip",
		Files: nil,
		NFO:   "readme.txt",
	}
	type args struct {
		values []sql.RawBytes
		path   string
	}
	tests := []struct {
		name    string
		args    args
		want    record.Record
		wantErr bool
	}{
		{"empty", args{}, empty, true},
		{"mock", args{mock, "dir"}, want, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := record.New(tt.args.values, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_Iterate(t *testing.T) {
	var empty record.Record
	type fields struct {
		ID    string
		UUID  string
		File  string
		Name  string
		Files []string
		NFO   string
	}
	type args struct {
		s *scan.Stats
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"empty", fields(empty), args{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &record.Record{
				ID:    tt.fields.ID,
				UUID:  tt.fields.UUID,
				File:  tt.fields.File,
				Name:  tt.fields.Name,
				Files: tt.fields.Files,
				NFO:   tt.fields.NFO,
			}
			if err := r.Iterate(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Record.Iterate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
