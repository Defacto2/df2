package cmd

import (
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

// todo: database fix nullify all empty "" deletedat records

var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Test by pinging select URLs of the website.",
	Aliases: []string{"t"},
	GroupID: "group3",
	Run: func(cmd *cobra.Command, args []string) {

		const pingCount = 10
		total, ids, err := sitemap.RandIDs(pingCount)
		if err != nil {
			logs.Fatal(err)
		}
		urls := ids.URLs()
		color.Primary.Printf("Requesting the <title> of %d random files from %d public records\n", pingCount, total)
		sitemap.LinkSuccess.Range(urls[:])

		const hideCount = 2
		total, ids, err = sitemap.RandDeleted(hideCount)
		if err != nil {
			logs.Fatal(err)
		}
		urls = ids.URLs()
		color.Primary.Printf("\nRequesting the <title> of %d random files from %d disabled records\n", hideCount, total)
		sitemap.LinkNotFound.Range(urls[:])

		invalid := []int{-99999999, -1, 0, 99999999}
		urls = sitemap.IDs(invalid).URLs()
		urls = append(urls, "https://defacto2.net/f/")
		urls = append(urls, "https://defacto2.net/f/womble-bomble")
		urls = append(urls, "https://defacto2.net/f/<script>")
		urls = append(urls, "https://defacto2.net/f/1+#51")
		color.Primary.Printf("\nRequesting the <title> of %d invalid file URLs\n", len(urls))
		sitemap.NotFound.Range(urls[:])

		paths, err := sitemap.AbsPaths()
		if err != nil {
			logs.Print(err)
		}
		color.Primary.Printf("\nRequesting %d static URLs used in the sitemap.xml\n", len(paths))
		sitemap.Success.Range(paths[:])

		html3s, err := sitemap.AbsHtml3s()
		if err != nil {
			logs.Println(err)
		}
		color.Primary.Printf("\nRequesting %d static URLs used by the HTML3 text mode\n", len(html3s))
		sitemap.Success.Range(html3s[:])
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(testCmd)
}
