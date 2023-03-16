package database_test

import (
	"fmt"
	"log"
	"os"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
)

func ExampleInit() {
	init := database.Init()
	fmt.Print(init.Port)
	// Output: 3306
}

func ExampleConnect() {
	cfg := configger.Defaults()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
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
	if _, err := database.ColTypes(nil, os.Stdout, database.Netresources); err != nil {
		fmt.Print(err)
	}
	// Output:
}

func ExampleLastUpdate() {
	if _, err := database.LastUpdate(nil); err != nil {
		fmt.Print(err)
	}
	// Output:
}

func ExampleTotal() {
	w := os.Stdout
	s := "SELECT * FROM `files` WHERE `id` = '1'"
	i, err := database.Total(nil, w, &s)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(i)
	// Output: 1
}

func ExampleWaiting() {
	i, err := database.Waiting(nil)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(i > 0)
	// Output: true
}
