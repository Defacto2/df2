package shrink

import (
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/shrink/internal/sql"
	"github.com/gookit/color"
)

// Files approves and then archives incoming files.
func Files(w io.Writer, incomingFiles string) error {
	color.Primary.Printf("Incoming files directory: %s\n", incomingFiles)
	if err := sql.Incoming.Approve(w); err != nil {
		return err
	}
	if err := sql.Incoming.Store(w, incomingFiles, "incoming-files"); err != nil {
		return err
	}
	fmt.Fprintln(w, "Incoming storage is complete.")
	return nil
}

// Previews approves and archives incoming preview images.
func Previews(w io.Writer, incomingImg string) error {
	color.Primary.Printf("Previews incoming directory: %s\n", incomingImg)
	if err := sql.Preview.Approve(w); err != nil {
		return nil //nolint: nilerr
	}
	if err := sql.Preview.Store(w, incomingImg, "incoming-preview"); err != nil {
		return err
	}
	fmt.Fprintln(w, "Previews storage is complete.")
	return nil
}

func SQL(w io.Writer, directory string) error {
	return sql.Init(w, directory)
}
