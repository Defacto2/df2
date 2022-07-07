package arg

import (
	"log"
	"os"
	"sort"
	"strings"

	"github.com/gookit/color"
)

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

type ZipCmmt struct {
	ASCII   bool
	Unicode bool
	OW      bool
}

// FilterFlag compairs the value of the filter flag against the list of slice values.
func FilterFlag(t any, flag, val string) {
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
		if !sup {
			log.Printf("%s %s\n%s %s\n", color.Warn.Sprintf("unsupported --%s flag value", flag),
				color.Bold.Sprintf("%q", val),
				color.Warn.Sprint("available flag values"),
				color.Primary.Sprint(strings.Join(t, ",")))
			os.Exit(1)
		}
	}
}

func CleanOpts(a ...string) string {
	sort.Strings(a)
	return "\noptions: " + strings.Join(a, ",")
}
