// Package cmd handles the commandline user interface and interactions.
//
//nolint:gochecknoglobals
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/config"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ErrCmd  = errors.New("invalid command, please use one of the available commands")
	ErrID   = errors.New("invalid id or uuid specified")
	ErrNoID = errors.New("requires an id or uuid argument")
)

var (
	cfg configger.Config   // Enviroment variables for configuration.
	flg arg.Execute        // Global, command line flags.
	log *zap.SugaredLogger // Zap sugared logger for printing and storing.
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "df2",
	Short: About,
	Long: fmt.Sprintf("%s\n%v\n%v",
		color.Info.Sprint(About),
		Copyright(),
		color.Primary.Sprint(URL)),
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.New(os.Stdout, log); err != nil {
			log.Fatal(err)
		}
	},
}

// Execute is a Cobra command that adds all child commands to the root and
// sets the appropriate flags. It is called by main.main() and only needs
// to be called once in the rootCmd.
func Execute(log *zap.SugaredLogger, c configger.Config) error {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	cfg = c
	if err := rootCmd.Execute(); err != nil {
		log.Warn(err)
		if e := err.Error(); strings.Contains(e, "required flag(s) \"name\"") {
			fmt.Fprintln(os.Stdout,
				"see Examples for usage or run to list setting choices:",
				color.Bold.Sprintf("%s config info", rootCmd.CommandPath()))
		}
		return nil
	}
	config.Check()
	return nil
}

func init() { //nolint:gochecknoinits
	cobra.OnInitialize()
	rootCmd.AddGroup(&cobra.Group{ID: "group1", Title: "Admin:"})
	rootCmd.AddGroup(&cobra.Group{ID: "group2", Title: "Drive:"})
	rootCmd.AddGroup(&cobra.Group{ID: "group3", Title: "Remote:"})
	rootCmd.PersistentFlags().BoolVar(&flg.ASCII, "ascii", false,
		"suppress all ANSI color feedback")
	rootCmd.PersistentFlags().BoolVar(&flg.Quiet, "quiet", false,
		"suppress all feedback except for errors")
	rootCmd.PersistentFlags().BoolVarP(&flg.Version, "version", "v", false,
		"version and information for this program")
	rootCmd.PersistentFlags().BoolVar(&flg.Panic, "panic", false,
		"panic in the disco")
	if err := rootCmd.PersistentFlags().MarkHidden("panic"); err != nil {
		log.Fatal(err)
	}
}
