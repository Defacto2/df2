// Package arg has structures used to store the flag values used by the
// cobra.Command methods.
package arg

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gookit/color"
)

var (
	ErrFlagValue = errors.New("value for flag is not valid")
	ErrNoCmd     = errors.New("no command argument was provided")
)

func Targets() []string {
	return []string{
		"all",
		"download",
		"emulation",
		"image",
	}
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

type APIs struct {
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

// Persistent global flags.
type Persistent struct {
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

type Import struct {
	Insert bool
	Limit  uint
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
	Stdout  bool // Stdout writes any found zip comment to the stdout.
	Unicode bool // Unicode attempts to convert any CP-437 encoded comments to Unicode.
	OW      bool // OW overwrites any existing save zip comments, otherwise they're skipped.
}

// Invalid returns instructions for invalid command arguments and exits with an error code.
func Invalid(w io.Writer, arg string, args ...string) error {
	if arg == "" {
		return ErrNoCmd
	}
	if w == nil {
		w = io.Discard
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
func FilterFlag(w io.Writer, t any, flag, val string) error {
	if val == "" {
		return nil
	}
	if w == nil {
		w = io.Discard
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
			return nil
		}
		fmt.Fprintf(w, "%s %s\n%s %s\n", color.Warn.Sprintf("unsupported --%s flag value", flag),
			color.Bold.Sprintf("%q", val),
			color.Warn.Sprint("available flag values"),
			color.Primary.Sprint(strings.Join(t, ",")))
		return ErrFlagValue
	}
	return nil
}

func CleanOpts(a ...string) string {
	const opts = "\noptions: "
	if len(a) == 0 {
		return opts + "MISSING"
	}
	sort.Strings(a)
	return opts + strings.Join(a, ", ")
}
