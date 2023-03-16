package role_test

import (
	"fmt"
	"os"

	"github.com/Defacto2/df2/pkg/people/internal/role"
)

func ExampleRole() {
	r := role.Musicians
	fmt.Println(r)
	// Output: musicians
}

func ExampleList() {
	r := role.Musicians
	_, total, err := role.List(nil, os.Stdout, r)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%s found? %v", r, total > 0)
	// Output: musicians found? true
}
