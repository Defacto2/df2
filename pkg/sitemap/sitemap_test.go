package sitemap_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/sitemap"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"create", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sitemap.Create(); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
