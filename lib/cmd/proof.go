package cmd

import (
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof"
	"github.com/spf13/cobra"
)

var (
	// proofAll scan for all proofs, not just new uploads
	proofAll bool
	// proofID is an auto-generated id or a uuid
	proofID string
	// proofOverwrite overwrite all existing images
	proofOverwrite bool
)

// proofCmd represents the proof command
var proofCmd = &cobra.Command{
	Use:   "proof",
	Short: "Batch handler files tagged as #release-proof",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch {
		case proofID != "":
			err = proof.Query(proofID, proofOverwrite, proofAll)
		default:
			err = proof.Queries(proofOverwrite, proofAll)
		}
		logs.Check(err)
	},
}

func init() {
	rootCmd.AddCommand(proofCmd)
	proofCmd.Flags().StringVarP(&proofID, "id", "i", "", "id or uuid to handle only one proof")
	proofCmd.Flags().BoolVar(&proofOverwrite, "overwrite", false, "rescan archives and overwrite all existing images")
	proofCmd.Flags().BoolVar(&proofAll, "all", false, "scan for all proofs, not just new uploads")
}
