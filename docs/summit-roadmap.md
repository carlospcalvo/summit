# Summit — Go + Wails Learning Roadmap

Summit is a local desktop app for exploring GPS tracks and topographic data.

The project should stay focused on one core idea:

> A fast desktop tool for understanding GPS tracks and terrain.

The roadmap is intentionally ordered so that each product feature introduces a useful Go or Wails concept. Keep GIS/domain logic in Go and use React/TypeScript mainly for presentation and interaction.

---

## Architecture rule

### React / TypeScript
- Rendering
- MapLibre
- Charts
- Forms and controls
- Selection and hover state
- Purely visual transformations

### Go
- Filesystem access
- GPX / GeoJSON parsing
- Geospatial calculations
- Track analysis
- Elevation processing
- DEM handling
- Spatial algorithms
- Persistence
- Import / export
- Caching
- Concurrency

If logic could reasonably live in either TypeScript or Go, prefer Go unless it is purely UI behavior.

---

## Phase 0 — Go refresher with a real CLI

Before Wails, create a tiny CLI in the same repository.

Example:

```bash
summit inspect my-track.gpx
```

Expected output:

```text
Track: Base Sur Lanín
Points: 13,482
Distance: 21.42 km
Elevation gain: 1,147 m
Elevation loss: 1,153 m
Minimum: 812 m
Maximum: 1,694 m
```

Learn:
- Packages
- Structs
- Methods
- Slices
- Pointers
- Errors
- `defer`
- `io.Reader`
- `encoding/xml`
- `os`
- Tests
- Table-driven tests

Suggested models:

```go
type TrackPoint struct {
    Lat       float64
    Lon       float64
    Elevation float64
    Time      time.Time
}

type Track struct {
    Name   string
    Points []TrackPoint
}
```

Implement yourself:

```go
func Distance(a, b TrackPoint) float64
func TotalDistance(t Track) float64
func ElevationGain(t Track) float64
func ElevationLoss(t Track) float64
func Bounds(t Track) Bounds
```

Do not use a GIS library yet.

---

## Phase 1 — First Wails application

Create the desktop shell.

Keep the UI intentionally simple.

```text
┌─────────────────────────────────────────────┐
│ Summit                                      │
├─────────────────────────────────────────────┤
│                                             │
│            [ Open GPX File ]                │
│                                             │
└─────────────────────────────────────────────┘
```

Flow:

```text
React
  ↓
Wails binding
  ↓
Go file/service layer
  ↓
GPX parser
  ↓
TrackAnalysis
  ↓
React
```

Suggested Go API:

```go
type App struct {
    analyzer *track.Analyzer
}

func (a *App) OpenTrack() (*TrackAnalysis, error)
```

Learn:
- Wails lifecycle
- Exported Go methods
- Generated TypeScript bindings
- Frontend ↔ backend calls
- Returning Go structs to JavaScript
- Application/service organization

Goal: open a GPX file and display its statistics.

---

## Phase 2 — Proper Go project structure

Refactor once the first vertical slice works.

Suggested structure:

```text
summit/
├── cmd/
│   └── summit/
├── internal/
│   ├── gpx/
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── geo/
│   │   ├── distance.go
│   │   ├── bounds.go
│   │   └── elevation.go
│   ├── track/
│   │   ├── analysis.go
│   │   └── segments.go
│   └── app/
│       └── service.go
├── frontend/
└── main.go
```

Learn:
- Small package design
- Composition
- Keeping framework code at the edges
- Avoiding unnecessary class-like abstractions

---

## Phase 3 — Elevation profile

Still no map.

Build an elevation profile:

```text
1700m               /\
                  __/  \_
1400m          ___/      \
             _/           \
1100m   _____/             \____
       0    5    10    15    21 km
```

Go produces profile data:

```go
type ProfilePoint struct {
    Distance  float64
    Elevation float64
}

func ElevationProfile(track Track) []ProfilePoint
```

React renders it.

Add hover interaction:
- Cursor distance
- Elevation at that point
- Selected track-point index

Learn:
- Cumulative calculations
- Normalization
- Filtering GPS noise
- Test fixtures

---

## Phase 4 — Add MapLibre

Now add the map.

The map is a visualization, not the core engine.

Initial responsibilities:
- Basemap
- Track polyline
- Fit to track bounds
- Waypoint markers
- Highlight selected track point

Go returns coordinates:

```go
type Coordinate struct {
    Lat float64 `json:"lat"`
    Lon float64 `json:"lon"`
}
```

Important interaction:

```text
hover elevation profile
        ↓
highlight corresponding map point
```

And ideally the reverse as well.

Do not implement your own map renderer.

---

## Phase 5 — Track segmentation

Move beyond being a simple GPX viewer.

Suggested model:

```go
type Segment struct {
    StartDistance float64
    EndDistance   float64
    Distance      float64
    ElevationGain float64
    ElevationLoss float64
    AverageGrade  float64
    MaximumGrade  float64
}
```

Start with fixed-distance segments.

Example UI:

```text
SEGMENTS

0–2 km      +85 m      4.2%
2–5 km      +412 m    13.7%
5–8 km      +251 m     8.4%
8–12 km      -40 m
```

Selecting a segment should:
- Highlight it on the map
- Highlight it on the elevation chart
- Show its statistics

Learn:
- Richer domain modeling
- Slice algorithms
- Sorting
- Filtering
- Using interfaces only where they help

---

## Phase 6 — Local track library

Introduce local persistence.

Suggested library:

```text
Tracks

Patagonia
├── Lanín Base Sur
├── Laguna de los Tres
├── Cajón del Azul
└── Refugio Frey
```

Use SQLite.

Possible entities:
- `tracks`
- `track_points`
- `waypoints`
- `tags`
- `imports`

Prefer storing source data and recomputing cheap derived statistics instead of storing everything.

Learn:
- `database/sql`
- SQLite
- Migrations
- Transactions
- Indexes
- Nullable values
- Persistence boundaries

---

## Phase 7 — Concurrency

Only introduce concurrency when there is a real use case.

Example: importing many GPX files.

```text
Import 50 GPX files
        ↓
worker pool
        ↓
parse
analyze
index
        ↓
progress events
```

Suggested API:

```go
func ImportTracks(
    ctx context.Context,
    paths []string,
    progress chan<- ImportProgress,
) error
```

Learn:
- Goroutines
- Channels
- `context.Context`
- Cancellation
- `sync.WaitGroup`
- Worker pools
- Mutexes when genuinely needed

Show progress in Wails.

---

## Phase 8 — Better GPS analysis

Add more useful track analysis:

- Moving time vs stopped time
- Speed
- Pace
- GPS outlier detection
- Elevation smoothing
- Grade
- Better ascent/descent filtering
- Coordinate interpolation

A key problem to solve:

```text
raw elevation
     ↓
noise filtering
     ↓
smoothed profile
     ↓
gain calculation
```

This is where Summit starts becoming an analysis tool rather than a viewer.

---

## Phase 9 — DEM data

Introduce terrain elevation data.

Instead of trusting GPS altitude:

```text
lat/lon
  ↓
DEM raster
  ↓
terrain elevation
```

Possible responsibilities in Go:
- Load raster/elevation tiles
- Convert coordinate → raster location
- Interpolate elevation
- Cache terrain tiles
- Correct or replace GPX elevation

Learn:
- Raster data
- Binary/file processing
- Interpolation
- Caching
- Large-file handling

---

## Phase 10 — Terrain profile

Allow users to choose two points on the map and generate a terrain cross-section without requiring a GPX route.

Example:

```text
A                                      B
│
│                      /\
│              /\     /  \
│        _____/  \___/    \____
│_______/
└──────────────────────────────────────
                 14.2 km
```

This should reuse the DEM infrastructure from Phase 9.

---

## Phase 11 — Slope visualization

Add slope analysis.

Possible buckets:

```text
< 5%
5–10%
10–15%
15–20%
> 20%
```

Go calculates slope/grade for track sections.

React/MapLibre only renders the visual styling.

Suggested model:

```go
type TrackSection struct {
    Start Coordinate
    End   Coordinate
    Grade float64
}
```

---

## Phase 12 — Waypoints

Add proper waypoint support.

Suggested types:
- Camp
- Water
- Summit
- Pass
- Parking
- Shelter
- Danger
- Custom

Waypoint data may include:
- Coordinates
- Elevation
- Notes
- Type/icon
- Track association

Keep geography central; avoid turning Summit into a general notes app.

---

## Phase 13 — GeoJSON

Add GeoJSON as the second major format.

Do not immediately add every GIS format.

Introduce a more general feature model:

```go
type Feature struct {
    Geometry   Geometry
    Properties map[string]any
}
```

This is the point where Summit begins moving from GPX-centric code toward more general GIS concepts.

---

## Phase 14 — Offline maps

Give the desktop architecture a real advantage.

Example:

```text
Offline region
Patagonia Norte

Zoom 8–15
2.4 GB

[ Download ]
```

Go handles:
- Downloads
- Concurrency
- Retry
- Disk cache
- Eviction
- Offline-region management
- Progress

Later, consider MBTiles:

```text
SQLite
└── z / x / y → tile
```

---

## Phase 15 — Export

Support useful exports.

Potential formats:
- GPX
- GeoJSON
- CSV statistics
- PNG snapshot later

Suggested API:

```go
func ExportTrack(
    track Track,
    format ExportFormat,
    target string,
) error
```

This is a good place to introduce a justified interface:

```go
type Exporter interface {
    Export(io.Writer, Track) error
}
```

---

## Phase 16 — Desktop polish

Once the core application is useful, add desktop-specific polish:

- `.gpx` file associations
- Drag and drop
- Recent files
- macOS menus
- Keyboard shortcuts
- Window state
- App icon
- Signing and notarization
- Packaging
- Update mechanism
- Platform-specific filesystem behavior

This phase teaches the parts of desktop development that web work usually avoids.

---

# Go learning progression

The roadmap intentionally follows this progression:

```text
Basic Go
   ↓
structs / slices / errors
   ↓
testing
   ↓
package design
   ↓
file I/O
   ↓
Wails boundaries
   ↓
SQL
   ↓
goroutines / channels / context
   ↓
binary/raster processing
   ↓
geospatial algorithms
   ↓
interfaces where justified
   ↓
desktop / OS integration
```

---

# Explicit non-goals for now

Avoid these until the core app is genuinely useful:

- Accounts
- Authentication
- Cloud sync
- Collaboration
- Server backend
- Microservices
- Subscription system
- Custom map engine
- 3D terrain
- Mobile app

---

# Immediate milestone

Focus only on the first three phases:

```text
GPX CLI parser
      ↓
Wails file opening
      ↓
Track statistics + elevation profile
```

Once those work, Summit is already a real Go desktop application rather than another language-learning exercise.
