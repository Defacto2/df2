package cmd

import (
	"log"

	"github.com/Defacto2/df2/pkg/dizzer"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:     "import (path to .rar)",
	Short:   "Import a .rar archive collection containing NFO and textfile.",
	Aliases: []string{"i"},
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Usage(); err != nil {
				logr.Fatal(err)
			}
			return
		}
		if err := dizzer.Run(args[0]); err != nil {
			log.Fatal(err)
		}

		// id
		// uuid
		// group_brand_for
		// record_title
		// date_issued_year / date_issued_month / .. day
		// filename
		// filesize
		// file_magic_type
		// file_integrity_strong
		// file_integrity_weak
		// file_last_modified
		// platform
		// section
		// comment
		// createdat
		// updatedat
		// deletedat
		// updatedby

		// head, ok := f.Header.(zip.FileHeader)
		// if ok {
		// 	fmt.Println("Filename:", head.Name)
		// 	fmt.Printf("%+v\n", head)
		// }
		// return nil

	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
