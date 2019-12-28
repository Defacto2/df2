package archive

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"testing"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func BenchmarkCalculate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Extract("/Users/ben/Downloads/Miitopia_EUR_MULTi6-TRSI.zip", "")
	}
}
