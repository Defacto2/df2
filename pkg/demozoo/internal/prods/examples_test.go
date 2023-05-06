package prods_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
)

var ErrVal = errors.New("unknown value")

const (
	ch1 = 1
	ch2 = 2
	ch3 = 3
)

func testDir() string {
	return filepath.Join("..", "..", "..", "..", "testdata", "json")
}

var example1, example2, example3 prods.ProductionsAPIv1 //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	v1 := make(chan prods.ProductionsAPIv1)
	v2 := make(chan prods.ProductionsAPIv1)
	v3 := make(chan prods.ProductionsAPIv1)
	go loadExample(ch1, v1)
	go loadExample(ch2, v2)
	go loadExample(ch3, v3)
	example1, example2, example3 = <-v1, <-v2, <-v3
}

// loadExample data from Demozoo.
func loadExample(r int, c chan prods.ProductionsAPIv1) {
	name := ""
	switch r {
	case ch1:
		name = "1"
	case ch2:
		name = "188796"
	case ch3:
		name = "267300"
	default:
		log.Print(fmt.Errorf("load r %d: %w", r, ErrVal))
	}
	path, err := filepath.Abs(filepath.Join(testDir(), fmt.Sprintf("record_%s.json", name)))
	if err != nil {
		log.Print(fmt.Errorf("path %q: %w", path, err))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Print(err)
	}
	dz := prods.ProductionsAPIv1{}
	if err := json.Unmarshal(data, &dz); err != nil {
		log.Print(fmt.Errorf("load json unmarshal: %w", err))
	}
	c <- dz
}
