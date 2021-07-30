package sitemap

import (
	"log"
	"os"
	"testing"
	"time"
)

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Create(); err != nil {
			log.Print(err)
		}
	}
}

type mockedFileInfo struct {
	// Embed this so we only need to add methods used by testable functions
	os.FileInfo
	modtime time.Time
}

func (m mockedFileInfo) ModTime() time.Time { return m.modtime }

func Test_lastmod(t *testing.T) {
	const year, month, day, hour, want = 1980, 1, 1, 12, "1980-01-01"
	nyd := time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	mfi := mockedFileInfo{
		modtime: nyd,
	}
	if got := lastmod(mfi); got != want {
		t.Errorf("lastmod() = %v, want %v", got, want)
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"create", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Create(); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
