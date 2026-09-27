package gpx

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/carlospcalvo/summit/internal/track"
)

// fixture opens a GPX file from the shared testdata directory at the repo root.
func fixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("open fixture %s: %v", name, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func parseFixture(t *testing.T, name string) (track.Track, error) {
	t.Helper()
	tr, err := Parse(fixture(t, name))
	return tr, err
}

func mustParseFixture(t *testing.T, name string) track.Track {
	t.Helper()
	tr, err := parseFixture(t, name)
	if err != nil {
		t.Fatalf("Parse(%s): unexpected error: %v", name, err)
	}
	return tr
}

func mustTime(t *testing.T, label string, v *time.Time) time.Time {
	t.Helper()
	if v == nil {
		t.Fatalf("%s = nil, want a timestamp", label)
	}
	return *v
}

func TestParseFixtureShapes(t *testing.T) {
	tests := []struct {
		fixture      string
		wantName     string
		wantPoints   int
		wantSegments int
		wantWaypts   int
	}{
		{fixture: "single-point.gpx", wantName: "Single Point", wantPoints: 1, wantSegments: 1},
		{fixture: "flat-track.gpx", wantName: "Flat Track", wantPoints: 4, wantSegments: 1},
		{fixture: "multi-segment.gpx", wantName: "Multi Segment", wantPoints: 4, wantSegments: 2},
		{fixture: "missing-elevation.gpx", wantName: "Missing Elevation", wantPoints: 5, wantSegments: 1},
		{fixture: "missing-time.gpx", wantName: "Missing Time", wantPoints: 3, wantSegments: 1},
		{fixture: "with-waypoints.gpx", wantName: "With Waypoints", wantPoints: 2, wantSegments: 1, wantWaypts: 2},
		{fixture: "antimeridian.gpx", wantName: "Antimeridian", wantPoints: 3, wantSegments: 1},
		{fixture: "gpx10.gpx", wantName: "GPX 1.0 Track", wantPoints: 2, wantSegments: 1},
		{fixture: "example.gpx", wantName: "Chamonix Morning Hike", wantPoints: 6, wantSegments: 1, wantWaypts: 1},
		{fixture: "realistic.gpx", wantName: "Synthetic Ridge Loop", wantPoints: 2500, wantSegments: 2},
	}

	for _, tc := range tests {
		t.Run(tc.fixture, func(t *testing.T) {
			tr, err := parseFixture(t, tc.fixture)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if tr.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", tr.Name, tc.wantName)
			}
			if got := len(tr.Points); got != tc.wantPoints {
				t.Errorf("len(Points) = %d, want %d", got, tc.wantPoints)
			}
			if got := len(tr.Waypoints); got != tc.wantWaypts {
				t.Errorf("len(Waypoints) = %d, want %d", got, tc.wantWaypts)
			}

			seen := map[int]bool{}
			for i, p := range tr.Points {
				seen[p.SourceSegment] = true
				if p.SourceSegment < 0 || p.SourceSegment >= tc.wantSegments {
					t.Errorf("point %d SourceSegment = %d, want 0..%d", i, p.SourceSegment, tc.wantSegments-1)
				}
			}
			if got := len(seen); got != tc.wantSegments {
				t.Errorf("found %d distinct segment indices, want %d", got, tc.wantSegments)
			}
		})
	}
}

// GPX 1.0 and GPX 1.1 differ only in namespace, so the same document shape must
// produce the same track from either.
func TestParseAcceptsBothGPXVersions(t *testing.T) {
	v11 := mustParseFixture(t, "flat-track.gpx")
	v10 := mustParseFixture(t, "gpx10.gpx")

	if len(v10.Points) == 0 {
		t.Fatal("GPX 1.0 fixture produced no points")
	}
	if v10.Points[0].SourceSegment != 0 {
		t.Errorf("GPX 1.0 SourceSegment = %d, want 0", v10.Points[0].SourceSegment)
	}
	if v10.Points[0].Elevation == nil {
		t.Error("GPX 1.0 elevation = nil, want a value")
	}
	// The first two coordinates are identical in both fixtures.
	if v10.Points[0].Lat != v11.Points[0].Lat || v10.Points[0].Lon != v11.Points[0].Lon {
		t.Errorf("GPX 1.0 coordinates %+v, want %+v",
			v10.Points[0], v11.Points[0])
	}
}

// Phase 0 reads only the first <trk>. Later phases may widen this, so the
// behaviour is pinned by a test rather than left implicit.
func TestParseUsesOnlyTheFirstTrack(t *testing.T) {
	tr := mustParseFixture(t, "multiple-tracks.gpx")

	if tr.Name != "First Track" {
		t.Errorf("Name = %q, want %q", tr.Name, "First Track")
	}
	if got := len(tr.Points); got != 2 {
		t.Fatalf("len(Points) = %d, want 2 from the first track only", got)
	}
	for i, p := range tr.Points {
		if p.Lat < 45 || p.Lat > 46 {
			t.Errorf("point %d latitude %v came from the second track", i, p.Lat)
		}
	}
}

func TestParseSegmentIndices(t *testing.T) {
	tr := mustParseFixture(t, "multi-segment.gpx")

	want := []int{0, 0, 1, 1}
	for i, w := range want {
		if got := tr.Points[i].SourceSegment; got != w {
			t.Errorf("point %d SourceSegment = %d, want %d", i, got, w)
		}
	}
}

func TestParseOptionalElevation(t *testing.T) {
	tr := mustParseFixture(t, "missing-elevation.gpx")

	wantNil := []bool{false, true, false, false, true}
	for i, w := range wantNil {
		got := tr.Points[i].Elevation == nil
		if got != w {
			t.Errorf("point %d elevation present = %v, want %v", i, !got, w)
		}
	}
	if tr.Points[0].Elevation == nil || *tr.Points[0].Elevation != 100 {
		t.Errorf("point 0 elevation = %v, want 100", tr.Points[0].Elevation)
	}
}

func TestParseOptionalTime(t *testing.T) {
	tr := mustParseFixture(t, "missing-time.gpx")

	for i, wantNil := range []bool{false, true, false} {
		if got := tr.Points[i].Time == nil; got != wantNil {
			t.Errorf("point %d time present = %v, want %v", i, !got, !wantNil)
		}
	}
	want := time.Date(2024, time.June, 1, 8, 0, 20, 0, time.UTC)
	if got := mustTime(t, "point 2 time", tr.Points[2].Time); !got.Equal(want) {
		t.Errorf("point 2 time = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestParseWaypoints(t *testing.T) {
	tr := mustParseFixture(t, "with-waypoints.gpx")

	if len(tr.Waypoints) != 2 {
		t.Fatalf("len(Waypoints) = %d, want 2", len(tr.Waypoints))
	}

	first := tr.Waypoints[0]
	if first.Name != "Trailhead" {
		t.Errorf("Waypoints[0].Name = %q, want %q", first.Name, "Trailhead")
	}
	if first.Lat != 45.9237 || first.Lon != 6.8694 {
		t.Errorf("Waypoints[0] coords = %v,%v, want 45.9237,6.8694", first.Lat, first.Lon)
	}
	if first.Elevation == nil || *first.Elevation != 1035 {
		t.Errorf("Waypoints[0].Elevation = %v, want 1035", first.Elevation)
	}
	if first.Time == nil {
		t.Error("Waypoints[0].Time = nil, want a value")
	}

	second := tr.Waypoints[1]
	if second.Name != "Bridge" {
		t.Errorf("Waypoints[1].Name = %q, want %q", second.Name, "Bridge")
	}
	if second.Elevation != nil {
		t.Errorf("Waypoints[1].Elevation = %v, want nil", *second.Elevation)
	}
	if second.Time != nil {
		t.Errorf("Waypoints[1].Time = %v, want nil", second.Time)
	}
}

// Waypoints appear before <trk> in a valid GPX document and must still be
// picked up, because <trk> is not the root element.
func TestParseWaypointsBeforeTrack(t *testing.T) {
	tr := mustParseFixture(t, "example.gpx")

	if len(tr.Waypoints) != 1 {
		t.Fatalf("len(Waypoints) = %d, want 1", len(tr.Waypoints))
	}
	if tr.Waypoints[0].Name != "Chamonix Trailhead" {
		t.Errorf("Waypoints[0].Name = %q, want %q", tr.Waypoints[0].Name, "Chamonix Trailhead")
	}
}

func TestParseEmptyButValidDocument(t *testing.T) {
	tr, err := Parse(strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?><gpx version="1.1" creator="test" ` +
			`xmlns="http://www.topografix.com/GPX/1/1"></gpx>`))
	if err != nil {
		t.Fatalf("Parse(empty document): %v", err)
	}
	if tr.Name != "" {
		t.Errorf("Name = %q, want an empty name", tr.Name)
	}
	if len(tr.Points) != 0 {
		t.Errorf("len(Points) = %d, want 0", len(tr.Points))
	}
	if tr.Points != nil {
		t.Error("Points = non-nil slice, want nil so callers can tell no data from empty data")
	}
}

// A zero byte file is not a GPX document at all, so it is an error rather than
// an empty track. This is a deliberate Phase 0 decision.
func TestParseEmptyReaderIsAnError(t *testing.T) {
	if _, err := Parse(strings.NewReader("")); err == nil {
		t.Fatal("Parse(empty reader) = nil error, want an error")
	}
}

func TestParseRejectsMalformedXML(t *testing.T) {
	_, err := parseFixture(t, "malformed.gpx")
	if err == nil {
		t.Fatal("Parse(malformed.gpx) = nil error, want an error")
	}

	var syntaxErr *xml.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("error %v does not wrap an XML syntax error, so errors.As(*xml.SyntaxError) fails", err)
	}
	if !strings.Contains(err.Error(), "trkpt") {
		t.Errorf("error %q does not name the offending element", err)
	}
}

func TestParseRejectsInvalidCoordinates(t *testing.T) {
	_, err := parseFixture(t, "invalid-coordinate.gpx")
	if err == nil {
		t.Fatal("Parse(invalid-coordinate.gpx) = nil error, want an error")
	}
	if !errors.Is(err, ErrInvalidCoordinate) {
		t.Errorf("error %v does not satisfy errors.Is(err, ErrInvalidCoordinate)", err)
	}
	if !strings.Contains(err.Error(), "95.5") {
		t.Errorf("error %q does not quote the offending latitude", err)
	}
}

func TestParseRejectsBadAttributes(t *testing.T) {
	// strconv.ParseFloat accepts "NaN" and "Inf", so a range check written as
	// lat >= -90 && lat <= 90 would let them through. Every coordinate case
	// below must be rejected.
	tests := []struct {
		name string
		doc  string
		want error
	}{
		{
			name: "latitude above ninety",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="90.1" lon="0"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "latitude below minus ninety",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="-91" lon="0"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "longitude above one eighty",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="0" lon="180.5"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "longitude below minus one eighty",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="0" lon="-181"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "latitude not a number",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="forty five" lon="0"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "latitude is NaN",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="NaN" lon="0"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "longitude is infinite",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="0" lon="Inf"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "latitude attribute missing",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lon="6"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "longitude attribute missing",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="45"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "latitude attribute empty",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="" lon="6"/></trkseg></trk></gpx>`,
			want: ErrInvalidCoordinate,
		},
		{
			name: "elevation is not a number",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="45" lon="6"><ele>high</ele></trkpt></trkseg></trk></gpx>`,
			want: ErrInvalidElevation,
		},
		{
			name: "timestamp is not a time",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="45" lon="6"><time>yesterday</time></trkpt></trkseg></trk></gpx>`,
			want: ErrInvalidTimestamp,
		},
		{
			name: "timestamp has no date",
			doc:  `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg><trkpt lat="45" lon="6"><time>08:00:00</time></trkpt></trkseg></trk></gpx>`,
			want: ErrInvalidTimestamp,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tc.doc))
			if err == nil {
				t.Fatalf("Parse = nil error, want %v", tc.want)
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("error %v does not satisfy errors.Is(err, %v)", err, tc.want)
			}
		})
	}
}

func TestParseNeverPanics(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace", "   \n\t  "},
		{"declaration only", `<?xml version="1.0"?>`},
		{"not xml at all", "this is a text file, not a track"},
		{"truncated mid element", `<gpx version="1.1"><trk><trkseg><trkpt lat="45"`},
		{"mismatched close tag", `<gpx version="1.1"><trk></trkseg></gpx>`},
		{"binary noise", "\x00\x01\x02\xff\xfe"},
		{"only a root", `<gpx/>`},
		{"trkseg with no points", `<gpx version="1.1"><trk><name>x</name><trkseg></trkseg></trk></gpx>`},
		{"bare trkpt", `<trkpt lat="45" lon="6"/>`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The only assertion is that we get here at all.
			_, _ = Parse(strings.NewReader(tc.in))
		})
	}
}

// Parse takes an io.Reader, so it must work on a stream it cannot seek or size
// up front, and it must surface a read failure rather than a partial track.
func TestParseStreamsAndPropagatesReadErrors(t *testing.T) {
	pr, pw := io.Pipe()
	go func() {
		_, _ = pw.Write([]byte(`<gpx version="1.1"><trk><trkseg>` +
			`<trkpt lat="45.0" lon="6.0"><ele>100</ele></trkpt>`))
		_ = pw.CloseWithError(errors.New("sensor disconnected"))
	}()

	_, err := Parse(pr)
	if err == nil {
		t.Fatal("Parse = nil error, want the reader failure")
	}
	if !strings.Contains(err.Error(), "sensor disconnected") {
		t.Errorf("error %q does not mention the underlying read failure", err)
	}
}

// syntheticGPX builds a valid in-memory document with n points in one segment.
// Benchmarks use it so that no large fixture has to be committed.
func syntheticGPX(n int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<gpx version="1.1" creator="bench" xmlns="http://www.topografix.com/GPX/1/1">` +
		`<trk><name>Bench</name><trkseg>`)
	for i := range n {
		lat := 45.0 + float64(i)*0.0001
		lon := 6.0 + float64(i)*0.0001
		fmt.Fprintf(&b, `<trkpt lat="%.4f" lon="%.4f"><ele>1000.5</ele></trkpt>`, lat, lon)
	}
	b.WriteString(`</trkseg></trk></gpx>`)
	return b.String()
}
