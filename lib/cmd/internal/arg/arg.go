package arg

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
)

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

type Demozoo struct {
	All       bool // scan for all proofs, not just new submissions
	Overwrite bool // overwrite all existing assets
	Simulate  bool
	New       bool
	Refresh   bool
	Sync      bool
	ID        string // auto-generated id or a uuid
	Extract   []string
	Ping      uint
	Download  uint
}

// Execute global flags.
type Execute struct {
	Filename string // Filename of the config file.
	Quiet    bool   // Quiet mode to reduce text amount of stdout text.
	Panic    bool   // Enable panic errors to help debug.
	Simulate bool   // Simulate...
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
	Filter   string
	Format   string
	Progress bool
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

// FilterFlag compairs the value of the filter flag against the list of slice values.
func FilterFlag(t interface{}, flag, val string) {
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
			fmt.Printf("%s %s\n%s %s\n", color.Warn.Sprintf("unsupported --%s flag value", flag),
				color.Bold.Sprintf("%q", val),
				color.Warn.Sprint("available flag values"),
				color.Primary.Sprint(strings.Join(t, ",")))
			os.Exit(1)
		}
	}
}
