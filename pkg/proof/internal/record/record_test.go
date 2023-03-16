package record_test

import (
	"database/sql"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/proof/internal/record"
	"github.com/Defacto2/df2/pkg/proof/internal/stat"
)

const uuid = "d37e5b5f-f5bf-4138-9078-891e41b10a12"

func TestNew(t *testing.T) {
	type args struct {
		values []sql.RawBytes
		path   string
	}
	tests := []struct {
		name string
		args args
		want record.Record
	}{
		{"empty", args{}, record.Record{}},
		{"okay", args{
			path: "someDir",
			values: []sql.RawBytes{
				sql.RawBytes("1"),
				sql.RawBytes(uuid),
				sql.RawBytes("placeholder"),
				sql.RawBytes("placeholder"),
				sql.RawBytes("file.txt"),
			},
		}, record.Record{"1", uuid, "someDir/" + uuid, "file.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := record.New(tt.args.values, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_Approve(t *testing.T) {
	type fields struct {
		ID   string
		UUID string
		File string
		Name string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty", fields{}, true},
		{"okay", fields{
			ID:   "1",
			UUID: uuid,
			File: filepath.Join("someDir", uuid),
			Name: "file.zip",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := record.Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
				File: tt.fields.File,
				Name: tt.fields.Name,
			}
			if err := r.Approve(nil, nil); (err != nil) != tt.wantErr {
				t.Errorf("Record.Approve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecord_Iterate(t *testing.T) {
	type fields struct {
		ID   string
		UUID string
		File string
		Name string
	}
	rec1 := fields{
		ID:   "1",
		UUID: uuid,
		File: "someDir/file.txt",
		Name: "file.txt",
	}
	stat1 := stat.Proof{
		Columns: []string{},
	}
	vals2 := []sql.RawBytes{
		sql.RawBytes("1"),
		sql.RawBytes(time.Now().String()),
		sql.RawBytes("file.txt"),
		sql.RawBytes("readme.txt,prog.exe"),
	}
	stat2 := stat.Proof{
		Columns: []string{"id", "createdat", "filename", "file_zip_content"},
		Values:  &vals2,
	}
	type args struct {
		s stat.Proof
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"empty", fields{}, args{}, true},
		{"no stat", rec1, args{}, true},
		{"empty vals", rec1, args{stat1}, false},
		{"okay", rec1, args{stat2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := record.Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
				File: tt.fields.File,
				Name: tt.fields.Name,
			}
			if err := r.Iterate(nil, nil, nil, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Record.Iterate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSkip(t *testing.T) {
	type args struct {
		s    stat.Proof
		r    record.Record
		hide bool
	}
	rec1 := record.Record{
		ID:   "1",
		UUID: uuid,
		File: "someDir/file.txt",
		Name: "file.txt",
	}
	stat1 := stat.Proof{
		Columns: []string{},
	}
	tests := []struct {
		name     string
		args     args
		wantSkip bool
		wantErr  bool
	}{
		{"empty", args{}, false, true},
		{"missing", args{stat1, rec1, false}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSkip, err := record.Skip(nil, tt.args.s, tt.args.r, tt.args.hide)
			if (err != nil) != tt.wantErr {
				t.Errorf("Skip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSkip != tt.wantSkip {
				t.Errorf("Skip() = %v, want %v", gotSkip, tt.wantSkip)
			}
		})
	}
}

func TestRecord_Zip(t *testing.T) {
	type fields struct {
		ID   string
		UUID string
		File string
		Name string
	}
	type args struct {
		col sql.RawBytes
		s   *stat.Proof
	}
	okay := stat.Proof{
		Overwrite: true,
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"empty", fields{}, args{}, true},
		{"ok", fields{
			Name: "file.txt",
			UUID: uuid,
		}, args{col: nil, s: &okay}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := record.Record{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
				File: tt.fields.File,
				Name: tt.fields.Name,
			}
			if err := r.Zip(nil, nil, nil, tt.args.col, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Record.Zip() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateZipContent(t *testing.T) {
	type args struct {
		id       string
		filename string
		content  string
		items    int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, false},
		{"okay", args{
			id:       "1",
			filename: "",
			content:  "somefile.txt",
			items:    1,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := record.UpdateZipContent(nil, nil, tt.args.id, tt.args.filename, tt.args.content, tt.args.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateZipContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
