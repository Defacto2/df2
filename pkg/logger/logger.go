// Package logger uses the zap logging library for application logs.
// The development mode prints all feedback to stdout.
package logger

import (
	"fmt"
	"io"
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
	cliEncoder := console()
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(cliEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Production logger prints all info and higher log levels to stdout.
func Production() *zap.Logger {
	cliEncoder := console()
	defaultLogLevel := zapcore.InfoLevel
	core := zapcore.NewTee(
		zapcore.NewCore(cliEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func console() zapcore.Encoder {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

// SprintPath returns the named file or directory path with all missing elements marked in red.
func SprintPath(name string) string {
	a := strings.Split(name, "/")
	var p, s string
	for i, e := range a {
		if e == "" {
			s = "/"
			continue
		}
		p = strings.Join(a[0:i+1], "/")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			s = filepath.Join(s, color.Danger.Sprint(e))
		} else {
			s = filepath.Join(s, e)
		}
	}
	return fmt.Sprint(s)
}

// Printcr otherwise erases the current line and writes to standard output.
func Printcr(w io.Writer, a ...any) {
	fmt.Fprintf(w, "\r%s\r", strings.Repeat(" ", int(terminal.Size())))
	fmt.Fprint(w, a...)
}

// Printcrf erases the current line and formats according to a format specifier.
func Printcrf(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, "\r%s\r%s",
		strings.Repeat(" ", int(terminal.Size())),
		fmt.Sprintf(format, a...))
}
