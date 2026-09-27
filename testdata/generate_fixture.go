//go:build ignore

// Command generate_fixture writes the synthetic end-to-end regression fixture
// used by the Phase 0 test suite.
//
// The track is not a real recording. It is a deterministic closed loop with
// enough points that naive float64 accumulation of per-edge distances drifts
// measurably away from a compensated reference sum, which is exactly the
// property the fixture exists to exercise.
//
// Regenerate with:
//
//	go run testdata/generate_fixture.go
package main

import (
	"fmt"
	"math"
	"os"
	"time"
)

const (
	pointCount   = 2500
	segmentSplit = 1500
	outPath      = "testdata/realistic.gpx"
)

var epoch = time.Date(2024, time.July, 14, 6, 0, 0, 0, time.UTC)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprint(f, `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="summit-fixture-generator"
     xmlns="http://www.topografix.com/GPX/1/1">
  <wpt lat="45.9000000" lon="6.8500000">
    <ele>1000.0</ele>
    <name>Synthetic Trailhead</name>
  </wpt>
  <wpt lat="45.9500000" lon="6.9200000">
    <ele>1044.0</ele>
    <name>Synthetic Junction</name>
  </wpt>
  <trk>
    <name>Synthetic Ridge Loop</name>
    <trkseg>
`)

	pause := 12 * time.Minute
	var elapsed time.Duration

	for i := range pointCount {
		if i == segmentSplit {
			fmt.Fprint(f, "    </trkseg>\n    <trkseg>\n")
		}
		if i == 1180 {
			elapsed += pause
		}

		th := 2 * math.Pi * float64(i) / float64(pointCount)
		lat := 45.90 + 0.05*math.Cos(th) + 0.004*math.Sin(3*th)
		lon := 6.85 + 0.07*math.Sin(th) + 0.003*math.Cos(2*th)
		ele := 1000 + 250*math.Sin(2*th) + 60*math.Cos(7*th) + 0.02*float64(i)

		ts := epoch.Add(elapsed).Format(time.RFC3339)
		elapsed += 8 * time.Second

		if i%997 == 0 {
			fmt.Fprintf(f, "      <trkpt lat=\"%.7f\" lon=\"%.7f\">\n", lat, lon)
			fmt.Fprintf(f, "        <time>%s</time>\n", ts)
			fmt.Fprint(f, "      </trkpt>\n")
			continue
		}

		fmt.Fprintf(f, "      <trkpt lat=\"%.7f\" lon=\"%.7f\">\n", lat, lon)
		fmt.Fprintf(f, "        <ele>%.1f</ele>\n", ele)
		fmt.Fprintf(f, "        <time>%s</time>\n", ts)
		fmt.Fprint(f, "      </trkpt>\n")
	}

	fmt.Fprint(f, `    </trkseg>
  </trk>
</gpx>
`)

	info, err := f.Stat()
	if err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d points, %d bytes)\n", outPath, pointCount, info.Size())
	return nil
}
