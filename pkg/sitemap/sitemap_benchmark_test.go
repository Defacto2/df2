package sitemap_test

import (
	"log"
	"testing"

	"github.com/Defacto2/df2/pkg/sitemap"
)

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := sitemap.Create(nil, ""); err != nil {
			log.Print(err)
		}
	}
}
