package database_test

import (
	"fmt"

	"github.com/Defacto2/df2/lib/database"
)

func ExampleInit() {
	init := database.Init()
	fmt.Print(init.Port)
	// Output: 3306
}

func ExampleConnect() {
	db := database.Connect()
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
	if _, err := database.ColTypes(database.Users); err != nil {
		fmt.Print(err)
	}
	// Output:
}

func ExampleLastUpdate() {
	if _, err := database.LastUpdate(); err != nil {
		fmt.Print(err)
	}
	// Output:
}

func ExampleTotal() {
	db := database.Connect()
	defer db.Close()
	s := "SELECT * FROM `files` WHERE `id` = '1'"
	i, err := database.Total(&s)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(i)
	// Output: 1
}

func ExampleWaiting() {
	i, err := database.Waiting()
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(i > 0)
	// Output: true
}
