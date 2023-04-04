// Package cnter is an optional progress counter for the remote file downloads.
package cnter

import (
	"io"

	"github.com/Defacto2/df2/pkg/logger"
	"github.com/dustin/go-humanize"
)

// Writer totals the number of bytes written.
type Writer struct {
	Name    string // Filename.
	Total   uint64 // Expected filesize.
	Written uint64 // Bytes written.
	W       io.Writer
}

// Write to stdout and also return the current write progress.
func (wc *Writer) Write(p []byte) (int, error) {
	n := len(p)
	wc.Written += uint64(n)
	wc.Percent()
	return n, nil
}

// Percent prints the current download progress.
func (wc Writer) Percent() {
	if pct := Percent(wc.Written, wc.Total); pct > 0 {
		logger.PrintfCR(wc.W, "downloading %s (%d%%) from %s", humanize.Bytes(wc.Written), pct, wc.Name)
		return
	}
	logger.PrintfCR(wc.W, "downloading %s from %s", humanize.Bytes(wc.Written), wc.Name)
}

// Percent of count in total.
func Percent(count, total uint64) uint64 {
	if total == 0 {
		return 0
	}
	const max = 100
	return uint64(float64(count) / float64(total) * max)
}
