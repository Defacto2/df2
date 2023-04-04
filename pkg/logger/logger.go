// Package logger uses the zap logging library for application logs. The
// development mode prints all feedback to stdout.
package logger

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/logger/internal/terminal"
	"github.com/gookit/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Development logger prints all log levels to stdout.
func Development() *zap.Logger {
	enc := console()
	level := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), level),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Production logger prints all info and higher log levels to stdout.
func Production() *zap.Logger {
	enc := console()
	level := zapcore.InfoLevel
	core := zapcore.NewTee(
		zapcore.NewCore(enc, zapcore.AddSync(os.Stderr), level),
	)
	return zap.New(core, zap.AddCaller())
}

func console() zapcore.Encoder { //nolint:ireturn
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeCaller = nil
	config.EncodeTime = nil
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

// SPrintPath returns the named file or directory path with all missing elements marked in red.
func SPrintPath(name string) string {
	const sep = string(filepath.Separator)
	paths := strings.Split(name, sep)
	s := ""
	for i, e := range paths {
		if e == "" {
			s = string(filepath.Separator)
			continue
		}
		p := strings.Join(paths[0:i+1], string(filepath.Separator))
		if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
			s = filepath.Join(s, color.Danger.Sprint(e))
		} else {
			s = filepath.Join(s, e)
		}
	}
	return fmt.Sprint(s)
}

// PrintCR otherwise erases the current line and writes to standard output.
func PrintCR(w io.Writer, a ...any) {
	if w == nil {
		w = io.Discard
	}
	fmt.Fprintf(w, "\r%s\r", strings.Repeat(" ", int(terminal.Columns())))
	fmt.Fprint(w, a...)
}

// PrintfCR erases the current line and formats according to a format specifier.
func PrintfCR(w io.Writer, format string, a ...any) {
	if w == nil {
		w = io.Discard
	}
	fmt.Fprintf(w, "\r%s\r%s",
		strings.Repeat(" ", int(terminal.Columns())),
		fmt.Sprintf(format, a...))
}
