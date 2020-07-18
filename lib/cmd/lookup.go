package cmd

import (
	"errors"
	"fmt"

	"github.com/Defacto2/df2/lib/database"
	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:     "lookup (ids|uuids)",
	Short:   "Lookup the file URLs of database ID or UUID",
	Aliases: []string{"l"},
	Example: `  id is a a unique numeric identifier
  uuid is a unique 35-character hexadecimal string representation of a 128-bit integer
  uuid character groups are 8-4-4-16 (xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx)`,
	Hidden: false,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("lookup: requires an id or uuid argument")
		}
		for _, a := range args {
			if err := database.CheckID(a); err != nil {
				return fmt.Errorf("lookup: invalid id or uuid specified %q", a)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, a := range args {
			id, err := database.LookupID(a)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("https://defacto2.net/f/%v\n", database.ObfuscateParam(fmt.Sprint(id)))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(lookupCmd)
}
