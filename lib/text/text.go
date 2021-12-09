// Package text generates images from text files using the Ansilove/C program.
package text

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/text/internal/tf"
)

const (
	fixStmt = "SELECT id, uuid, filename, filesize, retrotxt_no_readme, retrotxt_readme, platform " +
		"FROM files WHERE platform=\"text\" OR platform=\"textamiga\" OR platform=\"ansi\" ORDER BY id DESC"
)

// Fix generates any missing assets from downloads that are text based.
func Fix(simulate bool) error {
	dir, db := directories.Init(false), database.Connect()
	defer db.Close()
	rows, err := db.Query(fixStmt)
	if err != nil {
		return fmt.Errorf("fix db query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("fix rows: %w", rows.Err())
	}
	defer rows.Close()
	i, c := 0, 0
	for rows.Next() {
		if i, c, err = fixRow(i, c, &dir, rows); err != nil {
			return err
		}
	}
	fmt.Println("scanned", c, "fixes from", i, "text file records")
	if simulate && c > 0 {
		logs.Simulate()
	} else if c == 0 {
		logs.Println("everything is okay, there is nothing to do")
	}
	return nil
}

func fixRow(i, c int, dir *directories.Dir, rows *sql.Rows) (scanned, records int, err error) {
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
		err1 := t.Extract(dir)
		switch {
		case errors.Is(err1, tf.ErrMeUnk):
			return i, c, nil
		case errors.Is(err1, tf.ErrMeNo):
			return i, c, nil
		case err1 != nil:
			fmt.Println(t.String(), err1)
			return i, c, nil
		}
		if err1 := t.ExtractedImgs(dir.UUID); err1 != nil {
			fmt.Println(t.String(), err1)
		}
		return i, c, nil
	}
	// missing images + source is a textfile
	if !ok {
		c++
		if err := t.TextPng(c, dir.UUID); err != nil {
			return i, c, err
		}
	}
	// missing webp specific images that rely on PNG sources
	c, err = t.WebP(c, dir.Img000)
	if err != nil {
		logs.Println(err)
	}
	return i, c, nil
}
