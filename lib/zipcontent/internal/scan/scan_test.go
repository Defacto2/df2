package scan_test

import (
	"testing"

	"github.com/Defacto2/df2/lib/zipcontent/internal/scan"
)

func TestInit(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		var want scan.Stats
		want.BasePath = "/opt/assets/downloads"
		got := scan.Init()
		if got.BasePath != want.BasePath {
			t.Errorf("Init().BasePath = %v, want %v", got.BasePath, want.BasePath)
		}
	})
}
