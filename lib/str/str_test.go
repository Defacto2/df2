package str

import (
	"testing"

	"github.com/gookit/color" //nolint:misspell
)

func capString(test, text string) (output string) {
	color.Enable = false
	switch test {
	case "sec":
		output = Sec(text)
	case "warn":
		output = Warn(text)
	case "x":
		output = X()
	case "y":
		output = Y()
	}
	return output
}

func Test_capString(t *testing.T) {
	type args struct {
		test string
		text string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
	}{
		{"sec empty", args{"sec", ""}, ""},
		{"sec", args{"sec", "hello"}, "hello"},
		{"warn empty", args{"warn", ""}, ""},
		{"warn", args{"warn", "hello"}, "hello"},
		{"x", args{"x", ""}, "✗"},
		{"y", args{"y", ""}, "✓"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOutput := capString(tt.args.test, tt.args.text); gotOutput != tt.wantOutput {
				t.Errorf("capString() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func Test_Progress(t *testing.T) {
	type args struct {
		name  string
		count int
		total int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"ten", args{"", 1, 10}, float64(10)},
		{"hundred", args{"", 10, 10}, float64(100)},
		{"zero", args{"", 0, 10}, float64(0)},
		{"negative", args{"", -1, 10}, float64(-10)},
		{"decimal", args{"", 1, 99999}, float64(0.001000010000100001)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Progress(tt.args.name, tt.args.count, tt.args.total); got != tt.want {
				t.Errorf("Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	type args struct {
		text string
		len  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", 0}, ""},
		{"zero", args{"hello", 0}, "hello"},
		{"minus", args{"hello", -10}, "hello"},
		{"three", args{"hello", 3}, "he…"},
		{"too long", args{"hello", 600}, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.args.text, tt.args.len); got != tt.want {
				t.Errorf("Truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}
