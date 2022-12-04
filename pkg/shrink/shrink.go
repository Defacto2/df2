package shrink

import (
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/shrink/internal/sql"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

// Files approves and then archives incoming files.
func Files() error {
	dir := viper.GetString("directory.incoming.files")
	color.Primary.Printf("Incoming files directory: %s\n", dir)
	if err := sql.Incoming.Approve(); err != nil {
		return err
	}
	if err := sql.Incoming.Store(dir, "incoming-files"); err != nil {
		return err
	}
	logs.Println("Incoming storage is complete.")
	return nil
}

// Previews approves and archives incoming preview images.
func Previews() error {
	dir := viper.GetString("directory.incoming.previews")
	color.Primary.Printf("Previews incoming directory: %s\n", dir)
	if err := sql.Preview.Approve(); err != nil {
		return nil //nolint: nilerr
	}
	if err := sql.Preview.Store(dir, "incoming-preview"); err != nil {
		return err
	}
	logs.Println("Previews storage is complete.")
	return nil
}

func SQL() error {
	return sql.Init()
}
