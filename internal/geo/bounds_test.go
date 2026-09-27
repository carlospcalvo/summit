package geo

import (
	"testing"
)

func TestBoundsEmpty(t *testing.T) {
	var zero Bounds
	if !zero.Empty() {
		t.Error("the zero Bounds must report Empty so callers can distinguish an unavailable bound from a coordinate at 0,0")
	}

	var b Bounds
	b.Add(LatLon{Lat: 0, Lon: 0})
	if b.Empty() {
		t.Error("a Bounds holding the origin must not report Empty: 0,0 is a real coordinate, not an absent one")
	}
	if !b.Contains(LatLon{Lat: 0, Lon: 0}) {
		t.Error("a Bounds holding the origin must contain the origin")
	}
}

func TestBoundsSinglePoint(t *testing.T) {
	p := LatLon{Lat: 45.9237, Lon: 6.8694}

	var b Bounds
	b.Add(p)

	if got, want := b.MinLat, p.Lat; got != want {
		t.Errorf("MinLat = %v, want %v", got, want)
	}
	if got, want := b.MaxLat, p.Lat; got != want {
		t.Errorf("MaxLat = %v, want %v", got, want)
	}
	if got, want := b.MinLon, p.Lon; got != want {
		t.Errorf("MinLon = %v, want %v", got, want)
	}
	if got, want := b.MaxLon, p.Lon; got != want {
		t.Errorf("MaxLon = %v, want %v", got, want)
	}
}

func TestBoundsOrdinaryTrack(t *testing.T) {
	points := []LatLon{
		{Lat: 45.9237, Lon: 6.8694},
		{Lat: 45.9310, Lon: 6.8830},
		{Lat: 45.9100, Lon: 6.9000},
		{Lat: 45.9500, Lon: 6.8500},
	}

	var b Bounds
	for _, p := range points {
		b.Add(p)
	}

	if got, want := b.MinLat, 45.9100; got != want {
		t.Errorf("MinLat = %v, want %v", got, want)
	}
	if got, want := b.MaxLat, 45.9500; got != want {
		t.Errorf("MaxLat = %v, want %v", got, want)
	}
	if got, want := b.MinLon, 6.8500; got != want {
		t.Errorf("MinLon = %v, want %v", got, want)
	}
	if got, want := b.MaxLon, 6.9000; got != want {
		t.Errorf("MaxLon = %v, want %v", got, want)
	}

	for _, p := range points {
		if !b.Contains(p) {
			t.Errorf("Contains(%+v) = false, want true", p)
		}
	}
	if b.Contains(LatLon{Lat: 46.5, Lon: 7.5}) {
		t.Error("Contains(46.5, 7.5) = true, want false")
	}
}

// Bounds uses plain min/max on longitude and does not split a track that
// crosses the antimeridian. A track at +179.5 and -179.5 therefore reports the
// full globe, -180 to 180, rather than the 1 degree it actually spans.
//
// This is documented, not accidental: Phase 0 only needs bounds for reporting,
// and deciding how to render a wrapped span is a map-fitting concern for
// Phase 4. The invariant that must hold today is enclosure.
func TestBoundsAntimeridianEnclosesEveryPoint(t *testing.T) {
	points := []LatLon{
		{Lat: 0, Lon: 179.5},
		{Lat: 0, Lon: -179.5},
		{Lat: 0.5, Lon: 179.9},
	}

	var b Bounds
	for _, p := range points {
		b.Add(p)
	}

	if got, want := b.MinLon, -179.5; got != want {
		t.Errorf("MinLon = %v, want %v (naive min/max, see test comment)", got, want)
	}
	if got, want := b.MaxLon, 179.9; got != want {
		t.Errorf("MaxLon = %v, want %v (naive min/max, see test comment)", got, want)
	}

	for _, p := range points {
		if !b.Contains(p) {
			t.Errorf("Contains(%+v) = false, want true", p)
		}
	}
}

func TestBoundsAddIsOrderIndependent(t *testing.T) {
	forward := []LatLon{{Lat: 1, Lon: 2}, {Lat: 3, Lon: 4}, {Lat: -1, Lon: 0}}

	var a, b Bounds
	for _, p := range forward {
		a.Add(p)
	}
	for i := len(forward) - 1; i >= 0; i-- {
		b.Add(forward[i])
	}

	if a != b {
		t.Errorf("bounds depend on insertion order: %+v vs %+v", a, b)
	}
}
