package database_test

import (
	"fmt"
	"os"

	"github.com/Defacto2/df2/pkg/database"
)

func ExampleInit() {
	init := database.Init()
	fmt.Print(init.Port)
	// Output: 3306
}

func ExampleConnect() {
	db := database.Connect(os.Stdout)
	defer db.Close()
	fmt.Print(db.Stats().WaitCount)
	// Output: 0
}

func ExampleConnErr() {
	db, err := database.ConnErr()
	if err != nil {
		fmt.Print(err)
	}
	defer db.Close()
	// Output:
}

func ExampleConnInfo() {
	s := database.ConnInfo()
	fmt.Print(s)
	// Output:
}

func ExampleColTypes() {
	if _, err := database.ColTypes(os.Stdout, database.Netresources); err != nil {
		fmt.Print(err)
	}
	// Output:
}

func ExampleLastUpdate() {
	if _, err := database.LastUpdate(os.Stdout); err != nil {
		fmt.Print(err)
	}
	// Output:
}

func ExampleTotal() {
	w := os.Stdout
	db := database.Connect(w)
	defer db.Close()
	s := "SELECT * FROM `files` WHERE `id` = '1'"
	i, err := database.Total(w, &s)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(i)
	// Output: 1
}

func ExampleWaiting() {
	i, err := database.Waiting(os.Stdout)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(i > 0)
	// Output: true
}
