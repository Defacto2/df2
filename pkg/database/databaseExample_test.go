package database_test

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
)

func ExampleConnect() {
	db, err := database.Connect(conf.Defaults())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Print(db.Stats().WaitCount)
	// Output: 0
}

func ExampleColumns() {
	db, err := database.Connect(conf.Defaults())
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Columns(db, io.Discard, database.Netresources); err != nil {
		log.Fatal(err)
	}
	// Output:
}

func ExampleTotal() {
	db, err := database.Connect(conf.Defaults())
	if err != nil {
		log.Fatal(err)
	}
	s := "SELECT * FROM `files` WHERE `id` = '1'"
	i, err := database.Total(db, os.Stdout, &s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(i)
	// Output: 1
}

func ExampleWaiting() {
	db, err := database.Connect(conf.Defaults())
	if err != nil {
		log.Fatal(err)
	}
	i, err := database.Waiting(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(i >= 0)
	// Output: true
}
