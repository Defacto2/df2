package scan_test

import (
	"os"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
)

func ExampleStats_Summary() {
	s := scan.Init(configger.Defaults())
	s.Summary(os.Stdout)
	// Output: nothing to do
	// ───────────────────────────────────────────────────
	// Total archives scanned: 0, time elapsed 0.0 seconds
}
