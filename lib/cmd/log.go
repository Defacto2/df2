// nolint:gochecknoglobals
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/color" //nolint:misspell
	"github.com/hako/durafmt"
	"github.com/spf13/cobra"

	"github.com/Defacto2/df2/lib/logs"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Short:   "Display the df2 error log",
	Aliases: []string{},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		logs.Printf("%v%v %v\n",
			color.Cyan.Sprint("log file"),
			color.Red.Sprint(":"),
			logs.Filepath(logs.Filename))
		f, err := os.Open(logs.Filepath(logs.Filename))
		if err != nil {
			logs.Fatal(err)
		}
		scanner := bufio.NewScanner(f)
		c := 0
		scanner.Text()
		const maxSplit = 5
		for scanner.Scan() {
			c++
			s := strings.SplitN(scanner.Text(), " ", maxSplit)
			t, err := time.Parse("2006/01/02 15:04:05", strings.Join(s[0:2], " "))
			if err != nil {
				fmt.Printf("%d. %v\n", c, scanner.Text())
				continue
			}
			duration := durafmt.Parse(time.Since(t)).LimitFirstN(1)
			fmt.Printf("%v %v ago  %v %s\n",
				color.Secondary.Sprintf("%d.", c),
				duration, color.Info.Sprint(s[2]),
				strings.Join(s[3:], " "))
		}
		if err := scanner.Err(); err != nil {
			logs.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(logCmd)
}
