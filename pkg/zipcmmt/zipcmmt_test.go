package zipcmmt_test

import (
	"fmt"

	"github.com/Defacto2/df2/pkg/zipcmmt"
)

func ExampleFix() {
	const (
		ascii     = false
		unicode   = false
		overwrite = false
		summary   = false
	)
	if err := zipcmmt.Fix(nil, nil, ascii, unicode, overwrite, summary); err != nil {
		fmt.Println(err)
	}
	// Output:
}
