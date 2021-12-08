// nolint:gochecknoglobals
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/people"
	"github.com/Defacto2/df2/lib/text"
	"github.com/Defacto2/df2/lib/zipcmmt"
	"github.com/Defacto2/df2/lib/zipcontent"
	"github.com/spf13/cobra"
)

// fixCmd represents the fix command.
var fixCmd = &cobra.Command{
	Use:     "fix",
	Short:   "Fixes database entries and records",
	Aliases: []string{"f"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Usage(); err != nil {
				logs.Fatal(err)
			}
			os.Exit(0)
		}
		if err := cmd.Usage(); err != nil {
			logs.Fatal(err)
		}
		logs.Danger(fmt.Errorf("fix cmd %q: %w", args[0], ErrCmd))
	},
}

var fixArchivesCmd = &cobra.Command{
	Use:     "archives",
	Short:   "Repair archives listing empty content",
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := zipcontent.Fix(true); err != nil {
			log.Fatal(fmt.Errorf("archives fix: %w", err))
		}
	},
}

var fixDatabaseCmd = &cobra.Command{
	Use:     "database",
	Short:   "Repair malformed database entries",
	Aliases: []string{"d", "db"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.Fix(); err != nil {
			log.Fatal(err)
		}
		if err := groups.Fix(simulate); err != nil {
			log.Fatal(err)
		}
		if err := people.Fix(simulate); err != nil {
			log.Fatal(err)
		}
	},
}

var fixDemozooCmd = &cobra.Command{
	Use:     "demozoo",
	Short:   "Repair imported Demozoo data conflicts",
	Aliases: []string{"dz"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := demozoo.Fix(); err != nil {
			log.Fatal(fmt.Errorf("demozoo fix: %w", err))
		}
	},
}

var fixImagesCmd = &cobra.Command{
	Use:     "images",
	Short:   "Generate missing images",
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := images.Fix(simulate); err != nil {
			log.Fatal(err)
		}
	},
}

var fixTextCmd = &cobra.Command{
	Use:     "text",
	Short:   "Generate missing text previews",
	Aliases: []string{"t", "txt"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := text.Fix(simulate); err != nil {
			log.Fatal(err)
		}
	},
}

var fixZipCmmt struct {
	ascii   bool
	unicode bool
	ow      bool
}

var fixZipCmmtCmd = &cobra.Command{
	Use:     "zip",
	Short:   "Extract missing comments from zip archives",
	Aliases: []string{"z"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := zipcmmt.Fix(fixZipCmmt.ascii, fixZipCmmt.unicode, fixZipCmmt.ow); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(fixCmd)
	fixCmd.AddCommand(fixArchivesCmd)
	fixCmd.AddCommand(fixDatabaseCmd)
	fixCmd.AddCommand(fixDemozooCmd)
	fixCmd.AddCommand(fixImagesCmd)
	fixCmd.AddCommand(fixTextCmd)
	fixCmd.AddCommand(fixZipCmmtCmd)
	fixCmd.PersistentFlags().BoolVarP(&simulate, "dry-run", "d", false,
		"simulate the fixes and display the expected changes")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&fixZipCmmt.ascii, "print", "p", false,
		"also print saved comments to the stdout")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&fixZipCmmt.unicode, "unicode", "u", false,
		"also convert saved comments into Unicode and print to the stdout")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&fixZipCmmt.ow, "overwrite", "o", false,
		"overwrite all existing saved comments")
}
