// Package text generates preview images and thumbnails from text files using
// the system installed Ansilove/C program.
package text

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/text/internal/img"
	"github.com/Defacto2/df2/pkg/text/internal/tf"
	"go.uber.org/zap"
)

const (
	stmt = "SELECT id, uuid, filename, filesize, retrotxt_no_readme, retrotxt_readme, platform " +
		"FROM files WHERE platform=\"text\" OR platform=\"textamiga\" OR platform=\"ansi\" ORDER BY id DESC"
)

// Fix generates any missing assets from downloads that are text based.
func Fix(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg conf.Config) error {
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
		if _, i, c, err = fixRow(w, cfg, i, c, &dir, rows); err != nil {
			if errors.Is(tf.ErrReadmeOff, err) {
				// website admin has disabled the display of a readme
				continue
			}
			if errors.Is(err, img.ErrType) {
				// invalid mimetype
				continue
			}
			if !errors.Is(err, tf.ErrPNG) {
				fmt.Fprintf(w, "%s %s", str.X(), err)
				continue
			}
		}
	}
	fmt.Fprintln(w)
	str.Total(w, i, fmt.Sprintf("attempted to fix %d text files", i))
	return nil
}

func fixRow(
	w io.Writer, cfg conf.Config, i, c int, dir *directories.Dir, rows *sql.Rows,
) (tf.TextFile, int, int, error) {
	t := tf.TextFile{}
	i++
	if err1 := rows.Scan(&t.ID, &t.UUID, &t.Name, &t.Size, &t.NoReadme, &t.Readme, &t.Platform); err1 != nil {
		return t, i, c, fmt.Errorf("fix rows scan: %w", err1)
	}
	ok, err := t.Exist(dir)
	if err != nil {
		return t, i, c, fmt.Errorf("fix exist: %w", err)
	}
	if !ok {
		fmt.Fprintf(w, "\n%s%d. %s", str.PrePad, i, t.String())
	}
	// missing images + source is an archive
	if !ok && t.Archive() {
		c++
		if err := extract(w, cfg, t, dir); err != nil {
			return t, i, c, err
		}
	}
	// missing images + source is a textfile
	if !ok {
		c++
		if err := t.TextPNG(w, cfg, dir.UUID); err != nil {
			return t, i, c, err
		}
	}
	// missing webp specific images that rely on PNG sources
	c, err = t.WebP(w, c, dir.Img000)
	if err != nil {
		return t, i, c, err
	}
	return t, i, c, nil
}

func extract(w io.Writer, cfg conf.Config, t tf.TextFile, dir *directories.Dir) error {
	if dir == nil {
		return fmt.Errorf("dir %w", tf.ErrPointer)
	}
	if w == nil {
		w = io.Discard
	}
	if err := t.Extract(w, dir); err != nil {
		return err
	}
	if err := t.ExtractedImgs(w, cfg, dir.UUID); err != nil {
		fmt.Fprintln(w, t.String(), err)
	}
	return nil
}
