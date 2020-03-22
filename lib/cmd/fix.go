package cmd

import (
	"fmt"
	"os"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/text"
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

var fixDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Repair malformed database entries (SLOW)",
	Run: func(cmd *cobra.Command, args []string) {
		database.Fix()
		groups.Fix(simulate)
	},
}

var fixDemozooCmd = &cobra.Command{
	Use:   "demozoo",
	Short: "Repair imported Demozoo data conflicts",
	Run: func(cmd *cobra.Command, args []string) {
		demozoo.Fix()
	},
}

var fixImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Generate missing images",
	Run: func(cmd *cobra.Command, args []string) {
		err := images.Fix(simulate)
		logs.Check(err)
	},
}

var fixTextCmd = &cobra.Command{
	Use:   "text",
	Short: "Generate missing text previews",
	Run: func(cmd *cobra.Command, args []string) {
		err := text.Fix(simulate)
		logs.Check(err)
	},
}

func init() {
	rootCmd.AddCommand(fixCmd)
	fixCmd.AddCommand(fixDatabaseCmd)
	fixCmd.AddCommand(fixDemozooCmd)
	fixCmd.AddCommand(fixImagesCmd)
	fixCmd.AddCommand(fixTextCmd)
	fixCmd.PersistentFlags().BoolVarP(&simulate, "simulate", "s", true, "simulate the fixes and display the expected changes")
}
