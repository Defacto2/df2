package assets

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gookit/color"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

type item struct {
	name  string // os.FileInfo.Name()
	path  string // filepath
	flag  string // check tick, mark or blank
	human bool   // humanise sizes
	cnt   string // loop count
	fm    string // file mode
	fs    string // file size
	mt    string // file modified time
}

// parse is used by scanPath to remove matched orphans.
func parse(s *scan, ignore files, list *[]os.FileInfo) (results, error) {
	const padding = 2
	stat := results{count: 0, fails: 0, bytes: 0}
	for _, file := range *list {
		if file.IsDir() {
			continue // ignore directories
		}
		if _, ign := ignore[file.Name()]; ign {
			continue // ignore files
		}
		i := item{human: s.human, name: file.Name()}
		uuid := strings.TrimSuffix(i.name, filepath.Ext(i.name))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
		// search the map `m` for `UUID`, the result is saved as a boolean to `exists`
		_, exists := s.m[uuid]
		if !exists {
			stat.totals(file)
			if s.delete {
				i.path = path.Join(s.path, file.Name())
				i.erase(stat)
			}
			i.count(stat.count)
			i.mod(file)
			i.size(file)
			i.bits(file)
			if !logs.Quiet {
				fmt.Fprintf(w, "%v\t%v%v\t%v\t%v\t%v\n", i.cnt, i.flag, i.name, i.fs, i.fm, i.mt)
			}
		}
		if err := w.Flush(); err != nil {
			return stat, fmt.Errorf("parse tabwriter flush: %w", err)
		}
	}
	return stat, nil
}

func (i *item) bits(f os.FileInfo) {
	i.fm = color.Note.Sprint(f.Mode())
}

func (i *item) count(c int) {
	i.cnt = color.Secondary.Sprint(strconv.Itoa(c) + ".")
}

func (i *item) erase(r results) {
	i.flag = str.Y()
	if err := os.Remove(i.path); err != nil {
		i.flag = str.X()
		r.fails++
	}
}

func (i *item) mod(f os.FileInfo) {
	s := fmt.Sprint(f.ModTime())
	if i.human {
		// show date and time
		s = f.ModTime().Format("02 Jan 15:04")
		if time.Now().Year() != f.ModTime().Year() {
			// otherwise, hide the time and show the year
			s = f.ModTime().Format("02 Jan 2006")
		}
	}
	i.mt = color.Primary.Sprint(s)
}

func (i *item) size(f os.FileInfo) {
	s := fmt.Sprint(f.Size())
	if i.human {
		s = humanize.Bytes(uint64(f.Size()))
	}
	i.fs = color.Comment.Sprint(s)
}

func (sum *results) totals(f os.FileInfo) {
	sum.count++
	sum.bytes += f.Size()
}
