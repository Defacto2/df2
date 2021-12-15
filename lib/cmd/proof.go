// nolint:gochecknoglobals
package cmd

import (
	"github.com/Defacto2/df2/lib/cmd/internal/arg"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof"
	"github.com/spf13/cobra"
)

var proofs arg.Proof

// proofCmd represents the proof command.
var proofCmd = &cobra.Command{
	Use:     "proof",
	Short:   "Handler for files tagged as #releaseproof",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		r := proof.Request{
			Overwrite:   proofs.Overwrite,
			AllProofs:   proofs.All,
			HideMissing: proofs.HideMissing,
		}
		switch {
		case proofs.Id != "":
			if err := r.Query(proofs.Id); err != nil {
				logs.Danger(err)
			}
		default:
			if err := r.Queries(); err != nil {
				logs.Danger(err)
			}
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(proofCmd)
	proofCmd.Flags().StringVarP(&proofs.Id, "id", "i", "",
		"id or uuid to handle only one proof")
	proofCmd.Flags().BoolVar(&proofs.Overwrite, "overwrite", false,
		"rescan archives and overwrite all existing images")
	proofCmd.Flags().BoolVar(&proofs.All, "all", false,
		"scan for all proofs, not just new uploads")
	proofCmd.Flags().BoolVarP(&proofs.HideMissing, "hide-missing", "m", false,
		"hide proofs that are missing their file download")
}
