package logs_test

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/logs"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"
)

const p, hi = "print", "hello"

var (
	ErrEmpty = errors.New("")
	ErrATest = errors.New("test error: this is a test")
)

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
		{"args", args{
			arg:  "abc",
			args: []string{"abc", "def"},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := logs.Arg(tt.args.arg, false, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Arg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_save(t *testing.T) {
	const name = "test.log"
	fp, err := gap.NewScope(gap.User, logs.GapUser).LogPath(name)
	if err != nil {
		log.Print(err)
	}
	tests := []struct {
		name   string
		err    error
		wantOk bool
	}{
		{"nil", nil, false},
		{"empty", ErrEmpty, true},
		{"ok", ErrATest, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := logs.Save(name, tt.err); gotOk != tt.wantOk {
				t.Errorf("Save() = %v, want %v", gotOk, tt.wantOk)
			} else if gotOk {
				// cleanup
				if err := os.Remove(fp); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestFilepath(t *testing.T) {
	t.Run("file path", func(t *testing.T) {
		if got := logs.Filepath(logs.Filename); got == "" {
			t.Errorf("Filepath() = %q, want a directory path", got)
		}
	})
}

func capture(test, text string, quiet bool) (output string) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	logs.Quiet(false)
	if quiet {
		logs.Quiet(true)
	}
	switch test {
	case p:
		logs.Print(text)
	case "printcr":
		logs.Printcr(text)
	case "printf":
		logs.Printf("%s", text)
	case "println":
		logs.Println(text)
	case "printfcr":
		logs.Printcrf("%s", text)
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
		{p, args{p, "", false}, ""},
		{"print hello", args{p, hi, false}, hi},
		{"print !hello", args{p, hi, true}, ""},
		{"cr", args{"printcr", "", false}, ""},
		{"cr hello", args{"printcr", hi, false}, hi},
		{"cr !hello", args{"printcr", hi, true}, ""},
		{"f", args{"printf", "", false}, ""},
		{"f hello", args{"printf", hi, false}, hi},
		{"f !hello", args{"printf", hi, true}, ""},
		{"ln", args{"println", "", false}, ""},
		{"ln hello", args{"println", hi, false}, hi},
		{"ln !hello", args{"println", hi, true}, ""},
		{"fcr", args{"printfcr", "", false}, ""},
		{"fcr hello", args{"printfcr", hi, false}, hi},
		{"fcr !hello", args{"printfcr", hi, true}, ""},
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
			if got := logs.Path(tt.name); got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}
