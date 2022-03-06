package file

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/hako/durafmt"
)

// Thumb metadata for a JSON document.
type Thumb struct {
	UUID    string `json:"uuid"`
	URLID   string `json:"urlid"`
	Title   string `json:"title"`
	timeAgo string
	title   string
	group   string
	year    int
}

// Files data for a JSON document.
type Files struct {
	Cols [3]string `json:"COLUMNS"`
	Data []data    `json:"DATA"`
}

type data [3]string

const (
	id = iota
	uuid
	recordtitle
	groupbrandfor
	groupbrandby
	filename
	dateissuedyear
	createdat
)

// Scan the thumbnail for usable JSON metadata.
func (f *Thumb) Scan(values []sql.RawBytes) {
	if id := string(values[id]); id != "" {
		f.URLID = database.ObfuscateParam(id)
	}
	f.UUID = strings.ToLower(string(values[uuid]))
	if t, err := time.Parse(time.RFC3339, string(values[createdat])); err != nil {
		f.timeAgo = "Sometime"
	} else {
		f.timeAgo = fmt.Sprint(durafmt.Parse(time.Since(t)).LimitFirstN(1))
	}
	if rt := string(values[recordtitle]); rt != "" {
		f.title = fmt.Sprintf("%s (%s)", values[recordtitle], values[filename])
	} else {
		f.title = string(values[filename])
	}
	if g := string(values[groupbrandfor]); g != "" {
		f.group = g
	} else if g := string(values[groupbrandby]); g != "" {
		f.group = g
	} else {
		f.group = "an unknown group"
	}
	if y := string(values[dateissuedyear]); y != "" {
		i, err := strconv.Atoi(y)
		if err == nil {
			f.year = i
		}
	}
	f.Title = fmt.Sprintf("%s ago, %s for %s", f.timeAgo, f.title, f.group)
	const min = 1980
	if f.year >= min {
		f.Title += fmt.Sprintf(" in %d", f.year)
	}
}
