package prods

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/Defacto2/df2/pkg/download"
)

const df2 = "defacto2.net"

// DownloadsAPIv1 are DownloadLinks for ProductionsAPIv1.
type DownloadsAPIv1 struct {
	LinkClass string `json:"link_class"`
	URL       string `json:"url"`
}

// parse corrects any known errors with a Downloads API link.
func (dl *DownloadsAPIv1) parse() error {
	u, err := url.Parse(dl.URL) // validate url
	if err != nil {
		return err
	}
	m, err := Mutate(u)
	if err != nil {
		return err
	}
	dl.URL = m.String()
	return nil
}

// DownloadLink parses the Demozoo DownloadLinks to return the filename
// and link of the first suitable download.
func (p *ProductionsAPIv1) DownloadLink(w io.Writer) (string, string) { //nolint:cyclop
	if w == nil {
		w = io.Discard
	}
	const httpOk = 200
	link, name := "", ""
	total := len(p.DownloadLinks)
	for _, l := range p.DownloadLinks {
		var l DownloadsAPIv1 = l // apply the DownloadsAPIv1 type so we can use the value it with other methods
		if err := l.parse(); err != nil {
			fmt.Fprintf(w, "dl.Parse(%s) error = %q\n", l.URL, err)
			continue
		}
		// skip defacto2 links if others are available
		if u, err := url.Parse(l.URL); total > 1 && u.Hostname() == df2 {
			fmt.Fprintf(w, "url.Parse(%s) error = %q\n", l.URL, err)
			continue
		}
		ping, err := download.PingHead(l.URL)
		if err != nil {
			fmt.Fprintf(w, "download.Ping(%s) error = %q\n", l.URL, err)
			continue
		}
		if ping.StatusCode != httpOk {
			fmt.Fprintf(w, "download.Ping(%s) %v != %v\n", l.URL, ping.StatusCode, httpOk)
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
	if w == nil {
		w = io.Discard
	}
	const httpOk = 200
	if err := l.parse(); err != nil {
		fmt.Fprint(w, " not usable\n")
		return nil
	}
	ping, err := download.PingHead(l.URL)
	if err != nil {
		return fmt.Errorf("ping download from demozoo: %w", err)
	}
	defer ping.Body.Close()
	if ping.StatusCode != httpOk {
		fmt.Fprintf(w, " %s", ping.Status) // print the HTTP status
		return nil
	}
	save, err := SaveName(l.URL)
	if err != nil {
		return fmt.Errorf("save download from demozoo: %w", err)
	}
	tmp, err := os.MkdirTemp("", "demozoo-download")
	if err != nil {
		return fmt.Errorf("tmpdir download from demozoo: %w", err)
	}
	dest, err := filepath.Abs(filepath.Join(tmp, save))
	if err != nil {
		return fmt.Errorf("abs download from demozoo: %w", err)
	}
	_, err = download.GetSave(w, dest, l.URL)
	if err != nil {
		return fmt.Errorf("get download from demozoo: %w", err)
	}
	return nil
}

// Downloads parses the Demozoo DownloadLinks and saves the first suitable download.
func (p *ProductionsAPIv1) Downloads(w io.Writer) {
	if w == nil {
		w = io.Discard
	}
	for _, l := range p.DownloadLinks {
		if err := p.Download(w, l); err != nil {
			fmt.Fprintf(w, " %s", err)
			continue
		}
		break
	}
}
