package person_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/people/internal/person"
	"github.com/stretchr/testify/assert"
)

const tmpl = `{{range .}}<option value="{{.Nick}}" label="{{.Nick}}">{{end}}`

func TestPersons_Template(t *testing.T) {
	t.Parallel()
	p := person.Persons{}
	err := p.Template("", "")
	assert.NotNil(t, err)
	err = p.Template("", tmpl)
	assert.NotNil(t, err)
	err = p.Template(os.TempDir(), tmpl)
	assert.NotNil(t, err)
	name := filepath.Join(os.TempDir(), "testpersons_template.template")
	p = person.Persons{
		person.Person{
			ID:   "joe",
			Nick: "Joe",
			HR:   false,
		},
	}
	err = p.Template(name, tmpl)
	assert.Nil(t, err)
	defer os.Remove(name)
}

func TestPersons_TemplateW(t *testing.T) {
	t.Parallel()
	p := person.Persons{}
	err := p.TemplateW(nil, "")
	assert.Nil(t, err)
	err = p.TemplateW(io.Discard, tmpl)
	assert.Nil(t, err)
	bb := bytes.Buffer{}
	p = person.Persons{
		person.Person{
			ID:   "joe",
			Nick: "Joe Bloggs",
			HR:   false,
		},
	}
	err = p.TemplateW(&bb, tmpl)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), `<option value="Joe Bloggs"`)
}
