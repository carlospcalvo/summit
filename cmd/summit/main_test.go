package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func fixturePath(name string) string {
	return filepath.Join("..", "..", "testdata", name)
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	err = run(args, &out, &errOut)
	return out.String(), errOut.String(), err
}

func TestInspectFormatsKnownTracks(t *testing.T) {
	tests := []struct {
		fixture string
		want    string
	}{
		{
			fixture: "flat-track.gpx",
			want: strings.Join([]string{
				"Track: Flat Track",
				"Points: 4",
				"Distance: 3.34 km",
				"Elevation gain: 0 m",
				"Elevation loss: 0 m",
				"Minimum: 100 m",
				"Maximum: 100 m",
				"",
			}, "\n"),
		},
		{
			// One point means no continuous pair, so gain and loss are
			// unavailable. Zero would falsely claim a perfectly flat track.
			fixture: "single-point.gpx",
			want: strings.Join([]string{
				"Track: Single Point",
				"Points: 1",
				"Distance: 0.00 km",
				"Elevation gain: n/a",
				"Elevation loss: n/a",
				"Minimum: 1386 m",
				"Maximum: 1386 m",
				"",
			}, "\n"),
		},
		{
			// Gaps in the elevation series must not be printed as 0.
			fixture: "missing-elevation.gpx",
			want: strings.Join([]string{
				"Track: Missing Elevation",
				"Points: 5",
				"Distance: 4.45 km",
				"Elevation gain: 10 m",
				"Elevation loss: 0 m",
				"Minimum: 100 m",
				"Maximum: 160 m",
				"",
			}, "\n"),
		},
		{
			// No point carries an elevation at all.
			fixture: "no-elevation.gpx",
			want: strings.Join([]string{
				"Track: No Elevation",
				"Points: 2",
				"Distance: 1.11 km",
				"Elevation gain: n/a",
				"Elevation loss: n/a",
				"Minimum: n/a",
				"Maximum: n/a",
				"",
			}, "\n"),
		},
		{
			// The break between the two segments must not be measured.
			fixture: "multi-segment.gpx",
			want: strings.Join([]string{
				"Track: Multi Segment",
				"Points: 4",
				"Distance: 2.22 km",
				"Elevation gain: 60 m",
				"Elevation loss: 0 m",
				"Minimum: 500 m",
				"Maximum: 940 m",
				"",
			}, "\n"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.fixture, func(t *testing.T) {
			stdout, stderr, err := runCLI(t, "inspect", fixturePath(tc.fixture))
			if err != nil {
				t.Fatalf("run: %v (stderr %q)", err, stderr)
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want nothing on a successful run", stderr)
			}
			if stdout != tc.want {
				t.Errorf("stdout =\n%s\nwant\n%s", stdout, tc.want)
			}
		})
	}
}

// The name printed comes from the file name, without directory or extension,
// when the GPX document has no <trk><name>.
func TestInspectFallsBackToTheFileName(t *testing.T) {
	tests := []struct {
		fixture string
		as      string
		want    string
	}{
		{fixture: "no-name.gpx", as: "Ridgeline.gpx", want: "Track: Ridgeline"},
		{fixture: "no-name.gpx", as: "Col de la Croix.gpx", want: "Track: Col de la Croix"},
		{fixture: "no-name.gpx", as: "UPPER.CASE.GPX", want: "Track: UPPER.CASE"},
	}

	for _, tc := range tests {
		t.Run(tc.as, func(t *testing.T) {
			data, err := os.ReadFile(fixturePath(tc.fixture))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			path := filepath.Join(t.TempDir(), tc.as)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatalf("write temp fixture: %v", err)
			}

			stdout, _, err := runCLI(t, "inspect", path)
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			if !strings.HasPrefix(stdout, tc.want+"\n") {
				t.Errorf("stdout first line = %q, want prefix %q", firstLine(stdout), tc.want)
			}
		})
	}
}

func TestInspectRejectsFilesWithoutAGPXExtension(t *testing.T) {
	data, err := os.ReadFile(fixturePath("flat-track.gpx"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	path := filepath.Join(t.TempDir(), "track.txt")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write temp fixture: %v", err)
	}

	_, _, err = runCLI(t, "inspect", path)
	if !errors.Is(err, errUsage) {
		t.Errorf("run = %v, want a usage error for a non .gpx path", err)
	}
}

// The realistic fixture is the Phase 0 gate: real shapes, real magnitudes, and
// no synthetic edge across the segment break.
func TestInspectRealisticFixture(t *testing.T) {
	stdout, stderr, err := runCLI(t, "inspect", fixturePath("realistic.gpx"))
	if err != nil {
		t.Fatalf("run: %v (stderr %q)", err, stderr)
	}

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 7 {
		t.Fatalf("got %d lines, want 7:\n%s", len(lines), stdout)
	}

	kv := parseFields(t, lines)
	if kv["Track"] != "Synthetic Ridge Loop" {
		t.Errorf("Track = %q", kv["Track"])
	}
	if kv["Points"] != "2500" {
		t.Errorf("Points = %q, want 2500", kv["Points"])
	}

	distance, ok := parseMeasurement(t, kv["Distance"], `^([\d.]+) km$`)
	if !ok {
		t.Fatalf("Distance = %q, want a value with two decimals in km", kv["Distance"])
	}
	if distance <= 20 || distance >= 60 {
		t.Errorf("Distance = %v km, want a plausible loop between 20 and 60 km", distance)
	}

	gain, ok := parseMeasurement(t, kv["Elevation gain"], `^([\d.]+) m$`)
	if !ok {
		t.Fatalf("Elevation gain = %q, want whole metres", kv["Elevation gain"])
	}
	loss, ok := parseMeasurement(t, kv["Elevation loss"], `^([\d.]+) m$`)
	if !ok {
		t.Fatalf("Elevation loss = %q, want whole metres", kv["Elevation loss"])
	}
	if gain <= 0 || loss <= 0 {
		t.Errorf("gain %v m and loss %v m, want both positive on a rolling loop", gain, loss)
	}

	minEle, ok := parseMeasurement(t, kv["Minimum"], `^([\d.]+) m$`)
	if !ok {
		t.Fatalf("Minimum = %q, want whole metres", kv["Minimum"])
	}
	maxEle, ok := parseMeasurement(t, kv["Maximum"], `^([\d.]+) m$`)
	if !ok {
		t.Fatalf("Maximum = %q, want whole metres", kv["Maximum"])
	}
	if minEle >= maxEle {
		t.Errorf("Minimum %v m >= Maximum %v m", minEle, maxEle)
	}
	if minEle < 650 || maxEle > 1400 {
		t.Errorf("elevation range %v..%v m falls outside the generated 690..1360 m band", minEle, maxEle)
	}
}

func TestInspectReportsErrorsOnStderr(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "file does not exist", args: []string{"inspect", fixturePath("no-such-file.gpx")}},
		{name: "malformed document", args: []string{"inspect", fixturePath("malformed.gpx")}},
		{name: "invalid coordinate", args: []string{"inspect", fixturePath("invalid-coordinate.gpx")}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(t, tc.args...)
			if err == nil {
				t.Fatal("run = nil error, want a processing failure")
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want nothing written on failure", stdout)
			}
			if !strings.Contains(strings.ToLower(stderr), "summit") &&
				!strings.Contains(stderr, "gpx") &&
				!strings.Contains(stderr, fixturePath(tc.args[1])) {
				t.Errorf("stderr = %q, want a human readable explanation", stderr)
			}
		})
	}
}

func TestUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no arguments", args: nil},
		{name: "no subcommand", args: []string{"flat-track.gpx"}},
		{name: "unknown subcommand", args: []string{"frobnicate", "flat-track.gpx"}},
		{name: "inspect without a path", args: []string{"inspect"}},
		{name: "inspect with an extra argument", args: []string{"inspect", "a.gpx", "b.gpx"}},
		{name: "unknown flag", args: []string{"inspect", "--verbose", "a.gpx"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(t, tc.args...)
			if !errors.Is(err, errUsage) {
				t.Fatalf("run = %v, want an error satisfying errors.Is(err, errUsage)", err)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want usage to go to stderr", stdout)
			}
			if !strings.Contains(stderr, "usage") && !strings.Contains(stderr, "Usage") {
				t.Errorf("stderr = %q, want a usage message", stderr)
			}
		})
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

// parseFields turns "Key: value" lines into a map, failing on any other shape.
func parseFields(t *testing.T, lines []string) map[string]string {
	t.Helper()
	fields := make(map[string]string, len(lines))
	for _, line := range lines {
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			t.Fatalf("line %q is not in %q form", line, "Key: value")
		}
		fields[key] = value
	}
	return fields
}

func parseMeasurement(t *testing.T, s string, pattern string) (float64, bool) {
	t.Helper()
	m := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v, true
}
