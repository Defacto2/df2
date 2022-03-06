package role_test

import (
	"fmt"

	"github.com/Defacto2/df2/pkg/people/internal/role"
)

func ExampleRole() {
	r := role.Musicians
	fmt.Println(r)
	// Output: musicians
}

func ExampleList() {
	r := role.Musicians
	_, total, err := role.List(r)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%d %s", total, r)
	// Output: 452 musicians
}
