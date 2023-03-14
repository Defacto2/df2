package recent_test

import (
	"log"
	"testing"

	"github.com/Defacto2/df2/pkg/recent"
)

func BenchmarkCreate(b *testing.B) {
	const limit = 10
	for i := 0; i < b.N; i++ {
		if err := recent.List(nil, limit, true); err != nil {
			log.Print(err)
		}
	}
}
