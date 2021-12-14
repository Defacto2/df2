package recd_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/lib/database/internal/recd"
)

func Test_record_imagePath(t *testing.T) {
	type fields struct {
		uuid string
	}
	type args struct {
		path string
	}
	const (
		png  = ".png"
		uuid = "486070ae-f462-446f-b7e8-c70cb7a8a996"
	)
	p := os.TempDir()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"empty", fields{uuid}, args{""}, uuid + png},
		{"ok", fields{uuid}, args{p}, filepath.Join(p, uuid) + png},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := recd.Record{
				UUID: tt.fields.uuid,
			}
			if got := r.ImagePath(tt.args.path); got != tt.want {
				t.Errorf("ImagePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckGroups(t *testing.T) {
	type args struct {
		g1 string
		g2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{"", ""}, false},
		{"", args{"CHANGEME", ""}, false},
		{"", args{"", "Changeme"}, false},
		{"", args{"A group", ""}, true},
		{"", args{"", "A group"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := recd.Record{}
			if got := r.CheckGroups(tt.args.g1, tt.args.g2); got != tt.want {
				t.Errorf("CheckGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValid(t *testing.T) {
	type args struct {
		deleted sql.RawBytes
		updated sql.RawBytes
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"new", args{[]byte("2006-01-02T15:04:05Z"), []byte("2006-01-02T15:04:05Z")}, true, false},
		{"new+offset", args{[]byte("2006-01-02T15:04:06Z"), []byte("2006-01-02T15:04:05Z")}, true, false},
		{"old del", args{[]byte("2016-01-02T15:04:05Z"), []byte("2006-01-02T15:04:05Z")}, false, false},
		{"old upd", args{[]byte("2000-01-02T15:04:05Z"), []byte("2016-01-02T15:04:05Z")}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := recd.Valid(tt.args.deleted, tt.args.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("Valid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverseInt(t *testing.T) {
	tests := []struct {
		name         string
		value        uint
		wantReversed uint
	}{
		{"empty", 0, 0},
		{"count", 12345, 54321},
		{"seq", 555, 555},
		{"sign", 662211, 112266},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotReversed, _ := recd.ReverseInt(tt.value); gotReversed != tt.wantReversed {
				t.Errorf("ReverseInt() = %v, want %v", gotReversed, tt.wantReversed)
			}
		})
	}
}
