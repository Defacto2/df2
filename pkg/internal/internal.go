// Package internal is intended as package wide test data.
package internal

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	ErrNotDir = errors.New("testdata does not point to a directory")
)

const (
	DZ  = "demozoo"
	Img = "images"
	PNG = "test.png"
	Zip = "test.zip"

	File00UUID       = "00000000-0000-0000-0000-000000000000" // File00UUID is a blank UUID.
	File01UUID       = "c8cd0b4c-2f54-11e0-8827-cc1607e15609" // UUIDFile01 is the UUID value of the file record with ID 1.
	File01Name       = "Defacto2_Cracktro_Pack-2007.7z"       // File01Name is the filename of the file record with ID 1.
	File01Save       = "https://defacto2.net/d/9b1c6"         // File01Save is the URL to download the file with record ID 1.
	File01URL        = "https://defacto2.net/f/9b1c6"         // File01URL is the URL to the file record with ID 1.
	RandStr          = "epnGyShu6kPPv1bkhmkK"                 // RandStr is a nonsensical alphanumeric string.
	TestDemozooMSDOS = 164151                                 // TestDemozooMSDOS is a Demozoo ID for an MSDOS production.
	TestDemozooC64   = 309360                                 // TestDemozooC64 is a Demozoo ID for an Commodore 64 production.
)

func Testdata(subs int) string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dir := wd
	n := 1
	for n <= subs {
		n++
		dir = filepath.Join(dir, "..")
	}
	dir = filepath.Join(dir, "testdata")
	st, err := os.Stat(dir)
	if err != nil {
		log.Fatalln(fmt.Errorf("%w: %s", err, dir))
	}
	if !st.IsDir() {
		log.Fatalln(fmt.Errorf("%w: %s", ErrNotDir, dir))
	}
	return dir
}

func TestArchives(subs int) string {
	dir := Testdata(subs)
	return filepath.Join(dir, DZ)
}

func TestImg(subs int) string {
	dir := Testdata(subs)
	return filepath.Join(dir, Img, PNG)
}

func TestZip(subs int) string {
	dir := Testdata(subs)
	return filepath.Join(dir, DZ, Zip)
}
