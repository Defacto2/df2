package zipcontent_test

import (
	"fmt"

	"github.com/Defacto2/df2/pkg/zipcontent"
)

func ExampleFix() {
	const printSummary = false
	if err := zipcontent.Fix(printSummary); err != nil {
		fmt.Println(err)
	}
	// Output:
}
