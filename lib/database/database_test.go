package database

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gookit/color" //nolint:misspell
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
	color.Enable = false
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{[]byte("")}, ""},
		{"nottime", args{[]byte("hello world")}, "?"},
		{"invalid", args{[]byte("01-01-2000 00:00:00")}, "?"},
		{"old", args{[]byte("2000-01-01T00:00:00Z")}, "01 Jan 2000"},
		{"new", args{[]byte(fmt.Sprintf("%v-01-01T00:00:00Z", now.Year()))}, fmt.Sprintf("01 Jan %d", now.Year())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strings.TrimSpace(DateTime(tt.args.value)); got != tt.want {
				t.Errorf("DateTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_valid(t *testing.T) {
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
			got, err := valid(tt.args.deleted, tt.args.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("valid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reverseInt(t *testing.T) {
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
			if gotReversed, _ := reverseInt(tt.value); gotReversed != tt.wantReversed {
				t.Errorf("reverseInt() = %v, want %v", gotReversed, tt.wantReversed)
			}
		})
	}
}

func TestObfuscateParam(t *testing.T) {
	tests := []struct {
		name  string
		param string
		want  string
	}{
		{"empty", "", ""},
		{"0", "000001", "000001"},
		{"1", "1", "9b1c6"},
		{"999...", "999999999", "eb77359232"},
		{"rand", "69247541", "c06d44215"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ObfuscateParam(tt.param); got != tt.want {
				t.Errorf("ObfuscateParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVal(t *testing.T) {
	type args struct {
		col sql.RawBytes
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"null", args{nil}, null},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Val(tt.args.col); got != tt.want {
				t.Errorf("Val() = %v, want %v", got, tt.want)
			}
		})
	}
}
