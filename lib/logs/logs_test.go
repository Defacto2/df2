package logs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/gookit/color.v1"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "../../tests/", name)
}

func testTxt() string {
	return filepath.Join(testDir("logs"), "test.log")
}

func Test_ProgressPct(t *testing.T) {
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
		{"", args{"", 1, 10}, float64(10)},
		{"", args{"", 10, 10}, float64(100)},
		{"", args{"", 0, 10}, float64(0)},
		{"", args{"", -1, 10}, float64(-10)},
		{"", args{"", 1, 99999}, float64(0.001000010000100001)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProgressPct(tt.args.name, tt.args.count, tt.args.total); got != tt.want {
				t.Errorf("ProgressPct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_save(t *testing.T) {
	const msg = "test error: this is a test"
	err := fmt.Errorf(msg)
	type args struct {
		err  error
		path string
	}
	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{
		{"empty", args{nil, ""}, false},
		{"empty", args{fmt.Errorf(""), ""}, false},
		{"ok", args{err, testTxt()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := save(tt.args.err, tt.args.path); gotOk != tt.wantOk {
				t.Errorf("save() = %v, want %v", gotOk, tt.wantOk)
			} else if gotOk {
				// cleanup
				if err := os.Remove(testTxt()); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestFilepath(t *testing.T) {
	t.Run("file path", func(t *testing.T) {
		if got := Filepath(); got == "" {
			t.Errorf("Filepath() = %q, want a directory path", got)
		}
	})
}

func capture(test, text string, quiet bool) (output string) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	Quiet = false
	if quiet {
		Quiet = true
	}
	switch test {
	case "print":
		Print(text)
	case "printcr":
		Printcr(text)
	case "printf":
		Printf("%s", text)
	case "println":
		Println(text)
	case "printfcr":
		Printfcr("%s", text)
	}
	w.Close()
	bytes, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	return strings.TrimSpace(string(bytes))
}

func TestPrints(t *testing.T) {
	type args struct {
		test  string
		text  string
		quiet bool
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
	}{
		{"print", args{"print", "", false}, ""},
		{"print hello", args{"print", "hello", false}, "hello"},
		{"print !hello", args{"print", "hello", true}, ""},
		{"cr", args{"printcr", "", false}, ""},
		{"cr hello", args{"printcr", "hello", false}, "hello"},
		{"cr !hello", args{"printcr", "hello", true}, ""},
		{"f", args{"printf", "", false}, ""},
		{"f hello", args{"printf", "hello", false}, "hello"},
		{"f !hello", args{"printf", "hello", true}, ""},
		{"ln", args{"println", "", false}, ""},
		{"ln hello", args{"println", "hello", false}, "hello"},
		{"ln !hello", args{"println", "hello", true}, ""},
		{"fcr", args{"printfcr", "", false}, ""},
		{"fcr hello", args{"printfcr", "hello", false}, "hello"},
		{"fcr !hello", args{"printfcr", "hello", true}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOutput := capture(tt.args.test, tt.args.text, tt.args.quiet); gotOutput != tt.wantOutput {
				t.Errorf("capture() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestProgressSum(t *testing.T) {
	type args struct {
		count int
		total int
	}
	tests := []struct {
		name    string
		args    args
		wantSum string
	}{
		{"empty", args{}, "0/0"},
		{"zero", args{0, 0}, "0/0"},
		{"1%", args{1, 100}, "1/100"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSum := ProgressSum(tt.args.count, tt.args.total); gotSum != tt.wantSum {
				t.Errorf("ProgressSum() = %v, want %v", gotSum, tt.wantSum)
			}
		})
	}
}

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

func TestPath(t *testing.T) {
	color.Enable = false
	cw, _ := os.Getwd()
	tests := []struct {
		name string
		want string
	}{
		{"", "/"},
		{"/notfounddir", "/notfounddir"},
		{cw, cw},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Path(tt.name); got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPort(t *testing.T) {
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
			if got := Port(tt.port); got != tt.want {
				t.Errorf("Port() = %v, want %v", got, tt.want)
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
