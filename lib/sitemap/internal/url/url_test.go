package url_test

import (
	"os"
	"testing"
	"time"

	"github.com/Defacto2/df2/lib/sitemap/internal/url"
)

type mockedFileInfo struct {
	// Embed this so we only need to add methods used by testable functions
	os.FileInfo
	modtime time.Time
}

func (m mockedFileInfo) ModTime() time.Time { return m.modtime }

func TestLastmod(t *testing.T) {
	const year, month, day, hour, want = 1980, 1, 1, 12, "1980-01-01"
	nyd := time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	mfi := mockedFileInfo{
		modtime: nyd,
	}
	if got := url.Lastmod(mfi); got != want {
		t.Errorf("Lastmod() = %v, want %v", got, want)
	}
}
