// nolint:gochecknoglobals
package cmd

import (
	"fmt"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:     "lookup (ids|uuids)",
	Short:   "Lookup the file URL of a record's ID or UUID.",
	Aliases: []string{"l"},
	Example: `  id is a a unique numeric identifier
  uuid is a unique 35-character hexadecimal string representation of a 128-bit integer
  uuid character groups are 8-4-4-16 (xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx)`,
	Hidden: false,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Usage(); err != nil {
				logs.Fatal(err)
			}
		}
		for _, a := range args {
			if err := database.CheckID(a); err != nil {
				fmt.Printf("%s: %s\n", ErrID, a)
				continue
			}
			id, err := database.GetID(a)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("https://defacto2.net/f/%v\n",
				database.ObfuscateParam(fmt.Sprint(id)))
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(lookupCmd)
}
