package prods

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/Defacto2/df2/pkg/download"
)

// DownloadsAPIv1 are DownloadLinks for ProductionsAPIv1.
type DownloadsAPIv1 struct {
	LinkClass string `json:"link_class"`
	URL       string `json:"url"`
}

// parse corrects any known errors with a Downloads API link.
func (dl *DownloadsAPIv1) parse() bool {
	u, err := url.Parse(dl.URL) // validate url
	if err != nil {
		return false
	}
	u = Mutate(u)
	dl.URL = u.String()
	return true
}

// DownloadLink parses the Demozoo DownloadLinks to return the filename and link of the first suitable download.
func (p *ProductionsAPIv1) DownloadLink() (string, string) { //nolint:cyclop
	const (
		found       = 200
		internalErr = 500
	)
	link, name := "", ""
	total := len(p.DownloadLinks)
	for _, l := range p.DownloadLinks {
		var l DownloadsAPIv1 = l // apply type so we can use it with methods
		if ok := l.parse(); !ok {
			continue
		}
		// skip defacto2 links if others are available
		if u, err := url.Parse(l.URL); total > 1 && u.Hostname() == df2 {
			if flag.Lookup("test.v") != nil {
				log.Printf("url.Parse(%s) error = %q\n", l.URL, err)
			}
			continue
		}
		ping, err := download.PingHead(l.URL)
		if err != nil || ping.StatusCode != found {
			if flag.Lookup("test.v") != nil {
				if err != nil {
					log.Printf("download.Ping(%s) error = %q\n", l.URL, err)
				} else {
					log.Printf("download.Ping(%s) %v != %v\n", l.URL, ping.StatusCode, found)
				}
			}
			continue
		}
		defer ping.Body.Close()
		name = Filename(ping.Header)
		if name == "" {
			name, err = SaveName(l.URL)
			if err != nil {
				continue
			}
		}
		link = l.URL
		break
	}
	return name, link
}

func (p *ProductionsAPIv1) Download(w io.Writer, l DownloadsAPIv1) error {
	const found = 200
	if ok := l.parse(); !ok {
		fmt.Fprint(w, " not usable\n")
		return nil
	}
	ping, err := download.PingHead(l.URL)
	if err != nil {
		return fmt.Errorf("download off demozoo ping: %w", err)
	}
	defer ping.Body.Close()
	if ping.StatusCode != found {
		fmt.Fprintf(w, " %s", ping.Status) // print the HTTP status
		return nil
	}
	save, err := SaveName(l.URL)
	if err != nil {
		return fmt.Errorf("download off demozoo: %w", err)
	}
	temp, err := os.MkdirTemp("", "demozoo-download")
	if err != nil {
		return fmt.Errorf("download off demozoo temp dir: %w", err)
	}
	dest, err := filepath.Abs(filepath.Join(temp, save))
	if err != nil {
		return fmt.Errorf("download off demozoo abs filepath: %w", err)
	}
	_, err = download.GetSave(w, dest, l.URL)
	if err != nil {
		return fmt.Errorf("download off demozoo download: %w", err)
	}
	return nil
}

// Downloads parses the Demozoo DownloadLinks and saves the first suitable download.
func (p *ProductionsAPIv1) Downloads(w io.Writer) {
	for _, l := range p.DownloadLinks {
		if err := p.Download(w, l); err != nil {
			fmt.Fprintf(w, " %s", err)
		} else {
			break
		}
	}
}
