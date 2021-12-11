package shrink_test

import (
	"fmt"

	"github.com/Defacto2/df2/lib/shrink"
)

func ExampleFiles() {
	if err := shrink.Files(); err != nil {
		fmt.Print("shrink files error")
	}
	// Output: shrink files error
}

func ExamplePreviews() {
	if err := shrink.Previews(); err != nil {
		fmt.Print("shrink previews error")
	}
	// Output:
}

func ExampleSQL() {
	if err := shrink.SQL(); err != nil {
		fmt.Print("shrink sql error")
	}
	// Output: shrink sql error
}
