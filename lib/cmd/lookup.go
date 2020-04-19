package cmd

import (
	"errors"
	"fmt"

	"github.com/Defacto2/df2/lib/database"
	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:   "lookup (id|uuid)",
	Short: "Lookup the file URL of a database ID or UUID",
	Example: `  id is a a unique numeric identifier
  uuid is a unique 35-character hexadecimal string representation of a 128-bit integer
  uuid character groups are 8-4-4-16 (xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx)`,
	Hidden: false,
	Args: func(cmd *cobra.Command, args []string) error {
		const help = ""
		if len(args) != 1 {
			return errors.New("lookup: requires an id or uuid argument")
		}
		if err := database.CheckID(args[0]); err == nil {
			return nil
		}
		return fmt.Errorf("lookup: invalid id or uuid specified %q", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		id, err := database.LookupID(args[0])
		if err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("https://defacto2.net/f/%v\n", database.ObfuscateParam(fmt.Sprint(id)))
		}
	},
}

func init() {
	rootCmd.AddCommand(lookupCmd)
}
