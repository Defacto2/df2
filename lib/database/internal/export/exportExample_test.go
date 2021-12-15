package export_test

import (
	"fmt"

	"github.com/Defacto2/df2/lib/database/internal/export"
)

func ExampleTable() {
	fmt.Print(export.Files)
	// Output: files
}

func ExampleTbls() {
	s := export.Tbls()
	fmt.Print(s)
	// Output: files, groups, netresources, users
}
