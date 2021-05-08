package zipcmmt

import (
	"fmt"
	"os"
	"path/filepath"
)

func (z zipfile) checkDownload(path string) (ok bool) {
	file := filepath.Join(fmt.Sprint(path), z.UUID)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func (z zipfile) checkCmmtFile(path string) (ok bool) {
	if z.overwrite {
		return true
	}
	file := filepath.Join(fmt.Sprint(path), z.UUID+filename)
	if _, err := os.Stat(file); err == nil {
		return false
	}
	return true
}
