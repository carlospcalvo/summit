package gpx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkParseRealisticFixture(b *testing.B) {
	path := filepath.Join("..", "..", "testdata", "realistic.gpx")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("read fixture: %v", err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := Parse(strings.NewReader(string(data))); err != nil {
			b.Fatalf("Parse: %v", err)
		}
	}
}

// The realistic fixture is only 2,500 points, small enough that per-point
// allocation shows up in the profile. This scales the same document shape well
// past any real recording so the shape of the cost is visible.
func BenchmarkParseGenerated(b *testing.B) {
	for _, points := range []int{2_500, 50_000, 250_000} {
		doc := syntheticGPX(points)

		b.Run(fmt.Sprintf("%d_points", points), func(b *testing.B) {
			b.SetBytes(int64(len(doc)))
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if _, err := Parse(strings.NewReader(doc)); err != nil {
					b.Fatalf("Parse: %v", err)
				}
			}
		})
	}
}
