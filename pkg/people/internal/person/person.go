// Package person contains the shared Person object for individual people.
package person

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"text/template"
)

var ErrFilter = errors.New("invalid filter used")

// Person represent people who are listed as authors on the website.
type Person struct {
	ID   string // ID used in URLs to link to the person.
	Nick string // Nick of the person.
	HR   bool   // Inject a HR element to separate a collection of groups.
}

type Persons []Person

// Tempate saves to dest the HTML used by the website to list people.
func (p Persons) Template(dest, tmpl string) error {
	t, err := template.New("h2").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parse h2 template: %w", err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("parse create: %w", err)
	}
	defer f.Close()
	if err = t.Execute(f, p); err != nil {
		return fmt.Errorf("parse template execute: %w", err)
	}
	return nil
}

// Tempate writes the HTML used by the website to list people.
func (p Persons) TemplateW(w io.Writer, tmpl string) error {
	if w == nil {
		w = io.Discard
	}
	t, err := template.New("h2").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parse h2 template: %w", err)
	}
	var bb bytes.Buffer
	nr := bufio.NewWriter(&bb)
	if err = t.Execute(nr, p); err != nil {
		return fmt.Errorf("parse h2 execute template: %w", err)
	}
	if err := nr.Flush(); err != nil {
		return fmt.Errorf("parse writer flush: %w", err)
	}
	fmt.Fprintln(w, bb.String())
	return nil
}
