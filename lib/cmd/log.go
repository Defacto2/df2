package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	"github.com/hako/durafmt"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:    "log",
	Short:  "Display the df2 error log",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		logs.Printf("%v%v %v\n", color.Cyan.Sprint("log file"), color.Red.Sprint(":"), logs.Filepath())
		f, err := os.Open(logs.Filepath())
		logs.Check(err)
		scanner := bufio.NewScanner(f)
		c := 0
		scanner.Text()
		for scanner.Scan() {
			c++
			s := strings.SplitN(scanner.Text(), " ", 5)
			t, err := time.Parse("2006/01/02 15:04:05", strings.Join(s[0:2], " "))
			if err != nil {
				fmt.Printf("%d. %v\n", c, scanner.Text())
				continue
			}
			// todo get system local timezone and set it here
			// OR log file to use UTC
			logs.Printf("%v\n", t)
			duration := durafmt.Parse(time.Since(t)).LimitFirstN(1)
			fmt.Printf("%v %v ago  %v %s\n", color.Secondary.Sprintf("%d.", c), duration, color.Info.Sprint(s[2]), strings.Join(s[3:], " "))
		}
		if err := scanner.Err(); err != nil {
			logs.Check(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
