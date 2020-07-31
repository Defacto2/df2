// about

// +build ignore

// This program generates blah. It can be invoked by running
// go generate ./...
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

func main() {
	ids := []uint{1, 188796, 267300}
	c := make(chan error)
	go get(ids[0], c)
	go get(ids[1], c)
	go get(ids[2], c)
	a, b, d := <-c, <-c, <-c
	if a != nil {
		log.Fatal(a)
	}
	if b != nil {
		log.Fatal(b)
	}
	if d != nil {
		log.Fatal(d)
	}
}

func get(id uint, c chan error) {
	f, err := demozoo.Fetch(id)
	if err != nil {
		c <- err
	}
	if !str.Piped() {
		logs.Printf("Demozoo ID %v, HTTP status %v\n", id, f.Status)
	}
	b, err := f.API.JSON()
	if err != nil {
		c <- err
	}
	p, err := filepath.Abs(fmt.Sprintf("../../tests/json/record_%d.json", id))
	if err != nil {
		c <- err
	}
	file, err := os.Create(p)
	if err != nil {
		c <- err
	}
	defer file.Close()
	i, err := file.Write(b)
	if err != nil {
		c <- err
	}
	fmt.Printf("id: %d, %d bytes written to %s\n", id, i, p)
	c <- nil
}
