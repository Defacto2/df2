package person

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Defacto2/df2/pkg/people/internal/role"
	"github.com/spf13/viper"
)

var ErrFilter = errors.New("invalid filter used")

// Person represent people who are listed as authors on the website.
type Person struct {
	ID   string // ID used in URLs to link to the person.
	Nick string // Nick of the person.
	Hr   bool   // Inject a HR element to separate a collection of groups.
}

type Persons []Person

// Tempate creates the HTML used by the website to list people.
func (p Persons) Template(filename, tpl string, filter string) error {
	t, err := template.New("h2").Parse(tpl)
	if err != nil {
		return fmt.Errorf("parse h2 template: %w", err)
	}
	if filename == "" {
		var buf bytes.Buffer
		wr := bufio.NewWriter(&buf)
		if err = t.Execute(wr, p); err != nil {
			return fmt.Errorf("parse h2 execute template: %w", err)
		}
		if err := wr.Flush(); err != nil {
			return fmt.Errorf("parse writer flush: %w", err)
		}
		fmt.Println(buf.String())
		return nil
	}
	switch role.Roles(filter) {
	case role.Artists, role.Coders, role.Musicians, role.Writers:
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		if err != nil {
			return fmt.Errorf("parse create: %w", err)
		}
		defer f.Close()
		if err = t.Execute(f, p); err != nil {
			return fmt.Errorf("parse template execute: %w", err)
		}
	case role.Everyone:
		return fmt.Errorf("parse %v: %w", filter, ErrFilter)
	}
	return nil
}
