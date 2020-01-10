package cmd

import (
	"fmt"
	"os"

	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

// fixCmd represents the fix command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Fixes database entries and records",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Usage()
			os.Exit(0)
		}
		_ = cmd.Usage()
		logs.Check(fmt.Errorf("invalid command %v please use one of the available fix commands", args[0]))
	},
}

var fixGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Repair malformed group names",
	Run: func(cmd *cobra.Command, args []string) {
		groups.Fix(simulate)
	},
}

func init() {
	rootCmd.AddCommand(fixCmd)
	fixCmd.AddCommand(fixGroupsCmd)
	fixCmd.PersistentFlags().BoolVarP(&simulate, "simulate", "s", true, "simulate the fixes and display the expected changes")
}
