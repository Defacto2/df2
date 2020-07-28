package cmd

import (
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof"
	"github.com/spf13/cobra"
)

type proofFlags struct {
	all         bool   // scan for all proofs, not just new uploads
	hideMissing bool   // hide proofs that are missing their file download
	overwrite   bool   // overwrite all existing images
	id          string // auto-generated id or a uuid
}

var proo proofFlags

// proofCmd represents the proof command.
var proofCmd = &cobra.Command{
	Use:     "proof",
	Short:   "Handler for files tagged as #releaseproof",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		r := proof.Request{
			Overwrite:   proo.overwrite,
			AllProofs:   proo.all,
			HideMissing: proo.hideMissing,
		}
		switch {
		case proo.id != "":
			if err := r.Query(proo.id); err != nil {
				logs.Danger(err)
			}
		default:
			if err := r.Queries(); err != nil {
				logs.Danger(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(proofCmd)
	proofCmd.Flags().StringVarP(&proo.id, "id", "i", "", "id or uuid to handle only one proof")
	proofCmd.Flags().BoolVar(&proo.overwrite, "overwrite", false, "rescan archives and overwrite all existing images")
	proofCmd.Flags().BoolVar(&proo.all, "all", false, "scan for all proofs, not just new uploads")
	proofCmd.Flags().BoolVarP(&proo.hideMissing, "hide-missing", "m", false, "hide proofs that are missing their file download")
}
