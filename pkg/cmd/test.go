package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/spf13/cobra"
)

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

		fmt.Printf("Requesting the <title> of %d random files from %d public records\n", pingCount, total)
		wg := &sync.WaitGroup{}
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				url = strings.TrimSpace(url)
				s, code, err := sitemap.GetTitle(url)
				if err != nil {
					logs.Fatal(err)
				}
				fmt.Printf("%s\t%d  ↳ %s\n", url, code, s)
				wg.Done()
			}(url)
		}
		wg.Wait()
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(testCmd)
}
