package arg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gookit/color"
)

var ErrNoCmd = errors.New("no command argument was provided")

func Targets() []string {
	return []string{"all", "download", "emulation", "image"}
}

type Approve struct {
	Verbose bool
}

type Clean struct {
	Delete   bool
	Humanise bool
	MakeDirs bool
	Target   string
}

type Config struct {
	InfoSize  bool
	Name      string
	Overwrite bool
}

type Apis struct {
	Refresh bool
	Pouet   bool
	SyncDos bool
	SyncWin bool
}

type Demozoo struct {
	All       bool // scan for all proofs, not just new submissions
	Overwrite bool // overwrite all existing assets
	New       bool
	ID        string // auto-generated id or a uuid
	Extract   []string
	Ping      uint
	Download  uint
	Releaser  uint // id for a releaser
}

// Execute global flags.
type Execute struct {
	Panic   bool // Enable panic errors to help debug.
	ASCII   bool // Ascii is placeholder for Cobra to store the PersistentFlag value*
	Quiet   bool // Quiet is placeholder for Cobra to store the PersistentFlag value*
	Version bool // Version is placeholder for Cobra to store the PersistentFlag value*
	// * but the quiet flag is handled by main.go.
}

type Group struct {
	Counts   bool
	Cronjob  bool
	Forcejob bool
	Init     bool
	Progress bool
	Filter   string
	Format   string
}

type People struct {
	Cronjob  bool
	Forcejob bool
	Progress bool
	Filter   string
	Format   string
}

type Proof struct {
	ID          string // auto-generated id or a uuid
	All         bool   // scan for all proofs, not just new uploads
	HideMissing bool   // hide proofs that are missing their file download
	Overwrite   bool   // overwrite all existing images
}

type Recent struct {
	Compress bool
	Limit    uint
}

type TestSite struct {
	LocalHost bool
}

type ZipCmmt struct {
	ASCII   bool
	Unicode bool
	OW      bool
}

// Invalid returns instructions for invalid command arguments and exits with an error code.
func Invalid(w io.Writer, arg string, args ...string) error {
	if arg == "" {
		return ErrNoCmd
	}
	s := ""
	if len(args) == 0 {
		s = fmt.Sprintf("%s %s", color.Warn.Sprint("invalid command"),
			color.Bold.Sprintf("\"%s\"", arg))
	}
	if len(args) > 0 {
		s = fmt.Sprintf("%s %s",
			color.Warn.Sprint("invalid command"),
			color.Bold.Sprintf("\"%s %s\"", arg, args[0]))
	}
	s += fmt.Sprint("\n" + color.Warn.Sprint("please use one of the Available Commands shown above"))
	fmt.Fprintln(w, s)
	return nil
}

// FilterFlag compairs the value of the filter flag against the list of slice values.
func FilterFlag(w io.Writer, t any, flag, val string) {
	if val == "" {
		return
	}
	if t, ok := t.([]string); ok {
		sup := false
		for _, value := range t {
			if value == val || (val == value[:1]) {
				sup = true
				break
			}
		}
		if sup {
			return
		}
		fmt.Fprintf(w, "%s %s\n%s %s\n", color.Warn.Sprintf("unsupported --%s flag value", flag),
			color.Bold.Sprintf("%q", val),
			color.Warn.Sprint("available flag values"),
			color.Primary.Sprint(strings.Join(t, ",")))
		os.Exit(1)
	}
}

func CleanOpts(a ...string) string {
	const opts = "\noptions: "
	if len(a) == 0 {
		return opts + "MISSING"
	}
	sort.Strings(a)
	return opts + strings.Join(a, ", ")
}
