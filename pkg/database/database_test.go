package database

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gookit/color"
)

const (
	uuid1 = "c8cd0b4c-2f54-11e0-8827-cc1607e15609"
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

func TestCheckID(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr bool
	}{
		{"blank", "", true},
		{"alpha", "abc", true},
		{"one", "1", false},
		{"uuid", uuid1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckID(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("CheckID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckUUID(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr bool
	}{
		{"blank", "", true},
		{"alpha", "abc", true},
		{"one", "1", true},
		{"uuid", uuid1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckUUID(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("CheckUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileUpdate(t *testing.T) {
	type args struct {
		name string
		db   time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"empty", args{}, true, false},
		{"test", args{"database_test.go", time.Now()}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileUpdate(tt.args.name, tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetID(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    uint
		wantErr bool
	}{
		{"blank", "", 0, true},
		{"txt", "invalid", 0, true},
		{"one", "1", 1, false},
		{"uuid", uuid1, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetID(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	const f = "Defacto2_ISO_2007.7z"
	tests := []struct {
		name    string
		s       string
		want    string
		wantErr bool
	}{
		{"blank", "", "", true},
		{"txt", "invalid", "", true},
		{"one", "1", f, false},
		{"uuid", uuid1, f, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFile(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_StripChars(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"ooÖØöøO"}, "ooÖØöøO"},
		{"", args{"o.o|Ö+Ø=ö^ø#O"}, "ooÖØöøO"},
		{"", args{"A Café!"}, "A Café"},
		{"", args{"brunräven - över"}, "brunräven - över"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripChars(tt.args.s); got != tt.want {
				t.Errorf("StripChars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_StripStart(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"hello world"}, "hello world"},
		{"", args{"--argument"}, "argument"},
		{"", args{"!!!OMG-WTF"}, "OMG-WTF"},
		{"", args{"#ÖØöøO"}, "ÖØöøO"},
		{"", args{"!@#$%^&A(+)ooÖØöøO"}, "A(+)ooÖØöøO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripStart(tt.args.s); got != tt.want {
				t.Errorf("StripStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TrimSP(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"abc"}, "abc"},
		{"", args{"a b c"}, "a b c"},
		{"", args{"a  b  c"}, "a b c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimSP(tt.args.s); got != tt.want {
				t.Errorf("TrimSP() = %v, want %v", got, tt.want)
			}
		})
	}
}
