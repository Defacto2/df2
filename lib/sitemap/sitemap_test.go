package sitemap

import (
	"os"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Create()
		})
	}
}

type mockedFileInfo struct {
	// Embed this so we only need to add methods used by testable functions
	os.FileInfo
	modtime time.Time
}

func (m mockedFileInfo) ModTime() time.Time { return m.modtime }

func Test_lastmod(t *testing.T) {
	var nyd = time.Date(1980, 1, 1, 12, 00, 00, 0, time.UTC)
	var mfi = mockedFileInfo{
		modtime: nyd,
	}
	var want = "1980-01-01"
	if got := lastmod(mfi); got != want {
		t.Errorf("lastmod() = %v, want %v", got, want)
	}
}
