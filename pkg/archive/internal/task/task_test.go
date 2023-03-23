package task_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/task"
	"github.com/stretchr/testify/assert"
)

func testDir() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "testdata")
}

func TestRun(t *testing.T) {
	empty := task.Task{}
	th, tx, err := task.Run("")
	assert.NotNil(t, err)
	assert.Equal(t, empty, th)
	assert.Equal(t, empty, tx)
	th, tx, err = task.Run(testDir())
	assert.Nil(t, err)
	assert.Equal(t, empty, th)
	assert.NotEqual(t, empty, tx)
}
