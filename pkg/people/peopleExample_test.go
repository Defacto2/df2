package people_test

import (
	"log"

	"github.com/Defacto2/df2/pkg/people"
)

func ExampleFix() {
	if err := people.Fix(nil, nil); err != nil {
		log.Print(err)
	}
	// Output:
}
