package logs_test

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/logs"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"
	"github.com/stretchr/testify/assert"
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
	defer os.Remove(fp)

	err = logs.Saver(nil, name, nil)
	assert.NotNil(t, err)
	err = logs.Saver(nil, name, ErrEmpty)
	assert.Nil(t, err)
	err = logs.Saver(nil, name, ErrATest)
	assert.Nil(t, err)
}

func TestFilepath(t *testing.T) {
	t.Run("file path", func(t *testing.T) {
		if got := logs.Filepath(nil, logs.Filename); got == "" {
			t.Errorf("Filepath() = %q, want a directory path", got)
		}
	})
}

func printer(test, text string) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	switch test {
	case p:
		fmt.Fprint(os.Stdout, text)
	case "printcr":
		logs.Printcr(os.Stdout, text)
	case "printf":
		fmt.Fprintf(os.Stdout, "%s", text)
	case "println":
		fmt.Fprintln(os.Stdout, text)
	case "printfcr":
		logs.Printcrf(os.Stdout, "%s", text)
	}
	w.Close()
	bytes, _ := io.ReadAll(r)
	os.Stdout = rescueStdout
	return strings.TrimSpace(string(bytes))
}

func TestPrints(t *testing.T) {
	type args struct {
		test string
		text string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
	}{
		{p, args{p, ""}, ""},
		{"print hello", args{p, hi}, hi},
		{"cr", args{"printcr", ""}, ""},
		{"cr hello", args{"printcr", hi}, hi},
		{"f", args{"printf", ""}, ""},
		{"f hello", args{"printf", hi}, hi},
		{"ln", args{"println", ""}, ""},
		{"ln hello", args{"println", hi}, hi},
		{"fcr", args{"printfcr", ""}, ""},
		{"fcr hello", args{"printfcr", hi}, hi},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOutput := printer(tt.args.test, tt.args.text); gotOutput != tt.wantOutput {
				t.Errorf("printer = %v, want %v", gotOutput, tt.wantOutput)
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
