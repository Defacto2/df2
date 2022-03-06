package scan_test

import (
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
)

func ExampleStats_Summary() {
	s := scan.Init()
	s.Summary()
	// Output: nothing to do
	// ───────────────────────────────────────────────────
	// Total archives scanned: 0, time elapsed 0.0 seconds
}
