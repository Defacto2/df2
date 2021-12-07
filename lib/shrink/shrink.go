package shrink

import (
	"fmt"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/shrink/internal/sql"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

func Files() {
	s := viper.GetString("directory.incoming.files")
	color.Primary.Printf("Incoming files directory: %s\n", s)
	if err := sql.Approve("incoming"); err != nil {
		logs.Danger(err)
		return
	}
	if err := sql.Store(s, "Incoming", "incoming-files"); err != nil {
		logs.Danger(err)
		return
	}
	fmt.Println("Incoming storage is complete.")
}

func Previews() {
	s := viper.GetString("directory.incoming.previews")
	color.Primary.Printf("Previews incoming directory: %s\n", s)
	if err := sql.Approve("previews"); err != nil {
		return
	}
	if err := sql.Store(s, "Previews", "incoming-preview"); err != nil {
		logs.Danger(err)
		return
	}
	fmt.Println("Previews storage is complete.")
}

func SQL() {
	if err := sql.Init(); err != nil {
		logs.Danger(err)
	}
}
