package track

import "github.com/carlospcalvo/summit/internal/geo"

type Stats struct {
	PointCount                       int
	DistanceM                        float64
	Bounds                           geo.Bounds
	MinElevM, MaxElevM, GainM, LossM *float64 // nil = unavailable, never 0
}

func Analyze(t Track) Stats
