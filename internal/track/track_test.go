package track

import (
	"math"
	"math/big"
	"testing"

	"github.com/carlospcalvo/summit/internal/geo"
)

func ele(v float64) *float64 { return &v }

// point builds a track point. A nil elev models a GPX point with no <ele>.
func point(lat, lon float64, seg int, elev *float64) TrackPoint {
	return TrackPoint{Lat: lat, Lon: lon, SourceSegment: seg, Elevation: elev}
}

// flatStep builds n points spaced 0.01 degrees of latitude apart, all at the
// same longitude, alternating between the given elevations. A nil entry means
// that point has no elevation.
func latLine(elevs []*float64, segments []int) []TrackPoint {
	pts := make([]TrackPoint, 0, len(elevs))
	for i, e := range elevs {
		pts = append(pts, point(45+0.01*float64(i), 6, segments[i], e))
	}
	return pts
}

func sameSegment(n int) []int {
	segs := make([]int, n)
	return segs
}

// assertClose fails when two metre values differ by more than tol metres.
func assertClose(t *testing.T, label string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.6f m, want %.6f m +/- %g", label, got, want, tol)
	}
}

func requireEle(t *testing.T, label string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s = nil, want %.1f m", label, want)
	}
	assertClose(t, label, *got, want, 1e-9)
}

func TestAnalyzeEmptyTrack(t *testing.T) {
	stats := Analyze(Track{})

	if stats.PointCount != 0 {
		t.Errorf("PointCount = %d, want 0", stats.PointCount)
	}
	if stats.DistanceM != 0 {
		t.Errorf("DistanceM = %v, want 0", stats.DistanceM)
	}
	if !stats.Bounds.Empty() {
		t.Errorf("Bounds = %+v, want an empty Bounds", stats.Bounds)
	}
	for name, v := range map[string]*float64{
		"MinElevM": stats.MinElevM,
		"MaxElevM": stats.MaxElevM,
		"GainM":    stats.GainM,
		"LossM":    stats.LossM,
	} {
		if v != nil {
			t.Errorf("%s = %v, want nil for a track with no data", name, *v)
		}
	}
}

func TestAnalyzeSinglePoint(t *testing.T) {
	stats := Analyze(Track{Points: []TrackPoint{
		point(45.9237, 6.8694, 0, ele(1385.5)),
	}})

	if stats.PointCount != 1 {
		t.Errorf("PointCount = %d, want 1", stats.PointCount)
	}
	if stats.DistanceM != 0 {
		t.Errorf("DistanceM = %v, want 0 for a single point", stats.DistanceM)
	}
	requireEle(t, "MinElevM", stats.MinElevM, 1385.5)
	requireEle(t, "MaxElevM", stats.MaxElevM, 1385.5)

	// One point means no continuous pair, so gain and loss are unavailable.
	// Reporting 0 would claim the track was perfectly flat.
	if stats.GainM != nil {
		t.Errorf("GainM = %v, want nil for a single point", *stats.GainM)
	}
	if stats.LossM != nil {
		t.Errorf("LossM = %v, want nil for a single point", *stats.LossM)
	}
}

func TestAnalyzeSkipsSegmentBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		segments []int
		wantM    float64
		tol      float64
	}{
		{
			name:     "one segment, three continuous edges",
			segments: sameSegment(4),
			wantM:    3335.852407,
			tol:      0.001,
		},
		{
			name:     "two segments, edge across the break is not measured",
			segments: []int{0, 0, 1, 1},
			wantM:    2223.901605,
			tol:      0.001,
		},
		{
			name:     "three segments, two breaks",
			segments: []int{0, 0, 1, 1, 2, 2},
			wantM:    3335.852407,
			tol:      0.001,
		},
		{
			name:     "interleaved indices, no edge is continuous",
			segments: []int{0, 1, 0, 1, 0, 1},
			wantM:    0,
			tol:      0,
		},
		{
			name:     "a new segment per sample, no edge is continuous",
			segments: []int{0, 1, 2, 3, 4, 5},
			wantM:    0,
			tol:      0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stats := Analyze(Track{Points: latLine(
				[]*float64{nil, nil, nil, nil, nil, nil}, tc.segments,
			)})

			if stats.PointCount != len(tc.segments) {
				t.Errorf("PointCount = %d, want %d", stats.PointCount, len(tc.segments))
			}
			assertClose(t, "DistanceM", stats.DistanceM, tc.wantM, tc.tol)
		})
	}
}

// A pair of segments 200 km apart must never be joined by a synthetic edge.
func TestAnalyzeDoesNotBridgeARecordingGap(t *testing.T) {
	stats := Analyze(Track{Points: []TrackPoint{
		point(45.00, 6.00, 0, ele(500)),
		point(45.01, 6.00, 0, ele(520)),
		point(46.50, 8.00, 1, ele(900)),
		point(46.51, 8.00, 1, ele(940)),
	}})

	assertClose(t, "DistanceM", stats.DistanceM, 2223.901605, 0.001)
	if stats.DistanceM > 10000 {
		t.Errorf("DistanceM = %.1f m, which means the segment break was bridged", stats.DistanceM)
	}
}

func TestAnalyzeElevationSequences(t *testing.T) {
	consts := []*float64{ele(100), ele(100), ele(100), ele(100)}
	up := []*float64{ele(100), ele(150), ele(220), ele(220)}
	down := []*float64{ele(400), ele(300), ele(180), ele(180)}
	mixed := []*float64{ele(100), ele(300), ele(250), ele(500)}

	tests := []struct {
		name    string
		elevs   []*float64
		min     float64
		max     float64
		gain    float64
		loss    float64
		wantNil bool
	}{
		{name: "flat", elevs: consts, min: 100, max: 100, gain: 0, loss: 0},
		{name: "climbing", elevs: up, min: 100, max: 220, gain: 120, loss: 0},
		{name: "descending", elevs: down, min: 180, max: 400, gain: 0, loss: 220},
		{name: "mixed", elevs: mixed, min: 100, max: 500, gain: 450, loss: 50},
		{name: "no elevation at all", elevs: []*float64{nil, nil, nil, nil}, wantNil: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stats := Analyze(Track{Points: latLine(tc.elevs, sameSegment(4))})

			if tc.wantNil {
				for name, v := range map[string]*float64{
					"MinElevM": stats.MinElevM,
					"MaxElevM": stats.MaxElevM,
					"GainM":    stats.GainM,
					"LossM":    stats.LossM,
				} {
					if v != nil {
						t.Errorf("%s = %v, want nil", name, *v)
					}
				}
				return
			}

			requireEle(t, "MinElevM", stats.MinElevM, tc.min)
			requireEle(t, "MaxElevM", stats.MaxElevM, tc.max)
			requireEle(t, "GainM", stats.GainM, tc.gain)
			requireEle(t, "LossM", stats.LossM, tc.loss)
		})
	}
}

func TestAnalyzeMissingElevationIsNotZero(t *testing.T) {
	// Elevations 100, gap, 150, 160, gap. Only the 150 -> 160 pair is a
	// continuous edge. Substituting 0 for the gaps would report 160 m of gain
	// and 150 m of loss.
	stats := Analyze(Track{Points: latLine(
		[]*float64{ele(100), nil, ele(150), ele(160), nil},
		sameSegment(5),
	)})

	requireEle(t, "MinElevM", stats.MinElevM, 100)
	requireEle(t, "MaxElevM", stats.MaxElevM, 160)
	requireEle(t, "GainM", stats.GainM, 10)
	requireEle(t, "LossM", stats.LossM, 0)
}

func TestAnalyzeGainUnavailableWhenNoContinuousPair(t *testing.T) {
	stats := Analyze(Track{Points: latLine(
		[]*float64{ele(100), nil, ele(150), nil, ele(900)},
		sameSegment(5),
	)})

	requireEle(t, "MinElevM", stats.MinElevM, 100)
	requireEle(t, "MaxElevM", stats.MaxElevM, 900)
	if stats.GainM != nil {
		t.Errorf("GainM = %v, want nil when no adjacent pair both have elevation", *stats.GainM)
	}
	if stats.LossM != nil {
		t.Errorf("LossM = %v, want nil when no adjacent pair both have elevation", *stats.LossM)
	}
}

func TestAnalyzeElevationSkipsSegmentBoundaries(t *testing.T) {
	// A 400 m jump between segments is a recording gap, not a climb.
	stats := Analyze(Track{Points: latLine(
		[]*float64{ele(100), ele(140), ele(900), ele(905)},
		[]int{0, 0, 1, 1},
	)})

	requireEle(t, "MinElevM", stats.MinElevM, 100)
	requireEle(t, "MaxElevM", stats.MaxElevM, 905)
	requireEle(t, "GainM", stats.GainM, 40)
	requireEle(t, "LossM", stats.LossM, 0)
}

func TestAnalyzeBounds(t *testing.T) {
	stats := Analyze(Track{Points: []TrackPoint{
		point(45.9237, 6.8694, 0, nil),
		point(45.9310, 6.8830, 0, nil),
		point(45.9100, 6.9000, 0, nil),
	}})

	want := geo.Bounds{MinLat: 45.9100, MinLon: 6.8694, MaxLat: 45.9310, MaxLon: 6.9000}
	if stats.Bounds != want {
		t.Errorf("Bounds = %+v, want %+v", stats.Bounds, want)
	}
}

func TestAnalyzeBoundsOfSinglePoint(t *testing.T) {
	stats := Analyze(Track{Points: []TrackPoint{
		point(45.9237, 6.8694, 0, nil),
	}})

	want := geo.Bounds{MinLat: 45.9237, MinLon: 6.8694, MaxLat: 45.9237, MaxLon: 6.8694}
	if stats.Bounds != want {
		t.Errorf("Bounds = %+v, want %+v", stats.Bounds, want)
	}
}

// A long track accumulates one floating point addition per edge. The reference
// here is an exact 200 bit sum of the same per-edge distances, so this pins
// accumulation drift to well under a millimetre over tens of kilometres.
func TestAnalyzeLongTrackDoesNotDrift(t *testing.T) {
	const n = 20000

	pts := make([]TrackPoint, n)
	exact := new(big.Float).SetPrec(200)
	for i := range n {
		th := 2 * math.Pi * float64(i) / float64(n)
		pts[i] = point(
			45.92+0.05*math.Cos(th)+0.004*math.Sin(3*th),
			6.87+0.07*math.Sin(th)+0.003*math.Cos(2*th),
			i/1000, ele(1000+250*math.Sin(2*th)),
		)
	}

	for i := 0; i < n-1; i++ {
		edge := geo.Haversine(
			geo.LatLon{Lat: pts[i].Lat, Lon: pts[i].Lon},
			geo.LatLon{Lat: pts[i+1].Lat, Lon: pts[i+1].Lon},
		)
		exact.Add(exact, big.NewFloat(edge))
	}
	want, _ := exact.Float64()

	stats := Analyze(Track{Points: pts})

	if rel := math.Abs(stats.DistanceM-want) / want; rel > 1e-9 {
		t.Errorf("DistanceM = %.9f m, exact sum is %.9f m, relative drift %.3e exceeds 1e-9",
			stats.DistanceM, want, rel)
	}
	if stats.PointCount != n {
		t.Errorf("PointCount = %d, want %d", stats.PointCount, n)
	}
}

// Analyze must not mutate the track it is given; the Wails service and the
// profile renderer will hold the same Track concurrently in later phases.
func TestAnalyzeDoesNotMutateInput(t *testing.T) {
	original := Track{Name: "keep", Points: latLine(
		[]*float64{ele(100), ele(110), ele(90), ele(105)}, sameSegment(4),
	)}

	before := make([]TrackPoint, len(original.Points))
	copy(before, original.Points)
	Analyze(original)

	for i := range before {
		got, want := original.Points[i], before[i]
		if got.Lat != want.Lat || got.Lon != want.Lon || got.SourceSegment != want.SourceSegment {
			t.Fatalf("point %d coordinates mutated: %+v, want %+v", i, got, want)
		}
		if (got.Elevation == nil) != (want.Elevation == nil) {
			t.Fatalf("point %d elevation presence mutated: %+v, want %+v", i, got, want)
		}
		if got.Elevation != nil && *got.Elevation != *want.Elevation {
			t.Fatalf("point %d elevation mutated: %v, want %v", i, *got.Elevation, *want.Elevation)
		}
	}
	if original.Name != "keep" {
		t.Errorf("Name = %q, want %q", original.Name, "keep")
	}
}
