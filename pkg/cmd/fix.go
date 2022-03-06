// nolint:gochecknoglobals
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Defacto2/df2/pkg/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/people"
	"github.com/Defacto2/df2/pkg/text"
	"github.com/Defacto2/df2/pkg/zipcmmt"
	"github.com/Defacto2/df2/pkg/zipcontent"
	"github.com/spf13/cobra"
)

var zcf arg.ZipCmmt

// fixCmd represents the fix command.
var fixCmd = &cobra.Command{
	Use:     "fix",
	Short:   "Fixes database entries and records.",
	Long:    "Repair broken or invalid formatting for the database records and entries.",
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
	Use:   "archives",
	Short: "Repair archives listing empty content.",
	Long: `Records with downloads that are packaged into archives need to have
their file content added to the database. This command finds and repair
records that do not have this expected context.`,
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := zipcontent.Fix(true); err != nil {
			log.Print(fmt.Errorf("archives fix: %w", err))
		}
	},
}

var fixDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Repair malformed database entries.",
	Long: `Repair malformed records and entries in the database.
This includes the formatting and trimming of groups, people, platforms and sections.`,
	Aliases: []string{"d", "db"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.Fix(); err != nil {
			log.Print(err)
		}
		if err := groups.Fix(gf.Simulate); err != nil {
			log.Print(err)
		}
		if err := people.Fix(gf.Simulate); err != nil {
			log.Print(err)
		}
	},
}

var fixDemozooCmd = &cobra.Command{
	Use:     "demozoo",
	Short:   "Repair imported Demozoo data conflicts.",
	Aliases: []string{"dz"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := demozoo.Fix(); err != nil {
			log.Print(fmt.Errorf("demozoo fix: %w", err))
		}
	},
}

var fixImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Generate missing images.",
	Long: `Create missing previews, thumbnails and optimised formats for records
that are raster images.`,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := images.Fix(gf.Simulate); err != nil {
			log.Print(err)
		}
	},
}

var fixTextCmd = &cobra.Command{
	Use:   "text",
	Short: "Generate missing text previews.",
	Long: `Create missing previews, thumbnails and optimised formats for records
that are plain text files.`,
	Aliases: []string{"t", "txt"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := text.Fix(gf.Simulate); err != nil {
			log.Print(err)
		}
	},
}

var fixZipCmmtCmd = &cobra.Command{
	Use:   "zip",
	Short: "Extract missing comments from zip archives.",
	Long: `Extract and save missing comments from zip archives.

"A comment is optional text information that is embedded in a Zip file."`,
	Aliases: []string{"z"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := zipcmmt.Fix(zcf.ASCII, zcf.Unicode, zcf.OW, true); err != nil {
			log.Print(err)
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
	fixCmd.PersistentFlags().BoolVarP(&gf.Simulate, "dry-run", "d", false,
		"simulate the fixes and display the expected changes")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&zcf.ASCII, "print", "p", false,
		"also print saved comments to the stdout")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&zcf.Unicode, "unicode", "u", false,
		"also convert saved comments into Unicode and print to the stdout")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&zcf.OW, "overwrite", "o", false,
		"overwrite all existing saved comments")
}
