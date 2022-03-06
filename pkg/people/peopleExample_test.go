package people_test

import (
	"log"

	"github.com/Defacto2/df2/pkg/people"
)

func ExampleFix() {
	const simulate = true
	if err := people.Fix(simulate); err != nil {
		log.Print(err)
	}
	// Output: no people fixes needed
}
