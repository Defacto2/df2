//nolint:gochecknoglobals
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/images"
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
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
		if err := cmd.Usage(); err != nil {
			log.Fatal(err)
		}
		log.Errorf("fix cmd %q: %w", args[0], ErrCmd)
	},
}

var fixArchivesCmd = &cobra.Command{
	Use:   "archives",
	Short: "Repair archives listing empty content.",
	Long: `Records with downloads that are packaged into archives need to have
their file content added to the database. This command finds and repair
records that do not have this expected context.`,
	Aliases: []string{"a"},
	GroupID: "groupU",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		if err := zipcontent.Fix(db, os.Stdout, log, cfg, true); err != nil {
			log.Info(fmt.Errorf("archives fix: %w", err))
		}
	},
}

var fixDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Repair malformed database entries.",
	Long: `Repair malformed records and entries in the database.
This includes the formatting and trimming of groups, people, platforms and sections.`,
	Aliases: []string{"d", "db"},
	GroupID: "groupU",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		w := os.Stdout
		if err := database.Fix(db, w, log); err != nil {
			log.Info(err)
		}
		if err := groups.Fix(db, w); err != nil {
			log.Info(err)
		}
		if err := people.Fix(db, w); err != nil {
			log.Info(err)
		}
	},
}

var fixDemozooCmd = &cobra.Command{
	Use:     "demozoo",
	Short:   "Repair imported Demozoo data conflicts.",
	Aliases: []string{"dz"},
	GroupID: "groupU",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		if err := demozoo.Fix(db, os.Stdout); err != nil {
			log.Info(fmt.Errorf("demozoo fix: %w", err))
		}
	},
}

var fixImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Generate missing images.",
	Long: `Create missing previews, thumbnails and optimised formats for records
that are raster images.`,
	Aliases: []string{"i"},
	GroupID: "groupG",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		if err := images.Fix(db, os.Stdout, log); err != nil {
			log.Info(err)
		}
	},
}

var fixRenGroup = &cobra.Command{
	Use:     "rename group replacement",
	Short:   "Rename all instances of a group.",
	Aliases: []string{"ren", "r"},
	GroupID: "groupR",
	Example: `  df2 fix rename "The Group" "New Group Name"`,
	Run: func(cmd *cobra.Command, args []string) {
		// in the future this command could be adapted to use a --person flag
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		err = run.Rename(db, os.Stdout, args...)
		if errors.Is(err, run.ErrToFew) {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
		if err != nil {
			log.Info(err)
		}
	},
}

var fixTextCmd = &cobra.Command{
	Use:   "text",
	Short: "Generate missing text previews.",
	Long: `Create missing previews, thumbnails and optimised formats for records
that are plain text files.`,
	Aliases: []string{"t", "txt"},
	GroupID: "groupG",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		if err := text.Fix(db, os.Stdout, cfg); err != nil {
			log.Info(err)
		}
	},
}

var fixZipCmmtCmd = &cobra.Command{
	Use:   "zip",
	Short: "Extract missing comments from zip archives.",
	Long: `Extract and save missing comments from zip archives.

"A comment is optional text information that is embedded in a Zip file."`,
	Aliases: []string{"z"},
	GroupID: "groupG",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		if err := zipcmmt.Fix(db, os.Stdout, cfg, zcf.ASCII, zcf.Unicode, zcf.OW, true); err != nil {
			log.Info(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(fixCmd)
	fixCmd.AddGroup(&cobra.Group{ID: "groupU", Title: "Repair:"})
	fixCmd.AddGroup(&cobra.Group{ID: "groupG", Title: "Create:"})
	fixCmd.AddGroup(&cobra.Group{ID: "groupR", Title: "Update:"})
	fixCmd.AddCommand(fixArchivesCmd)
	fixCmd.AddCommand(fixDatabaseCmd)
	fixCmd.AddCommand(fixDemozooCmd)
	fixCmd.AddCommand(fixImagesCmd)
	fixCmd.AddCommand(fixRenGroup)
	fixCmd.AddCommand(fixTextCmd)
	fixCmd.AddCommand(fixZipCmmtCmd)
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&zcf.ASCII, "print", "p", false,
		"also print saved comments to the stdout")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&zcf.Unicode, "unicode", "u", false,
		"also convert saved comments into Unicode and print to the stdout")
	fixZipCmmtCmd.PersistentFlags().BoolVarP(&zcf.OW, "overwrite", "o", false,
		"overwrite all existing saved comments")
}
