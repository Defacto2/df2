package prompt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/prompt"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) { //nolint:tparallel
	t.Parallel()
	s, err := prompt.Read(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "", s)

	stdin := bytes.Buffer{}
	tests := []struct {
		name      string
		input     string
		wantInput string
		wantErr   bool
	}{
		{"empty", "", "", false},
		{"hello", "hello", "hello", false},
		{"trim", "        hello", "hello", false},
		{"sentence", "I am hello world.", "I am hello world.", false},
		{"nl", "\n\t\n\t\tb", "", false},
	}
	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			if _, err := stdin.Write([]byte(tt.input)); err != nil {
				t.Error(err)
			}
			s, err := prompt.Read(&stdin)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if s != tt.wantInput {
				t.Errorf("Read() = %v, want %v", s, tt.wantInput)
			}
		})
	}
}

func TestYN(t *testing.T) {
	t.Parallel()
	b, err := prompt.YN(nil, "", false)
	assert.Nil(t, err)
	assert.Equal(t, b, false)
	b, err = prompt.YN(io.Discard, "blah-blah", true)
	assert.Nil(t, err)
	assert.Equal(t, b, true)
}
