package sitemap_test

import (
	"log"
	"testing"

	"github.com/Defacto2/df2/lib/sitemap"
)

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := sitemap.Create(); err != nil {
			log.Print(err)
		}
	}
}

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
