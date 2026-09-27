// Package geo owns Summit's coordinate primitives and distance mathematics.
//
// It is the leaf of the dependency graph: track analysis depends on it, and it
// depends on nothing from Summit. It deliberately knows nothing about GPX
// documents, files, databases, or the desktop application, so its results stay
// reusable and independently testable.
//
// Coordinates are always latitude first, then longitude, in degrees. Unit
// conversion to GeoJSON longitude/latitude order happens at the frontend map
// boundary, never here.
package geo

import "math"

// earthRadiusM is the IUGG mean Earth radius in metres. A spherical model is
// accurate to roughly 0.3% against the WGS84 ellipsoid, which is well inside
// the tolerance of consumer GPS tracks.
const earthRadiusM = 6371008.8

// LatLon is a geographic coordinate in degrees, latitude before longitude.
type LatLon struct {
	Lat float64
	Lon float64
}

// Haversine returns the great-circle distance in metres between two
// coordinates. Identical coordinates return 0. The result is always finite
// because every inverse trigonometric input is clamped.
func Haversine(a, b LatLon) float64 {
	lat1 := radians(a.Lat)
	lat2 := radians(b.Lat)
	dLat := lat2 - lat1
	dLon := radians(b.Lon - a.Lon)

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)

	return 2 * earthRadiusM * math.Asin(math.Sqrt(clampUnit(h)))
}

// radians converts degrees to radians.
func radians(deg float64) float64 {
	return deg * math.Pi / 180
}

// clampUnit constrains a squared half-angle to [0,1]. Rounding can push the
// value a few ulps outside that range, and math.Sqrt or math.Asin would then
// return NaN in application-facing data.
func clampUnit(v float64) float64 {
	return math.Min(1, math.Max(0, v))
}
