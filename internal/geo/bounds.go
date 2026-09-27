package geo

import "math"

type Bounds struct {
	MinLat float64
	MinLon float64
	MaxLat float64
	MaxLon float64
	isSet  bool
}

func (b *Bounds) Add(p LatLon) {
	if !b.isSet {
		b.MinLat, b.MinLon, b.MaxLat, b.MaxLon = p.Lat, p.Lon, p.Lat, p.Lon
		b.isSet = true
		return
	}
	b.MinLat, b.MaxLat = math.Min(b.MinLat, p.Lat), math.Max(b.MaxLat, p.Lat)
	b.MinLon, b.MaxLon = math.Min(b.MinLon, p.Lon), math.Max(b.MaxLon, p.Lon)
}

func (b Bounds) Empty() bool {
	// zero value is Empty; 0,0 added is NOT
	return !b.isSet
}

func (b Bounds) Contains(p LatLon) bool {
	if b.Empty() {
		return false
	}

	return b.isWitihnLat(p.Lat) && b.isWitihnLon(p.Lon)
}

func (b Bounds) isWitihnLat(lat float64) bool {
	return b.MinLat <= lat && b.MaxLat >= lat
}

func (b Bounds) isWitihnLon(lon float64) bool {
	return b.MinLon <= lon && b.MaxLon >= lon
}
