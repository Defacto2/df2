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
