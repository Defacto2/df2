package cmd

import (
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof"
	"github.com/spf13/cobra"
)

type proofFlags struct {
	all         bool   // scan for all proofs, not just new uploads
	id          string // auto-generated id or a uuid
	hideMissing bool   // hide proofs that are missing their file download
	overwrite   bool   // overwrite all existing images
}

var prff proofFlags

// proofCmd represents the proof command
var proofCmd = &cobra.Command{
	Use:   "proof",
	Short: "Batch handler files tagged as #release-proof",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		r := proof.Request{
			Overwrite: prff.overwrite,
			All:       prff.all,
			HideMiss:  prff.hideMissing}
		switch {
		case prff.id != "":
			err = r.Query(prff.id)
		default:
			err = r.Queries()
		}
		logs.Check(err)
	},
}

func init() {
	rootCmd.AddCommand(proofCmd)
	proofCmd.Flags().StringVarP(&prff.id, "id", "i", "", "id or uuid to handle only one proof")
	proofCmd.Flags().BoolVar(&prff.overwrite, "overwrite", false, "rescan archives and overwrite all existing images")
	proofCmd.Flags().BoolVar(&prff.all, "all", false, "scan for all proofs, not just new uploads")
	proofCmd.Flags().BoolVarP(&prff.hideMissing, "hide-missing", "m", false, "hide proofs that are missing their file download")
}
