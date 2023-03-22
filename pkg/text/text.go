// Package text generates images from text files using the Ansilove/C program.
package text

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/text/internal/tf"
)

const (
	stmt = "SELECT id, uuid, filename, filesize, retrotxt_no_readme, retrotxt_readme, platform " +
		"FROM files WHERE platform=\"text\" OR platform=\"textamiga\" OR platform=\"ansi\" ORDER BY id DESC"
)

// Fix generates any missing assets from downloads that are text based.
func Fix(db *sql.DB, w io.Writer, cfg configger.Config) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	dir, err := directories.Init(cfg, false)
	if err != nil {
		return err
	}
	rows, err := db.Query(stmt)
	if err != nil {
		return fmt.Errorf("fix db query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("fix rows: %w", rows.Err())
	}
	defer rows.Close()
	i, c := 0, 0
	for rows.Next() {
		if i, c, err = fixRow(w, cfg, i, c, &dir, rows); err != nil {
			if !errors.Is(err, tf.ErrPNG) {
				return err
			}
		}
	}
	fmt.Fprintln(w, "scanned", c, "fixes from", i, "text file records")
	if c == 0 {
		fmt.Fprintln(w, "everything is okay, there is nothing to do")
	}
	return nil
}

func fixRow(w io.Writer, cfg configger.Config, i, c int, dir *directories.Dir, rows *sql.Rows) (int, int, error) {
	var t tf.TextFile
	i++
	if err1 := rows.Scan(&t.ID, &t.UUID, &t.Name, &t.Size, &t.NoReadme, &t.Readme, &t.Platform); err1 != nil {
		return i, c, fmt.Errorf("fix rows scan: %w", err1)
	}
	ok, err := t.Exist(dir)
	if err != nil {
		return i, c, fmt.Errorf("fix exist: %w", err)
	}
	// missing images + source is an archive
	if !ok && t.Archive() {
		c++
		return extract(w, cfg, t, i, c, dir)
	}
	// missing images + source is a textfile
	if !ok {
		c++
		if err := t.TextPng(w, cfg, c, dir.UUID); err != nil {
			return i, c, err
		}
	}
	// missing webp specific images that rely on PNG sources
	c, err = t.WebP(w, c, dir.Img000)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	return i, c, nil
}

func extract(w io.Writer, cfg configger.Config, t tf.TextFile, i, c int, dir *directories.Dir) (int, int, error) {
	err := t.Extract(w, dir)
	switch {
	case errors.Is(err, tf.ErrMeUnk):
		return i, c, nil
	case errors.Is(err, tf.ErrMeNo):
		return i, c, nil
	case err != nil:
		fmt.Fprintln(w, t.String(), err)
		return i, c, nil
	}
	if err := t.ExtractedImgs(w, cfg, dir.UUID); err != nil {
		fmt.Fprintln(w, t.String(), err)
	}
	return i, c, nil
}
