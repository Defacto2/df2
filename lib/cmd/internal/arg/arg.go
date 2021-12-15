package arg

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
)

type Clean struct {
	Delete   bool
	Humanise bool
	MakeDirs bool
	Target   string
}

type Proof struct {
	Id          string // auto-generated id or a uuid
	All         bool   // scan for all proofs, not just new uploads
	HideMissing bool   // hide proofs that are missing their file download
	Overwrite   bool   // overwrite all existing images
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
