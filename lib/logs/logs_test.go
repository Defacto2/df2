package logs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
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
	if quiet {
		Quiet = true
	}
	switch test {
	case "print":
		Print(text)
	}
	w.Close()
	bytes, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	return string(bytes)
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
		{"print hello", args{"print", "hello", true}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOutput := capture(tt.args.test, tt.args.text, tt.args.quiet); gotOutput != tt.wantOutput {
				t.Errorf("capture() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}
