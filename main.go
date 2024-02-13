package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

//go:embed assets/brand.txt
var brand []byte

type ExitCode int // ExitCode is the exit code for this program.

const (
	NoExit       ExitCode = iota - 1 // NoExit is a special case to indicate the program should not exit.
	ExitOK                           // ExitOK is the exit code for a successful run.
	GenericError                     // GenericError is the exit code for a generic error.
	UsageError                       // UsageError is the exit code for an incorrect command line argument or usage.
)

func main() {

	app := &cli.App{
		Name:      name(),
		Version:   "2.0.0",
		Compiled:  time.Now(),
		Authors:   authors(),
		Copyright: "© 2024 Defacto2 & Ben Garrett",
		HelpName:  "df2",
		Usage:     "demonstrate available API",
		UsageText: "contrive - demonstrating the available API",
		ArgsUsage: "[args and such]", // ?
		Flags:     flags(),
		Action: func(cCtx *cli.Context) error {
			name := "Nefertiti"
			if cCtx.NArg() > 0 {
				name = cCtx.Args().Get(0)
			}
			if cCtx.String("lang") == "spanish" {
				fmt.Println("Hola", name)
			} else {
				fmt.Println("Hello", name)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:        "demozoo new",
				Aliases:     []string{"dzn"},
				Category:    "demozoo.org",
				Usage:       "interact with the Demozoo submissions",
				UsageText:   "demozoo - interact with the Demozoo submissions",
				Description: "Interact with the Demozoo submissions.",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("added task: ", cCtx.Args().First())
					return nil
				},
			},
			{
				Name:     "demozoo id",
				Aliases:  []string{"dzi"},
				Category: "demozoo.org",
				Usage:    "replace any empty data cells of a local file with linked demozoo data",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("completed task: ", cCtx.Args().First())
					return nil
				},
			},
			{
				Name:     "demozoo test",
				Aliases:  []string{"dzt"},
				Category: "demozoo.org",
				Usage:    "fetch and display a production record from the demozoo API",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("completed task: ", cCtx.Args().First())
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func authors() []*cli.Author {
	bg := cli.Author{
		Name:  "Ben Garrett",
		Email: "contact@defacto2.net",
	}
	return []*cli.Author{
		&bg,
	}
}

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			// Name:  "lang",
			// Value: "english",
			// Usage: "language for the greeting",
		},
	}
}

func name() string {
	b := strings.Builder{}
	io.WriteString(&b, string(brand))
	fmt.Fprintf(&b, "\n\thttps://github.com/Defacto2/df2")
	fmt.Fprintf(&b, "\n\n\tDefacto2 CLI")
	return b.String()
}
