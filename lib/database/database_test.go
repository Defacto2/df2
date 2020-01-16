package database

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func TestIsID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"blank", args{""}, false},
		{"letters", args{"abcde"}, false},
		{"zeros", args{"00000"}, false},
		{"zeros", args{"00000876786"}, false},
		{"negative", args{"-1"}, false},
		{"valid 1", args{"1"}, true},
		{"valid 9", args{"99999"}, true},
		{"float", args{"1.0000"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsID(tt.args.id); got != tt.want {
				t.Errorf("IsID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUUID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{"x"}, false},
		{"", args{"0000"}, false},
		{"", args{""}, false},
		{"zeros", args{"00000000-0000-0000-0000-000000000000"}, true},
		{"random", args{uuid.New().String()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUUID(tt.args.id); got != tt.want {
				t.Errorf("IsUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateTime(t *testing.T) {
	type args struct {
		value sql.RawBytes
	}
	now := time.Now()
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{[]byte("")}, "?"},
		{"nottime", args{[]byte("hello world")}, "?"},
		{"invalid", args{[]byte("01-01-2000 00:00:00")}, "?"},
		{"old", args{[]byte("2000-01-01T00:00:00Z")}, "01 Jan 2000  "},
		{"new", args{[]byte(fmt.Sprintf("%v-01-01T00:00:00Z", now.Year()))}, "01 Jan 00:00 "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DateTime(tt.args.value); got != tt.want {
				t.Errorf("DateTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_isNew(t *testing.T) {
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
			got, err := isNew(tt.args.deleted, tt.args.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("isNew() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isNew() = %v, want %v", got, tt.want)
			}
		})
	}
}
