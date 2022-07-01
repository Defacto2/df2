package people_test

import (
	"log"
	"os"

	"github.com/Defacto2/df2/pkg/people"
)

func ExampleFix() {
	// suppress dynamic output for this example
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer os.Stdout.Close()
	if err := people.Fix(); err != nil {
		log.Print(err)
	}
	// Output:
}
