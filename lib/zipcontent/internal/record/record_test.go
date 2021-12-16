package record_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Defacto2/df2/lib/zipcontent/internal/record"
	"github.com/Defacto2/df2/lib/zipcontent/internal/scan"
)

func TestNew(t *testing.T) {
	const id, uuid, filename, readme = 0, 1, 4, 6
	var empty record.Record
	mock := make([]sql.RawBytes, 7)
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

func TestIterate(t *testing.T) {
	var empty record.Record
	mockV := []sql.RawBytes{
		sql.RawBytes("0"),
		sql.RawBytes(time.Now().String()),
		sql.RawBytes("somefile.zip"),
		sql.RawBytes(""),
	}
	mockR := record.Record{
		ID: "1",
	}
	mockSBad := scan.Stats{
		Columns: []string{"1ee21218-5898-11ec-bf63-0242ac130002"},
		Values:  &mockV,
	}
	mockS := scan.Stats{
		Columns: []string{"id", "createdat", "filename", "uuid"},
		Values:  &mockV,
	}
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
		{"no stats", fields(mockR), args{}, true},
		{"no fields", fields(empty), args{&mockSBad}, true},
		{"missing cols", fields(mockR), args{&mockSBad}, true},
		// we want an error because it attempts to read a non-existent file.
		{"okay, but want error", fields(mockR), args{&mockS}, true},
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
				t.Errorf("Iterate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecord_Save(t *testing.T) {
	type fields struct {
		ID    string
		UUID  string
		File  string
		Name  string
		Files []string
		NFO   string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{"empty", fields{}, 0, true},
		{"bad id", fields{ID: "abcde"}, 0, false},
		{"good id", fields{ID: "1"}, 1, false},
		// use the time now value to force an update to the database.
		{"nfo", fields{ID: "1", NFO: "sometextfile-" + time.Now().String() + ".txt"}, 1, false},
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
			got, err := r.Save()
			if (err != nil) != tt.wantErr {
				t.Errorf("Record.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Record.Save() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_Nfo(t *testing.T) {
	const fn = "test.zip"
	files := []string{"prog.exe", "data.1", "data.2", "data.3", "prog.txt", "test.txt", "test.nfo"}
	mockS := scan.Stats{
		BasePath: "",
	}
	type fields struct {
		ID    string
		UUID  string
		File  string
		Name  string
		Files []string
		NFO   string
	}
	tests := []struct {
		name    string
		fields  fields
		s       *scan.Stats
		wantErr bool
	}{
		{"empty", fields{}, nil, true},
		// this test will fail when attempting to extract a non-existent file
		{"fake file", fields{File: fn, Files: files}, &mockS, true},
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
			if err := r.Nfo(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("Record.Nfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecord_Read(t *testing.T) {
	const (
		dir  = "../../../../tests/demozoo"
		fn   = "test.zip"
		uuid = "c1fab556-58a0-11ec-bf63-0242ac130002"
	)
	mockS := scan.Stats{
		BasePath: dir,
	}
	type fields struct {
		ID    string
		UUID  string
		File  string
		Name  string
		Files []string
		NFO   string
	}
	tests := []struct {
		name    string
		fields  fields
		s       *scan.Stats
		wantErr bool
	}{
		{"empty", fields{}, nil, true},
		{"test.zip", fields{
			ID:   "1",
			UUID: uuid,
			File: filepath.Join(dir, fn),
			Name: fn,
		}, &mockS, false},
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
			if err := r.Read(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("Record.Read() error = %v, wantErr %v", err, tt.wantErr)
			} else if tt.name == "test.zip" && err == nil {
				defer os.Remove(filepath.Join(dir, uuid+".txt"))
			}
		})
	}
}
