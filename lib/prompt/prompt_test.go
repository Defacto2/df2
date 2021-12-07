package prompt

import (
	"bytes"
	"testing"

	"github.com/gookit/color"
)

func TestYN(t *testing.T) {
	type args struct {
		query string
		yes   bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"y", args{"", false}, false},
		{"n", args{"", true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := YN(tt.args.query, tt.args.yes); got != tt.want {
				t.Errorf("YN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_read(t *testing.T) {
	color.Enable = false
	var stdin bytes.Buffer
	tests := []struct {
		name      string
		in        string
		wantInput string
		wantErr   bool
	}{
		{"empty", "", "", false},
		{"hello", "hello", "hello", false},
		{"trim", "        hello", "hello", false},
		{"sentence", "I am hello world.", "I am hello world.", false},
		{"nl", "\n\t\n\t\tb", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin.Write([]byte(tt.in))
			gotInput, err := read(&stdin)
			if (err != nil) != tt.wantErr {
				t.Errorf("read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotInput != tt.wantInput {
				t.Errorf("read() = %v, want %v", gotInput, tt.wantInput)
			}
		})
	}
}

func Test_parseyn(t *testing.T) {
	tests := []struct {
		name  string
		input string
		def   bool
		want  bool
	}{
		{"default y", "", true, true},
		{"default n", "", false, false},
		{"yes", "y", true, true},
		{"no", "n", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseyn(tt.input, tt.def); got != tt.want {
				t.Errorf("parseyn(%v) = %v, want %v", tt.def, got, tt.want)
			}
		})
	}
}

func Test_port(t *testing.T) {
	tests := []struct {
		name string
		port int
		want bool
	}{
		{"-1", -1, false},
		{"million", 1000000, false},
		{"zero", 0, true},
		{"1024", 1024, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := port(tt.port); got != tt.want {
				t.Errorf("Port() = %v, want %v", got, tt.want)
			}
		})
	}
}
