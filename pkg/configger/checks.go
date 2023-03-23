package configger

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c Config) Checks(log *zap.SugaredLogger) {
	// DownloadDir(c.DownloadDir, log)
	// LogDir(c.LogDir, log)
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
func DownloadDir(name string, log *zap.SugaredLogger) {
	if name == "" {
		log.Warn("The download directory path is empty, the server cannot send record downloads.")
		return
	}
	dir, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		log.Warnf("The download directory path does not exist, the server cannot send record downloads: %s", name)
		return
	}
	if !dir.IsDir() {
		log.Fatalf("The download directory path points to the file: %s", dir.Name())
	}
	files, err := os.ReadDir(name)
	if err != nil {
		log.Fatalf("The download directory path could not be read: %s.", err)
	}
	const minFiles = 10
	if len(files) < minFiles {
		log.Warnf("The download directory path contains only a few items, is the directory correct:  %s",
			dir.Name())
		return
	}
}

// LogDir runs checks against the named log directory.
// Problems will either log warnings or fatal errors.
func LogDir(name string, log *zap.SugaredLogger) {
	if name == "" {
		// recommended
		return
	}
	dir, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("The log directory path does not exist, the server cannot log to files: %s", name)
	}
	if !dir.IsDir() {
		log.Fatalf("The log directory path points to the file: %s", dir.Name())
	}
	empty := filepath.Join(name, ".defacto2_touch_test")
	f, err := os.Create(empty)
	if err != nil {
		log.Fatalf("Could not create a file in the log directory path: %s.", err)
	}
	defer f.Close()
	if err := os.Remove(empty); err != nil {
		log.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
		return
	}
}
