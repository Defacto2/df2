package stat_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/proof/internal/stat"
	"github.com/stretchr/testify/assert"
)

func TestProof_Summary(t *testing.T) {
	p := stat.Proof{}
	s := p.Summary("")
	assert.Contains(t, s, "nothing")
	s = p.Summary("1")
	assert.Contains(t, s, "")
	p, err := stat.Init(configger.Defaults())
	assert.Nil(t, err)
	p.Total = 5
	p.Count = 4
	p.Missing = 1
	s = p.Summary("1")
	assert.Contains(t, s, "Total proofs handled")
}
