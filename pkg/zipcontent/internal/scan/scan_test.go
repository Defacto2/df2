package scan_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
)

func TestInit(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		var want scan.Stats
		want.BasePath = "/opt/assets/downloads"
		got := scan.Init(configger.Defaults())
		if got.BasePath != want.BasePath {
			t.Errorf("Init().BasePath = %v, want %v", got.BasePath, want.BasePath)
		}
	})
}
