package input_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/prompt/internal/input"
	"github.com/gookit/color"
)

func TestDir(t *testing.T) {
	d, err := filepath.Abs("../../../../tests/empty")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		r       io.Reader
		want    string
		wantErr bool
	}{
		{"empty", nil, "", true},
		{"enter key", strings.NewReader(""), "", false},
		{"skip", strings.NewReader("-"), "", true},
		{"okay", strings.NewReader(d), d, false},
	}
	if _, err := os.Stat(d); os.IsNotExist(err) {
		os.Mkdir(d, 0755)
		defer os.Remove(d)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := input.Dir(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Dir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRead(t *testing.T) {
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
			gotInput, err := input.Read(&stdin)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotInput != tt.wantInput {
				t.Errorf("Read() = %v, want %v", gotInput, tt.wantInput)
			}
		})
	}
}

func TestYN(t *testing.T) {
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
			if got := input.YN(tt.input, tt.def); got != tt.want {
				t.Errorf("YN(%v) = %v, want %v", tt.def, got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		want    string
		wantErr bool
	}{
		{"empty", nil, "", true},
		{"enter key", strings.NewReader(""), "", true},
		{"random", strings.NewReader("random"), "random", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := input.String(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPort(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		want    int64
		wantErr bool
	}{
		{"empty", nil, 0, true},
		{"string", strings.NewReader("hello"), 0, false},
		{"too low", strings.NewReader("-1"), 0, false},
		{"too high", strings.NewReader("1000000"), 0, false},
		{"lowest", strings.NewReader("0"), 0, false},
		{"common", strings.NewReader("8080"), 8080, false},
		{"multiple tries", strings.NewReader("\n\n\n\n"), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := input.Port(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Port() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Port() = %v, want %v", got, tt.want)
			}
		})
	}
}
