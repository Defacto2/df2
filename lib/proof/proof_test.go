package proof

import (
	"database/sql"
	"testing"
	"time"
)

const uuid = "10000000-0000-0000-0000000000000000"

func TestRequest_Query(t *testing.T) {
	type fields struct {
		Overwrite   bool
		AllProofs   bool
		HideMissing bool
	}
	no := fields{false, false, false}
	tests := []struct {
		name    string
		id      string
		fields  fields
		wantErr bool
	}{
		{"empty", "", no, true},
		{"missing", "1", no, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := Request{
				Overwrite:   tt.fields.Overwrite,
				AllProofs:   tt.fields.AllProofs,
				HideMissing: tt.fields.HideMissing,
			}
			if err := request.Query(tt.id); (err != nil) != tt.wantErr {
				t.Errorf("Request.Query() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlSelect(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int
	}{
		{"empty", "", 141},
		{"id", "1", 154},
		{"uuid", uuid, 141},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqlSelect(tt.id); len(got) != tt.want {
				t.Errorf("sqlSelect() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func Test_stat_summary(t *testing.T) {
	type fields struct {
		base      string
		basePath  string
		columns   []string
		count     int
		missing   int
		overwrite bool
		start     time.Time
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
			s := stat{
				base:      tt.fields.base,
				basePath:  tt.fields.basePath,
				columns:   tt.fields.columns,
				count:     tt.fields.count,
				missing:   tt.fields.missing,
				overwrite: tt.fields.overwrite,
				start:     tt.fields.start,
				total:     tt.fields.total,
				values:    tt.fields.values,
			}
			s.summary("")
		})
	}
}

func Test_val(t *testing.T) {
	type args struct {
		col sql.RawBytes
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"null", args{nil}, "NULL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := val(tt.args.col); got != tt.want {
				t.Errorf("val() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stat_fileSkip(t *testing.T) {
	type fields struct {
		base      string
		basePath  string
		columns   []string
		count     int
		missing   int
		overwrite bool
		start     time.Time
		total     int
		values    *[]sql.RawBytes
	}
	type args struct {
		r    Record
		hide bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantSkip bool
	}{
		{"", fields{}, args{Record{}, false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stat{
				base:      tt.fields.base,
				basePath:  tt.fields.basePath,
				columns:   tt.fields.columns,
				count:     tt.fields.count,
				missing:   tt.fields.missing,
				overwrite: tt.fields.overwrite,
				start:     tt.fields.start,
				total:     tt.fields.total,
				values:    tt.fields.values,
			}
			if gotSkip, gotErr := s.fileSkip(tt.args.r, tt.args.hide); gotSkip != tt.wantSkip {
				t.Errorf("stat.fileSkip() = %v, want %v", gotSkip, tt.wantSkip)
			} else if gotErr != nil {
				t.Error("gotErr = ", gotErr)
			}
		})
	}
}
