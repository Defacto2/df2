//nolint:gochecknoglobals,gochecknoinits
package cmd

import (
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/proof"
	"github.com/spf13/cobra"
)

var proofs arg.Proof

// proofCmd represents the proof command.
var proofCmd = &cobra.Command{
	Use:   "proof",
	Short: "Manage records tagged as #releaseproof.",
	Long: `Group release proofs verify the use of retail-ready physical media
for scene releases. These proofs often come in archives containing
photos and text NFO files.`,
	Aliases: []string{"p"},
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		r := proof.Request{
			Overwrite:   proofs.Overwrite,
			All:         proofs.All,
			HideMissing: proofs.HideMissing,
		}
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		w := os.Stdout
		switch {
		case proofs.ID != "":
			if err := r.Query(db, w, confg, proofs.ID); err != nil {
				logr.Error(err)
			}
		default:
			if err := r.Queries(db, w, confg); err != nil {
				logr.Error(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(proofCmd)
	proofCmd.Flags().StringVarP(&proofs.ID, "id", "i", "",
		"ID or UUID to handle only one proof")
	proofCmd.Flags().BoolVar(&proofs.Overwrite, "overwrite", false,
		"rescan archives and overwrite all existing images")
	proofCmd.Flags().BoolVar(&proofs.All, "all", false,
		"scan for all proofs, not only new uploads")
	proofCmd.Flags().BoolVarP(&proofs.HideMissing, "hide-missing", "m", false,
		"hide proofs that are missing a file to download")
}
