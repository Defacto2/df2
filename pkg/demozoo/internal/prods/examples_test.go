package prods_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
)

const (
	testDir  = "../../../../tests/json/"
	channel1 = 1
	channel2 = 2
	channel3 = 3
)

var example1, example2, example3 prods.ProductionsAPIv1 //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	c1 := make(chan prods.ProductionsAPIv1)
	c2 := make(chan prods.ProductionsAPIv1)
	c3 := make(chan prods.ProductionsAPIv1)
	go loadExample(channel1, c1)
	go loadExample(channel2, c2)
	go loadExample(channel3, c3)
	example1, example2, example3 = <-c1, <-c2, <-c3
}

// loadExample data from Demozoo.
func loadExample(r int, c chan prods.ProductionsAPIv1) {
	var name string
	switch r {
	case channel1:
		name = "1"
	case channel2:
		name = "188796"
	case channel3:
		name = "267300"
	default:
		log.Print(fmt.Errorf("load r %d: %w", r, ErrVal))
	}
	path, err := filepath.Abs(filepath.Join(testDir, fmt.Sprintf("record_%s.json", name)))
	if err != nil {
		log.Print(fmt.Errorf("path %q: %w", path, err))
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
	}
	var dz prods.ProductionsAPIv1
	if err := json.Unmarshal(data, &dz); err != nil {
		log.Print(fmt.Errorf("load json unmarshal: %w", err))
	}
	c <- dz
}
