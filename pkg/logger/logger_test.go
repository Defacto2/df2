package logger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestSPrintPath(t *testing.T) {
	s := logger.SPrintPath("")
	assert.Equal(t, "/", s)
	dir, err := os.Getwd()
	assert.Nil(t, err)

	s = logger.SPrintPath(dir)
	assert.Contains(t, s, filepath.Join("pkg", "logger"))
	errDir := filepath.Join(dir, "some-subdir-not-exist", "another-false-dir")

	s = logger.SPrintPath(errDir)
	assert.Contains(t, s, "another-false-dir")
}
