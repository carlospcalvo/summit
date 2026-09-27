// Package track owns Summit's track domain model and every statistic derived
// from it.
//
// A track is an ordered list of points plus the waypoints recorded alongside
// it. Two semantics here are load-bearing for all later phases: missing values
// are represented explicitly rather than as zero, and every point remembers
// which source GPX segment produced it so that distance and elevation
// calculations never bridge a recording gap.
//
// This package depends on geo for coordinate primitives and distance
// mathematics, and on nothing else. Parsing, persistence, and user interfaces
// all translate through this one model so they cannot disagree about what a
// track is.
package track

import "time"

// Track is a single recorded route with its ordered points and waypoints.
type Track struct {
	Name      string
	Points    []TrackPoint
	Waypoints []Waypoint
}

// TrackPoint is one recorded sample along a track.
//
// Elevation and Time are pointers because GPX makes both optional and a zero
// elevation or a zero time.Time would be indistinguishable from a real reading.
// Consumers must check for nil rather than assume a value.
//
// SourceSegment is the zero-based index of the <trkseg> element that produced
// this point. Two consecutive points with different SourceSegment values were
// not recorded continuously, so no edge between them may be measured.
type TrackPoint struct {
	Lat           float64
	Lon           float64
	Elevation     *float64
	Time          *time.Time
	SourceSegment int
}

// Waypoint is a named point of interest recorded with a track, such as a trail
// junction or a summit marker. Optional fields follow the same nil convention
// as TrackPoint.
type Waypoint struct {
	Name      string
	Lat       float64
	Lon       float64
	Elevation *float64
	Time      *time.Time
}
