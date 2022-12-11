package urlset

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	cfm    = ".cfm"
	index  = "index.cfm"
	static = "https://defacto2.net/"
)

// Set is a sitemap XML template.
type Set struct {
	XMLName xml.Name `xml:"urlset,omitempty"`
	XMLNS   string   `xml:"xmlns,attr,omitempty"`
	Urls    []Tag    `xml:"url,omitempty"`
}

// Tag composes the <url> tag in the sitemap.
type Tag struct {
	Location string `xml:"loc,omitempty"`
	// optional attributes
	LastModified string `xml:"lastmod,omitempty"`
	ChangeFreq   string `xml:"changefreq,omitempty"`
	Priority     string `xml:"priority,omitempty"`
}

func Paths() [28]string {
	return [...]string{
		"welcome",
		"file",
		"file/list/new",
		"organisation/list/group",
		"organisation/list/bbs",
		"organisation/list/ftp",
		"organisation/list/magazine",
		"person/list/artists",
		"person/list/coders",
		"person/list/musicians",
		"person/list/writers",
		"search/result",
		"defacto2",
		"defacto2/donate",
		"defacto2/history",
		"defacto2/subculture",
		"contact",
		"commercial",
		"code",
		"help",
		"help/creative-commons",
		"help/privacy",
		"help/browser-support",
		"help/keyboard",
		"help/viruses",
		"help/allowed-uploads",
		"help/categories",
		"link/list",
	}
}

func Html3Paths() [7]string {
	const s = "html3/"
	return [...]string{
		s,
		s + "art",
		s + "documents",
		s + "software",
		s + "groups",
		s + "platforms",
		s + "categories",
	}
}

func (set *Set) StaticURLs() (c, i int) {
	if set == nil || len(set.Urls) < len(Paths()) {
		return 0, 0
	}
	// sitemap priorities
	const top, veryHigh, high, standard = "1", "0.9", "0.8", "0.7"
	uri := func(path string) string {
		return static + path
	}
	c, i, view := 0, 0, viper.GetString("directory.views")
	for i, path := range Paths() {
		file := filepath.Join(view, path, index)
		if s, err := os.Stat(file); !os.IsNotExist(err) {
			set.Urls[i] = Tag{uri(path), Lastmod(s), "", veryHigh}
			c++
			continue
		}
		j := filepath.Join(view, path) + cfm
		if s, err := os.Stat(j); !os.IsNotExist(err) {
			set.Urls[i] = Tag{uri(path), Lastmod(s), "", high}
			c++
			continue
		}
		k := filepath.Join(view, strings.ReplaceAll(path, "-", "")+cfm)
		if s, err := os.Stat(k); !os.IsNotExist(err) {
			set.Urls[i] = Tag{uri(path), Lastmod(s), "", standard}
			c++
			continue
		}
		set.Urls[i] = Tag{uri(path), "", "", top}
	}
	return c, i
}

func Lastmod(s os.FileInfo) string {
	return s.ModTime().UTC().Format("2006-01-02")
}
