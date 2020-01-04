package cmd

import (
	"fmt"

	"github.com/Defacto2/df2/lib/people"
	"github.com/spf13/cobra"
)

// authorCmd represents the authors command
var authorCmd = &cobra.Command{
	Use:   "author",
	Short: "A HTML snippet generator to list authors",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("authors called")
		r := people.Request{}
		Print(r)
	},
}

func init() {
	rootCmd.AddCommand(authorCmd)
}
