// nolint:gochecknoglobals
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof"
)

type proofArg struct {
	all         bool   // scan for all proofs, not just new uploads
	hideMissing bool   // hide proofs that are missing their file download
	overwrite   bool   // overwrite all existing images
	id          string // auto-generated id or a uuid
}

var prf proofArg

// proofCmd represents the proof command.
var proofCmd = &cobra.Command{
	Use:     "proof",
	Short:   "Handler for files tagged as #releaseproof",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		r := proof.Request{
			Overwrite:   prf.overwrite,
			AllProofs:   prf.all,
			HideMissing: prf.hideMissing,
		}
		switch {
		case prf.id != "":
			if err := r.Query(prf.id); err != nil {
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
	proofCmd.Flags().StringVarP(&prf.id, "id", "i", "", "id or uuid to handle only one proof")
	proofCmd.Flags().BoolVar(&prf.overwrite, "overwrite", false, "rescan archives and overwrite all existing images")
	proofCmd.Flags().BoolVar(&prf.all, "all", false, "scan for all proofs, not just new uploads")
	proofCmd.Flags().BoolVarP(&prf.hideMissing, "hide-missing", "m", false, "hide proofs that are missing their file download")
}
