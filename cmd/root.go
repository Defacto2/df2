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
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ErrCfg    = errors.New("config cannot be empty")
	ErrCmd    = errors.New("invalid command, please use one of the available commands")
	ErrID     = errors.New("invalid id or uuid specified")
	ErrLogger = errors.New("logger cannot be nil")
	ErrNoID   = errors.New("requires an id or uuid argument")
)

var (
	cfg  configger.Config   // Enviroment variables for configuration.
	logr *zap.SugaredLogger // Zap sugared logger for printing and storing.
	pers arg.Persistant     // Persistant, command-line bool flags.
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
		db, err := database.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := run.New(db, os.Stdout, logr, cfg); err != nil {
			logr.Fatal(err)
		}
	},
}

// Execute is a Cobra command that adds all child commands to the root and
// sets the appropriate flags. It is called by main.main() and only needs
// to be called once in the rootCmd.
func Execute(log *zap.SugaredLogger, c configger.Config) error {
	if log == nil {
		return ErrLogger
	}
	if c == (configger.Config{}) {
		return ErrCfg
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	cfg = c
	if err := rootCmd.Execute(); err != nil {
		logr.Warnln(err)
		if e := err.Error(); strings.Contains(e, "required flag(s) \"name\"") {
			fmt.Fprintln(os.Stdout,
				"see Examples for usage or run to list setting choices:",
				color.Bold.Sprintf("%s config info", rootCmd.CommandPath()))
		}
		return nil
	}
	// config.Check()
	return nil
}

func init() { //nolint:gochecknoinits
	// as init runs before anything, check for logger to avoid panics
	if logr == nil {
		logr = logger.Production().Sugar()
	}
	cobra.OnInitialize()
	rootCmd.AddGroup(&cobra.Group{ID: "group1", Title: "Admin:"})
	rootCmd.AddGroup(&cobra.Group{ID: "group2", Title: "Drive:"})
	rootCmd.AddGroup(&cobra.Group{ID: "group3", Title: "Remote:"})
	rootCmd.PersistentFlags().BoolVar(&pers.ASCII, "ascii", false,
		"suppress all ANSI color feedback")
	rootCmd.PersistentFlags().BoolVar(&pers.Quiet, "quiet", false,
		"suppress all feedback except for errors")
	rootCmd.PersistentFlags().BoolVarP(&pers.Version, "version", "v", false,
		"version and information for this program")
	rootCmd.PersistentFlags().BoolVar(&pers.Panic, "panic", false,
		"panic in the disco")
	if err := rootCmd.PersistentFlags().MarkHidden("panic"); err != nil {
		logr.Fatal(err)
	}
}
