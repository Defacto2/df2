package conf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var (
	ErrPointer = errors.New("pointer value cannot be nil")
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c Config) Checks(l *zap.SugaredLogger) error {
	if l == nil {
		return fmt.Errorf("l %w", ErrPointer)
	}
	// DownloadDir(c.DownloadDir, log)
	// LogDir(c.LogDir, log)
	return nil
}

func BinPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("bin path error: %w", err)
	}
	return exe, nil
}

// DownloadDir runs checks against the named directory containing the UUID record downloads.
// Problems will either log warnings or fatal errors.
func DownloadDir(name string, l *zap.SugaredLogger) error {
	if l == nil {
		return fmt.Errorf("l %w", ErrPointer)
	}
	if name == "" {
		l.Warn("The download directory path is empty, the server cannot send record downloads.")
		return nil
	}
	dir, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		l.Warnf("The download directory path does not exist, the server cannot send record downloads: %s", name)
		return nil
	}
	if !dir.IsDir() {
		l.Fatalf("The download directory path points to the file: %s", dir.Name())
	}
	files, err := os.ReadDir(name)
	if err != nil {
		l.Fatalf("The download directory path could not be read: %s.", err)
	}
	const minFiles = 10
	if len(files) < minFiles {
		l.Warnf("The download directory path contains only a few items, is the directory correct:  %s",
			dir.Name())
		return nil
	}
	return nil
}

// LogDir runs checks against the named log directory.
// Problems will either log warnings or fatal errors.
func LogDir(name string, l *zap.SugaredLogger) error {
	if l == nil {
		return fmt.Errorf("l %w", ErrPointer)
	}
	if name == "" {
		// recommended
		return nil
	}
	dir, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		l.Fatalf("The log directory path does not exist, the server cannot log to files: %s", name)
	}
	if !dir.IsDir() {
		l.Fatalf("The log directory path points to the file: %s", dir.Name())
	}
	empty := filepath.Join(name, ".defacto2_touch_test")
	f, err := os.Create(empty)
	if err != nil {
		l.Fatalf("Could not create a file in the log directory path: %s.", err)
	}
	defer f.Close()
	if err := os.Remove(empty); err != nil {
		l.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
		return nil
	}
	return nil
}
