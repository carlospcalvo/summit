package geo

import (
	"math"
	"testing"
)

// Expected distances below are derived from the mean Earth radius used by
// Haversine, so they pin the current formula rather than a hand calculation:
//
//	1 degree of longitude at the equator = 2*pi*R/360 = 111195.080234 m
//	quarter meridian                   = pi*R/2      = 10007557.221 m
//	antipodes                          = pi*R        = 20015114.442 m
func TestHaversine(t *testing.T) {
	tests := []struct {
		name string
		a    LatLon
		b    LatLon
		want float64
		tol  float64
	}{
		{
			name: "identical points",
			a:    LatLon{Lat: 45.9237, Lon: 6.8694},
			b:    LatLon{Lat: 45.9237, Lon: 6.8694},
			want: 0,
			tol:  0,
		},
		{
			name: "one degree of longitude at the equator",
			a:    LatLon{Lat: 0, Lon: 0},
			b:    LatLon{Lat: 0, Lon: 1},
			want: 111195.080234,
			tol:  0.01,
		},
		{
			name: "one degree of latitude at the equator",
			a:    LatLon{Lat: 0, Lon: 0},
			b:    LatLon{Lat: 1, Lon: 0},
			want: 111195.080234,
			tol:  0.01,
		},
		{
			name: "quarter meridian",
			a:    LatLon{Lat: 0, Lon: 0},
			b:    LatLon{Lat: 90, Lon: 0},
			want: 10007557.221,
			tol:  0.01,
		},
		{
			name: "antipodes, where an unclamped half-angle would become NaN",
			a:    LatLon{Lat: 0, Lon: 0},
			b:    LatLon{Lat: 0, Lon: 180},
			want: 20015114.442,
			tol:  0.01,
		},
		{
			name: "antipodes across the other side of the planet",
			a:    LatLon{Lat: 45, Lon: 6},
			b:    LatLon{Lat: -45, Lon: -174},
			want: 20015114.442,
			tol:  0.01,
		},
		{
			name: "london to paris",
			a:    LatLon{Lat: 51.5074, Lon: -0.1278},
			b:    LatLon{Lat: 48.8566, Lon: 2.3522},
			want: 343556.535,
			tol:  1,
		},
		{
			name: "short hop across the antimeridian",
			a:    LatLon{Lat: 0, Lon: 179},
			b:    LatLon{Lat: 0, Lon: -179},
			want: 222390.160,
			tol:  0.01,
		},
		{
			name: "short hop along a parallel",
			a:    LatLon{Lat: 45, Lon: 6},
			b:    LatLon{Lat: 45.01, Lon: 6},
			want: 1111.950802,
			tol:  0.01,
		},
		{
			name: "distance is symmetric",
			a:    LatLon{Lat: 51.5074, Lon: -0.1278},
			b:    LatLon{Lat: 48.8566, Lon: 2.3522},
			want: 343556.535,
			tol:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Haversine(tc.a, tc.b)
			if math.Abs(got-tc.want) > tc.tol {
				t.Errorf("Haversine(%+v, %+v) = %.6f m, want %.6f m +/- %g",
					tc.a, tc.b, got, tc.want, tc.tol)
			}
			if reverse := Haversine(tc.b, tc.a); math.Abs(reverse-got) > 1e-6 {
				t.Errorf("Haversine is not symmetric: %.9f vs %.9f", got, reverse)
			}
		})
	}
}

func TestHaversineAlwaysFinite(t *testing.T) {
	tests := []struct {
		name string
		a    LatLon
		b    LatLon
	}{
		{"antipodal equator", LatLon{Lat: 0, Lon: 0}, LatLon{Lat: 0, Lon: 180}},
		{"antipodal poles", LatLon{Lat: 90, Lon: 0}, LatLon{Lat: -90, Lon: 180}},
		{"identical extreme coordinates", LatLon{Lat: -90, Lon: -180}, LatLon{Lat: -90, Lon: -180}},
		{"extreme to extreme", LatLon{Lat: -90, Lon: -180}, LatLon{Lat: 90, Lon: 180}},
		{"adjacent longitudes at the pole", LatLon{Lat: 90, Lon: 0}, LatLon{Lat: 90, Lon: 1}},
		{"antimeridian pair at high latitude", LatLon{Lat: 85, Lon: 179.999}, LatLon{Lat: 85, Lon: -179.999}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Haversine(tc.a, tc.b)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("Haversine(%+v, %+v) = %v, want a finite result", tc.a, tc.b, got)
			}
		})
	}
}
