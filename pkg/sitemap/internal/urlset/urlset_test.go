package urlset_test

import (
	"encoding/xml"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/sitemap/internal/urlset"
)

type mockedFileInfo struct {
	// Embed this so we only need to add methods used by testable functions
	os.FileInfo
	modtime time.Time
}

func (m mockedFileInfo) ModTime() time.Time { return m.modtime }

func TestLastmod(t *testing.T) {
	t.Parallel()
	const year, month, day, hour, want = 1980, 1, 1, 12, "1980-01-01"
	nyd := time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	mfi := mockedFileInfo{
		modtime: nyd,
	}
	if got := urlset.Lastmod(mfi); got != want {
		t.Errorf("Lastmod() = %v, want %v", got, want)
	}
}

func TestPaths(t *testing.T) {
	t.Parallel()
	got := urlset.Paths()
	const expected = 28
	if l := len(got); l != expected {
		t.Errorf("len(Paths()) = %v, want %v", l, expected)
	}
	want := "welcome"
	if g := got[0]; g != want {
		t.Errorf("Paths()[0] = %v, want %v", g, want)
	}
	want = "link/list"
	if g := got[expected-1]; g != want {
		t.Errorf("Paths()[28] = %v, want %v", g, want)
	}
}

func tags() []urlset.Tag {
	size := 29
	urls := make([]urlset.Tag, size)
	tag := urlset.Tag{Location: "/url-path-"}
	for i := 1; i < size; i++ {
		urls[i] = tag
		urls[i].Location += fmt.Sprint(i)
	}
	return urls
}

func TestSet_StaticURLs(t *testing.T) {
	t.Parallel()
	tag := urlset.Tag{Location: "/url-path-1"}
	type fields struct {
		XMLName xml.Name     `xml:"urlset,omitempty"`
		XMLNS   string       `xml:"xmlns,attr,omitempty"`
		URLs    []urlset.Tag `xml:"url,omitempty"`
	}
	tests := []struct {
		name   string
		fields fields
		wantC  int
		wantI  int
	}{
		{"empty", fields{}, 0, 0},
		{"too small", fields{
			XMLName: xml.Name{Space: " ", Local: "set.Urls"},
			XMLNS:   "pretend namespace",
			URLs:    []urlset.Tag{tag},
		}, 0, 0},
		{"okay", fields{
			XMLName: xml.Name{Space: " ", Local: "set.Urls"},
			XMLNS:   "pretend namespace",
			URLs:    tags(),
		}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &urlset.Set{
				XMLName: tt.fields.XMLName,
				XMLNS:   tt.fields.XMLNS,
				URLs:    tt.fields.URLs,
			}
			gotC, gotI := set.StaticURLs("")
			if gotC != tt.wantC {
				t.Errorf("Set.StaticURLs() gotC = %v, want %v", gotC, tt.wantC)
			}
			if gotI != tt.wantI {
				t.Errorf("Set.StaticURLs() gotI = %v, want %v", gotI, tt.wantI)
			}
		})
	}
}
