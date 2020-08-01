package logs

import (
	"errors"
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

func TestArg(t *testing.T) {
	type args struct {
		arg  string
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Arg(tt.args.arg, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Arg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_save(t *testing.T) {
	eerr, terr := errors.New(""), errors.New("test error: this is a test")
	tests := []struct {
		name   string
		err    error
		wantOk bool
	}{
		{"empty", nil, false},
		{"empty", eerr, false},
		{"ok", terr, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := save(tt.err); gotOk != tt.wantOk {
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
		Printcrf("%s", text)
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
