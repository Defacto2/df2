package database

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_readPassword(t *testing.T) {

	// create a temporary file with the content EXAMPLE
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	text := []byte("EXAMPLE")
	if _, err = tmpFile.Write(text); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	tests := []struct {
		name   string
		pwPath string
		want   string
	}{
		{"empty", "", "password"},
		{"invalid", "/tmp/Test_readPassword/validfile", "password"},
		{"temp", tmpFile.Name(), "EXAMPLE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwPath = tt.pwPath
			if got := readPassword(); got != tt.want {
				t.Errorf("readPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}