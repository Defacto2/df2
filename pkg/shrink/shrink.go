// Package shrink reduces the hard drive space used by the web application.
package shrink

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/shrink/internal/data"
	"github.com/gookit/color"
)

// Files approves and then archives incoming files.
func Files(db *sql.DB, w io.Writer, incomingFiles string) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	fmt.Fprint(w, color.Primary.Sprintf("Incoming files directory: %s\n", incomingFiles))
	if err := data.Incoming.Approve(db); err != nil {
		return err
	}
	if _, err := data.Incoming.Store(w, incomingFiles, "incoming-files", true); err != nil {
		return err
	}
	fmt.Fprintln(w, "Incoming storage is complete.")
	return nil
}

// Previews approves and archives incoming preview images.
func Previews(db *sql.DB, w io.Writer, incomingImg string) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	fmt.Fprint(w, color.Primary.Sprintf("Previews incoming directory: %s\n", incomingImg))
	if err := data.Preview.Approve(db); err != nil {
		return nil //nolint: nilerr
	}
	if _, err := data.Preview.Store(w, incomingImg, "incoming-preview", true); err != nil {
		return err
	}
	fmt.Fprintln(w, "Previews storage is complete.")
	return nil
}

func SQL(w io.Writer, directory string) error {
	if w == nil {
		w = io.Discard
	}
	return data.Init(w, directory)
}
