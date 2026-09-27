# Summit — Phase 0–8 Implementation Checklist

This is the working checklist for building Summit from an empty repository into a useful local GPS track analysis application. Complete the phases in order. Do not start the next phase until the current phase gate passes.

The high-level product direction remains in [`summit-roadmap.md`](./summit-roadmap.md). This file is deliberately more prescriptive: it records the chosen stack, task order, expected behavior, tests, manual checks, and suggested commit boundaries.

## Progress

| Phase | Outcome                                                            | Status      | Completed |
| ----- | ------------------------------------------------------------------ | ----------- | --------- |
| 0     | GPX inspection CLI and tested Go domain foundation                 | Not started | —         |
| 1     | Wails app opens one GPX and displays statistics                    | Not started | —         |
| 2     | Clear package and dependency boundaries                            | Not started | —         |
| 3     | Interactive elevation profile                                      | Not started | —         |
| 4     | MapLibre map synchronized with the profile                         | Not started | —         |
| 5     | Fixed-distance track segmentation                                  | Not started | —         |
| 6     | Persistent local SQLite track library                              | Not started | —         |
| 7     | Concurrent bulk import with progress and cancellation              | Not started | —         |
| 8     | Hiking-focused time, speed, grade, smoothing, and outlier analysis | Not started | —         |

Update the status to `In progress`, `Blocked`, or `Complete` as work advances. Add the completion date only after the phase gate passes.

---

## Decisions already made

- [ ] Use the Go module path `github.com/carlospcalvo/summit`.
- [ ] Use Wails v3 beta and pin the resolved version in `go.mod` and `go.sum`.
- [ ] Use the Wails React + TypeScript template, Vite, pnpm, and `pnpm-lock.yaml`.
- [ ] Develop and accept on Apple Silicon macOS first; keep ordinary application code portable.
- [ ] Use metric units only through Phase 8: metres, kilometres, km/h, and min/km.
- [ ] Tune Phase 8 defaults for hiking and trail running.
- [ ] Keep filesystem, parsing, calculations, persistence, and concurrency in Go.
- [ ] Keep rendering, pointer/keyboard interaction, and shared view-selection state in React.
- [ ] Use Mantine as the sole application component, layout, form, feedback, and theming system.
- [ ] Use Zustand for shared track/map/chart interaction state; keep isolated component state in React and substantial form state in Mantine Form.
- [ ] Limit hand-written CSS to MapLibre's required stylesheet/container rules and specialized SVG chart behavior.
- [ ] Render the elevation chart as responsive SVG rather than adding a chart dependency.
- [ ] Use MapLibre GL JS v6 for the map.
- [ ] Use `database/sql` with `modernc.org/sqlite` for the local library.
- [ ] Copy imported GPX files into Summit-managed application data storage.
- [ ] Treat generated Wails bindings as generated files: regenerate them, never edit them.
- [ ] Defer accounts, cloud sync, tags/folders, additional GIS formats, offline maps, DEM data, packaging, and signing.

## Reference documentation

- [Wails v3 installation](https://v3.wails.io/getting-started/installation/)
- [Wails v3 CLI](https://v3.wails.io/reference/cli/)
- [Wails v3 method bindings](https://v3.wails.io/features/bindings/methods/)
- [Wails v3 data models](https://v3.wails.io/features/bindings/models/)
- [Wails v3 file dialogs](https://v3.wails.io/features/dialogs/file/)
- [Wails v3 events](https://v3.wails.io/reference/events/)
- [MapLibre GL JS](https://maplibre.org/maplibre-gl-js/docs/)
- [Mantine](https://mantine.dev/)
- [Mantine AppShell](https://mantine.dev/core/app-shell/)
- [Mantine Form](https://mantine.dev/form/package/)
- [Zustand](https://zustand.docs.pmnd.rs/)
- [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite)

Wails v3 is beta software. If an example here differs from the pinned version, inspect `wails3 <command> --help` and the documentation for that exact version. Adapt the call site without changing the architecture or phase outcome.

---

## Before Phase 0

### Toolchain

- [ ] Confirm the local architecture and macOS version:

  ```bash
  uname -sm
  sw_vers
  ```

- [ ] Confirm Go 1.25 or newer:

  ```bash
  go version
  ```

- [ ] Confirm Node and pnpm:

  ```bash
  node --version
  pnpm --version
  ```

- [ ] Confirm Xcode command-line tools:

  ```bash
  xcode-select -p
  ```

- [ ] Install the Wails v3 CLI if `wails3` is unavailable:

  ```bash
  go install github.com/wailsapp/wails/v3/cmd/wails3@latest
  ```

- [ ] Ensure Go's binary directory is on `PATH`.
- [ ] Run `wails3 doctor` and resolve required failures before scaffolding.
- [ ] Record the outputs of `go version`, `node --version`, `pnpm --version`, and `wails3 version` in the first setup commit or project notes.

### Repository safety

- [ ] Confirm `summit-roadmap.md` and this checklist are present before scaffolding.
- [ ] Confirm Git sees only expected work:

  ```bash
  git status --short
  ```

- [ ] If this repository does not yet have a first commit, commit both roadmap files before generating the application scaffold.
- [ ] Generate the Wails project in a temporary directory because this repository is already nonempty.
- [ ] Copy the generated project contents into the repository without overwriting either roadmap document.
- [ ] Use these scaffold values:
  - Name: `summit`
  - Template: `react`
  - Module: `github.com/carlospcalvo/summit`
- [ ] Run `wails3 init --help` first and use the exact flags supported by the pinned CLI. The intended command shape is:

  ```bash
  wails3 init -n summit -t react -d <temporary-directory>/summit -mod github.com/carlospcalvo/summit
  ```

- [ ] Copy the generated files into this repository, remove only the temporary scaffold directory, and inspect the diff.
- [ ] Confirm `go.mod` contains the intended module path.
- [ ] Run `pnpm install` in the generated frontend directory if the template did not already do so.
- [ ] Commit `go.mod`, `go.sum`, `package.json`, and `pnpm-lock.yaml`; do not leave dependency versions floating.
- [ ] Start the unmodified scaffold with `wails3 dev` and verify that a desktop window opens.
- [ ] Build it once with `wails3 build`.

Suggested commit:

```text
chore: bootstrap summit project
```

---

# Phase 0 — GPX inspection CLI and Go foundation

## Outcome

The command below parses a GPX file and prints trustworthy statistics without using Wails or a GIS library:

```bash
go run ./cmd/summit inspect testdata/example.gpx
```

## Domain decisions

- [ ] Model a track as ordered points plus waypoints.
- [ ] Give every track point a source GPX segment index so calculations can avoid bridging separate `<trkseg>` elements.
- [ ] Represent missing elevation and missing timestamps explicitly. Do not use numeric zero or `time.Time{}` as an implicit missing marker.
- [ ] Keep coordinates in the domain as latitude then longitude; convert to GeoJSON longitude/latitude only at the frontend map boundary.
- [ ] Use metres internally for distance/elevation and `time.Duration` for durations.
- [ ] Preserve input order and source segment boundaries.
- [ ] Do not silently repair malformed coordinates or timestamps in Phase 0.

An appropriate initial model is:

```go
type TrackPoint struct {
    Lat          float64
    Lon          float64
    Elevation    *float64
    Time         *time.Time
    SourceSegment int
}

type Waypoint struct {
    Name      string
    Lat       float64
    Lon       float64
    Elevation *float64
    Time      *time.Time
}

type Track struct {
    Name      string
    Points    []TrackPoint
    Waypoints []Waypoint
}
```

Names may be adjusted to match Go style, but the missing-data and segment-boundary semantics must remain.

## GPX parser

- [ ] Create `internal/gpx`.
- [ ] Implement `Parse(io.Reader) (track.Track, error)` with `encoding/xml.Decoder`.
- [ ] Read through `io.Reader`; do not require a path or load the entire file before parsing.
- [ ] Parse GPX 1.0 and 1.1 tracks using local XML element names rather than hard-coding one namespace URI.
- [ ] Parse the first track name when present.
- [ ] Parse all `<trkseg>` elements and assign stable zero-based segment indices.
- [ ] Parse `lat`, `lon`, optional `<ele>`, and optional `<time>` for every track point.
- [ ] Parse waypoint name, coordinates, optional elevation, and optional time.
- [ ] Accept an absent name and let the calling layer derive a display name from the source filename.
- [ ] Reject latitude outside `[-90, 90]` and longitude outside `[-180, 180]`.
- [ ] Reject missing or nonnumeric track-point coordinate attributes.
- [ ] Return contextual errors that identify the operation and approximate offending element.
- [ ] Preserve the underlying error with `%w` where callers may need `errors.Is` or `errors.As`.
- [ ] Make an empty but valid GPX document return a valid empty track; let analysis decide which statistics are available.
- [ ] Never panic on malformed XML or missing optional fields.

## Geospatial calculations

- [ ] Create a minimal `internal/geo` package.
- [ ] Implement Haversine point-to-point distance in metres.
- [ ] Clamp floating-point input to inverse trigonometric functions when rounding could place it slightly outside the valid range.
- [ ] Implement total distance while skipping edges that cross source GPX segment boundaries.
- [ ] Implement raw elevation gain and loss from consecutive points with elevation, again skipping source-segment boundaries.
- [ ] Implement minimum and maximum elevation with explicit unavailable results when no point has elevation.
- [ ] Implement bounds for empty, one-point, ordinary, and antimeridian-crossing tracks.
- [ ] Document the bounds representation and how it distinguishes an unavailable bound from a zero coordinate.
- [ ] Keep all results finite; never return `NaN` or infinity in application-facing data.

## CLI

- [ ] Create `cmd/summit/main.go`.
- [ ] Support exactly `summit inspect <path.gpx>` in this phase.
- [ ] Open the file in the CLI layer and `defer` its close immediately after confirming the open succeeded.
- [ ] Pass the open file to the GPX parser as an `io.Reader`.
- [ ] Derive the track name from the filename when GPX metadata has no name.
- [ ] Print:
  - Track name
  - Point count
  - Distance in kilometres with two decimals
  - Elevation gain/loss in whole metres when available
  - Minimum/maximum elevation in whole metres when available
- [ ] Print `n/a` for unavailable elevation statistics.
- [ ] Send human-readable errors to stderr.
- [ ] Return distinct nonzero exit behavior for invalid usage and processing failure. It is enough to centralize command execution in a `run(args, stdout, stderr) error`-style function and map errors at `main`.
- [ ] Keep formatting outside the core analysis functions.

## Fixtures and tests

Create small, readable fixtures under a shared `testdata` directory:

- [ ] `single-point.gpx`
- [ ] `flat-track.gpx`
- [ ] `multi-segment.gpx`
- [ ] `missing-elevation.gpx`
- [ ] `missing-time.gpx`
- [ ] `with-waypoints.gpx`
- [ ] `antimeridian.gpx`
- [ ] `malformed.gpx`
- [ ] `invalid-coordinate.gpx`
- [ ] One anonymized realistic track with enough points to expose cumulative rounding errors

Tests:

- [ ] Table-test known distances, including identical points and a known city-to-city pair.
- [ ] Test that total distance does not connect separate GPX segments.
- [ ] Test empty and single-point tracks.
- [ ] Test flat, climbing, descending, and mixed elevation sequences.
- [ ] Test missing elevations between valid samples without treating missing as zero.
- [ ] Test ordinary and antimeridian bounds.
- [ ] Test GPX 1.0 and GPX 1.1 namespaces.
- [ ] Test multiple tracks or explicitly document and test the chosen first-track behavior.
- [ ] Test malformed XML, invalid coordinate ranges, missing attributes, bad numbers, and bad timestamps.
- [ ] Test CLI usage, output formatting, fallback naming, and error output with in-memory buffers.
- [ ] Compare floating-point values with stated tolerances rather than exact equality.
- [ ] Add a parser/analysis benchmark using a generated large track; do not commit a huge fixture solely for the benchmark.

## Learning checks

- [ ] Explain why the parser accepts `io.Reader` rather than a filename.
- [ ] Explain where pointers are useful for optional GPX values and where they are unnecessary.
- [ ] Explain why segment boundaries affect distance and elevation calculations.
- [ ] Explain error wrapping and identify at least one place `errors.Is` is useful.
- [ ] Explain why table-driven tests suit coordinate and parser cases.

## Phase 0 gate

- [ ] `gofmt` reports no required formatting changes.
- [ ] `go vet ./...` passes.
- [ ] `go test ./...` passes.
- [ ] The inspect command produces the expected summary for the realistic fixture.
- [ ] Malformed input produces a contextual error and no panic.
- [ ] Missing elevation produces `n/a`, not misleading zeroes.
- [ ] Multiple GPX segments do not gain a synthetic connecting edge.
- [ ] The Wails scaffold still starts after adding the CLI.

Suggested commits:

```text
feat: parse GPX tracks
feat: add geospatial track statistics
feat: add track inspection CLI
```

---

# Phase 1 — First Wails vertical slice

## Outcome

The desktop application opens one GPX file through a native dialog and displays its summary. Cancelling the dialog leaves the current screen unchanged.

## Go application boundary

- [ ] Create a Wails-facing `TrackService` rather than adding domain behavior to the bootstrap `main` package.
- [ ] Register it with Wails v3 through `application.NewService`.
- [ ] Inject the parser/analyzer dependency into the service.
- [ ] Give the service access to Wails' application/dialog manager without importing Wails into domain packages.
- [ ] Implement an exported binding with the semantics:

  ```go
  func (s *TrackService) OpenTrack() (*TrackDetails, error)
  ```

- [ ] Open a native single-file dialog titled `Open GPX Track`.
- [ ] Filter for `*.gpx` and allow the platform's normal all-files fallback only if the pinned Wails API requires it.
- [ ] Treat an empty selected path as cancellation and return `nil, nil`.
- [ ] Parse and analyze the selected file in Go.
- [ ] Do not return the absolute file path to React.
- [ ] Add operation context to errors: opening, parsing, or analyzing.

## Boundary DTOs

- [ ] Create JSON-tagged application DTOs separate from domain models.
- [ ] Define `TrackDetails` with a stable identity for the current session, display name, point count, summary, bounds, and capability flags.
- [ ] Define summary fields in base units: metres and optional values where data may be unavailable.
- [ ] Include `hasElevation` and `hasTime` flags.
- [ ] Ensure no unavailable Go float serializes as `NaN` or infinity.
- [ ] Keep raw track points private until Phase 3/4 requires them.
- [ ] Add one explicit mapper from domain analysis to Wails DTOs.

## Bindings

- [ ] Run `wails3 generate bindings` using the flags expected by the generated React template.
- [ ] Confirm generated TypeScript types reflect JSON nullability.
- [ ] Import the generated `OpenTrack` binding rather than using an untyped runtime call.
- [ ] Add a documented package script or Taskfile task for regenerating bindings.
- [ ] Add a check that regeneration leaves no unintended diff.

## React UI

- [ ] Remove the Wails greeting/demo UI.
- [ ] Install and pin `@mantine/core`, `@mantine/hooks`, `@mantine/form`, `@mantine/notifications`, `@tabler/icons-react`, and `zustand`.
- [ ] Import Mantine's required package styles once at the frontend entry point and in the documented order.
- [ ] Create one centralized Summit theme with compact desktop spacing, accessible focus behavior, light/dark color schemes, and component defaults.
- [ ] Wrap the application once with `MantineProvider` and render one `Notifications` host inside it.
- [ ] Build the shell with `AppShell`, including application title and main content area.
- [ ] Add an `Open GPX File` Mantine button with a Tabler icon.
- [ ] Disable duplicate opens while a request is pending.
- [ ] Show a Mantine loading indicator or skeleton with text that remains understandable without animation.
- [ ] Preserve the previous result if the user cancels the file dialog.
- [ ] Show a Mantine-composed empty state before the first successful open.
- [ ] Show track name, points, distance, gain, loss, minimum, and maximum with Mantine typography, groups, and cards.
- [ ] Render unavailable elevation values as `—` or `n/a` consistently.
- [ ] Format base-unit values in TypeScript only for presentation; do not recalculate statistics.
- [ ] Catch rejected binding promises and render a Mantine alert/notification with a retry/open action.
- [ ] Make the button and statistics usable with keyboard navigation and screen readers.
- [ ] Use Mantine component props, spacing tokens, and theme values for ordinary appearance; do not create page-specific CSS for controls or layout.

## Tests

- [ ] Add Vitest and React Testing Library if the Wails template does not include them.
- [ ] Add a TypeScript type-check script that does not emit files.
- [ ] Add a shared test renderer that wraps components with `MantineProvider` and any required notification portal setup.
- [ ] Mock only the generated service boundary in component tests.
- [ ] Test the initial empty state.
- [ ] Test loading and disabled-button behavior.
- [ ] Test successful summary rendering.
- [ ] Test unavailable optional statistics.
- [ ] Test cancellation (`null`) without clearing an existing result.
- [ ] Test a rejected backend promise and retry.
- [ ] Test the Go DTO mapper independently from the Wails dialog.
- [ ] Keep the native dialog and actual Wails bridge as manual integration checks.

## Learning checks

- [ ] Identify the Wails application lifecycle points used by the generated project.
- [ ] Explain how service method bindings are generated.
- [ ] Explain why Go errors become rejected frontend promises.
- [ ] Explain why boundary DTOs are separate from domain and database types.

## Phase 1 gate

- [ ] Phase 0 gate still passes.
- [ ] Generated bindings are current.
- [ ] Go tests and `go vet ./...` pass.
- [ ] Frontend tests and type checking pass.
- [ ] The production frontend build passes.
- [ ] `wails3 dev` opens the app and successfully loads a GPX.
- [ ] Cancelling the native picker is a no-op, not an error.
- [ ] A malformed GPX produces a visible, recoverable error.
- [ ] `wails3 build` completes on macOS.

Suggested commits:

```text
feat: add Wails track service
feat: display GPX statistics
test: cover desktop open-track flow
```

---

# Phase 2 — Package and dependency refactor

## Outcome

The CLI and desktop application behave exactly as before, but the code has explicit framework, application, parsing, and analysis boundaries.

## Target structure

```text
cmd/
  summit/
internal/
  app/
  geo/
  gpx/
  track/
frontend/
main.go
```

The generated Wails project may add build/config directories. Preserve its expected layout.

## Refactor tasks

- [ ] Capture Phase 1 CLI and UI behavior before moving code.
- [ ] Move GPX XML structs and parsing into `internal/gpx`.
- [ ] Keep coordinate primitives and reusable distance/bounds functions in `internal/geo`.
- [ ] Move track-domain models and analysis orchestration into `internal/track`.
- [ ] Move Wails services, DTOs, and domain-to-DTO mappers into `internal/app`.
- [ ] Keep `main.go` limited to dependency construction, Wails configuration, service registration, and application startup.
- [ ] Keep `cmd/summit` limited to CLI argument handling, filesystem opening, output formatting, and exit behavior.
- [ ] Make CLI and Wails call the same concrete `track.Analyzer`.
- [ ] Prefer a concrete analyzer type; do not create an interface only to imitate classes.
- [ ] Introduce interfaces at I/O boundaries only when they enable a real alternate implementation or focused test.
- [ ] Prevent imports from `internal/track`, `internal/geo`, or `internal/gpx` into Wails packages.
- [ ] Prevent domain packages from importing database, browser, React, or generated-binding concepts.
- [ ] Keep DTO mapping in one direction and one package.
- [ ] Add short package comments that explain ownership rather than restating package names.
- [ ] Update imports and regenerate Wails bindings once, after public Go types stabilize.

## Dependency rule

Document and verify this direction:

```text
CLI ───────┐
           ├──> application/analyzer ───> track ───> geo
Wails ─────┘                  │
                              └──────────> gpx

React ──generated bindings──> Wails service DTOs
```

- [ ] Search imports to confirm lower layers do not depend on upper layers.
- [ ] Use `go list -deps ./...` when an unexpected dependency appears.
- [ ] Do not add a dependency-injection framework.
- [ ] Do not add generic repositories, factories, or managers without a current second implementation.

## Regression tests

- [ ] Move tests with their packages and keep fixtures discoverable through relative `testdata` paths.
- [ ] Preserve every Phase 0/1 test during moves.
- [ ] Add a CLI-versus-Wails mapper consistency test for shared summary values.
- [ ] Verify the production binary and CLI both compile.
- [ ] Compare representative CLI output before and after the refactor.

## Learning checks

- [ ] Explain why Wails belongs at an edge.
- [ ] Point to one place composition is used instead of inheritance-like embedding.
- [ ] Identify any interface introduced and state the concrete need that justifies it.
- [ ] Explain why package count should follow responsibilities rather than files.

## Phase 2 gate

- [ ] There is no intended user-visible behavior change.
- [ ] All Phase 1 gate checks still pass.
- [ ] Both entry points use the same analyzer.
- [ ] No geospatial formula is duplicated in TypeScript or Wails service code.
- [ ] Domain packages have no framework or persistence imports.
- [ ] Generated bindings are current after moved public DTOs.

Suggested commits:

```text
refactor: separate track domain packages
refactor: share analyzer across CLI and desktop
```

---

# Phase 3 — Interactive elevation profile

## Outcome

The loaded track displays a responsive elevation profile. Pointer and keyboard interaction identify a stable source track-point index.

## Go profile generation

- [ ] Define a profile DTO/domain value with cumulative distance, elevation availability, elevation value, and original point index.
- [ ] Add source segment index if the frontend needs to render visible gaps.
- [ ] Implement cumulative distance once in Go and reuse it for total distance, profiles, and later segmentation.
- [ ] Do not add distance across source GPX segment boundaries.
- [ ] Keep points with missing elevation in the profile sequence but mark their value unavailable.
- [ ] Do not interpolate missing elevations yet.
- [ ] Return profile data with `TrackDetails` or a focused details result rather than making one binding call per point.
- [ ] Confirm the serialized payload remains acceptable for the realistic fixture.
- [ ] Regenerate bindings.

Suggested shape:

```go
type ProfilePoint struct {
    PointIndex    int      `json:"pointIndex"`
    SourceSegment int     `json:"sourceSegment"`
    DistanceM     float64  `json:"distanceM"`
    ElevationM    *float64 `json:"elevationM"`
}
```

## Shared frontend selection state

- [ ] Add a small typed Zustand interaction store with selector-based subscriptions.
- [ ] Store `selectedPointIndex` as the canonical cross-view selection.
- [ ] Distinguish persistent selection from transient hover if both are implemented.
- [ ] Clear selection when a different track is loaded.
- [ ] Keep loaded Go-owned domain data outside the interaction store unless multiple views require the same frontend-owned value.
- [ ] Keep isolated chart-local state in React rather than adding every value to Zustand.
- [ ] Expose focused actions/selectors for point and track changes instead of letting components replace the whole store.

## SVG chart

- [ ] Create a responsive elevation-profile component using SVG.
- [ ] Derive x coordinates from cumulative distance and y coordinates from available elevation extent.
- [ ] Handle flat profiles by adding visual y padding so the domain is nonzero.
- [ ] Break the rendered path at missing elevations and source GPX segment boundaries.
- [ ] Draw labelled distance and elevation axes with metric units.
- [ ] Add an accessible text summary or table-equivalent description.
- [ ] Add a focusable interaction surface.
- [ ] Convert pointer x position into track distance.
- [ ] Binary-search the full profile for the nearest point by distance.
- [ ] Show cursor distance, elevation, and selected source index in a tooltip.
- [ ] Support left/right arrow movement between valid profile points.
- [ ] Keep tooltip content within the visible chart bounds.

## Visual downsampling

- [ ] Keep all profile data for selection and calculations.
- [ ] Downsample only the SVG path when there are significantly more points than horizontal pixels.
- [ ] Bucket by x pixel and retain each bucket's first, minimum, maximum, and last relevant point in source order.
- [ ] Never use downsampled array position as the source point index.
- [ ] Recompute the visual series when chart width changes.
- [ ] Confirm prominent peaks and valleys remain visible.

## Tests

- [ ] Test cumulative distance and source-index preservation.
- [ ] Test profile behavior across GPX segment boundaries.
- [ ] Test missing elevation gaps and flat profiles.
- [ ] Test nearest-point binary search at start, middle, exact point, between points, and end.
- [ ] Test min/max bucket downsampling and retained source ordering.
- [ ] Test pointer coordinate conversion with a mocked SVG bounding rectangle.
- [ ] Test keyboard selection and selection reset after loading another track.
- [ ] Test tooltip formatting for present and absent elevation.

## Learning checks

- [ ] Explain why cumulative distance belongs in Go.
- [ ] Explain why chart pixel reduction is a purely visual TypeScript transformation.
- [ ] Explain how stable source indices allow independent views to synchronize.
- [ ] Identify the complexity of profile generation and nearest-point lookup.

## Phase 3 gate

- [ ] Phase 2 gate still passes.
- [ ] A track with missing elevation renders gaps rather than zero-altitude drops.
- [ ] Pointer hover selects the expected original point.
- [ ] Keyboard navigation works without a mouse.
- [ ] The realistic fixture resizes and hovers without visible lag.
- [ ] No frontend code recalculates total distance, gain, or loss.

Suggested commits:

```text
feat: generate elevation profiles
feat: add interactive elevation chart
```

---

# Phase 4 — MapLibre map

## Outcome

The application renders a basemap, complete track, waypoints, and selected point. Profile interaction updates the map, and map interaction updates the profile.

## Dependency and setup

- [ ] Install and commit MapLibre GL JS v6 and its lockfile changes.
- [ ] Import MapLibre's CSS once at the frontend entry point.
- [ ] Use ESM/namespace or named imports supported by v6; do not use the removed CommonJS/default-import pattern.
- [ ] Configure the worker URL only if the pinned Vite/Wails build requires it.
- [ ] Start with a no-key development style, such as MapLibre demo tiles.
- [ ] Keep the style URL in one frontend configuration constant so it can be replaced later.
- [ ] Treat online basemap access as optional visualization input; track data must remain usable when tiles fail.

## Track coordinate DTOs

- [ ] Extend track details with map coordinates carrying source point index and source segment index.
- [ ] Return waypoint coordinate/name data required for rendering.
- [ ] Keep Go DTO coordinate fields named `lat` and `lon`.
- [ ] Regenerate bindings.
- [ ] Build GeoJSON in TypeScript as a presentation transformation.
- [ ] Convert every coordinate to GeoJSON `[longitude, latitude]` order in one tested helper.
- [ ] Create one LineString feature per source GPX segment; do not connect separate segments.
- [ ] Add stable `pointIndex` properties to point hit-test features.

## Map lifecycle

- [ ] Create the MapLibre instance only after its container exists.
- [ ] Hold the instance in a React ref rather than React state.
- [ ] Add navigation controls appropriate for desktop use.
- [ ] Wait for the style to load before adding Summit sources/layers.
- [ ] Update source data instead of recreating the entire map when selection changes.
- [ ] Remove event handlers and call `map.remove()` when the component unmounts.
- [ ] Guard asynchronous style/load callbacks against an already-unmounted component.

## Layers and fit behavior

- [ ] Add a route line source/layer.
- [ ] Add a waypoint source and circle/symbol layer.
- [ ] Add a selected-point source/layer above the route.
- [ ] Add a transparent or low-opacity point layer for hit testing, with a sensible hit radius.
- [ ] Fit the map only when a new track is loaded, not after every hover.
- [ ] Add padding so the route is not hidden beneath panels.
- [ ] Choose a practical zoom for a one-point or zero-area bound.
- [ ] Use the Go bounds representation to handle antimeridian tracks without zooming to the entire world.
- [ ] Avoid automatically refitting after the user pans or zooms.

## Two-way interaction

- [ ] When `selectedPointIndex` changes, update the selected-point GeoJSON.
- [ ] When the pointer hits a track point layer, dispatch that point's original index.
- [ ] Clear transient map hover when leaving the hit-test layer.
- [ ] Make click selection available for users who cannot maintain hover.
- [ ] Keep map/profile selection state in React; do not call Go on every pointer move.
- [ ] Confirm a point selected from a downsampled chart still highlights the exact map point.

## Error and empty states

- [ ] Render a stable empty map area before a track is loaded.
- [ ] Show a nonblocking message if map style or tiles fail.
- [ ] Keep statistics and elevation profile usable after basemap failure.
- [ ] Report WebGL initialization failure clearly.
- [ ] Avoid retry loops that repeatedly recreate the map.

## Tests

- [ ] Test `[lat, lon]` to GeoJSON `[lon, lat]` conversion.
- [ ] Test one LineString per GPX source segment.
- [ ] Test waypoint feature properties.
- [ ] Test selected-point feature lookup by original index.
- [ ] Test reducer synchronization from chart and map actions.
- [ ] Test track-load behavior resets selection and requests one fit.
- [ ] Use small MapLibre mocks for lifecycle tests; do not reproduce MapLibre internals in tests.
- [ ] Manually verify actual WebGL rendering in the Wails window.

## Learning checks

- [ ] Explain why MapLibre is a renderer rather than Summit's geospatial engine.
- [ ] Explain the longitude/latitude ordering difference.
- [ ] Explain why map lifecycle objects belong in refs.
- [ ] Explain why selection does not need a Wails bridge call.

## Phase 4 gate

- [ ] Phase 3 gate still passes.
- [ ] Opening a GPX fits and displays every source segment.
- [ ] Waypoints appear at the correct coordinates.
- [ ] Chart hover/click highlights the corresponding map point.
- [ ] Map hover/click updates the chart point and tooltip.
- [ ] Repeatedly opening different tracks does not duplicate layers or event listeners.
- [ ] Disconnecting the network degrades the basemap without crashing the app.
- [ ] Production Wails build includes MapLibre assets correctly.

Suggested commits:

```text
feat: add MapLibre track map
feat: synchronize map and elevation profile
```

---

# Phase 5 — Fixed-distance track segmentation

## Outcome

Go divides the route into 2 km analysis segments. Selecting a segment highlights the same range in the segment list, map, and elevation chart.

## Domain model

- [ ] Name this concept `AnalysisSegment` to distinguish it from GPX `<trkseg>` source segments.
- [ ] Use a 2,000 m default length through Phase 5.
- [ ] Keep the segment length as an analyzer option or function argument, even though the UI does not edit it yet.
- [ ] Define:
  - Stable segment index/ID
  - Start/end cumulative distance
  - Start/end source point index
  - Actual distance
  - Elevation gain/loss when available
  - Average grade when available
  - Provisional maximum grade when available
- [ ] Let the final segment be shorter than 2 km.
- [ ] Omit a zero-length trailing segment when total distance is an exact multiple of the segment length.
- [ ] Never create distance across GPX source-segment boundaries.

## Segmentation algorithm

- [ ] Assign points/edges to buckets from cumulative distance.
- [ ] Decide and document which segment owns a point exactly on a boundary; use the later segment except for the route's final endpoint.
- [ ] Preserve enough source indices to extract/highlight the segment geometry.
- [ ] Calculate distance from the same cumulative-distance series used by the profile.
- [ ] Calculate gain/loss from available raw elevation in Phase 5.
- [ ] Calculate average grade as net elevation change divided by horizontal distance, not gain divided by distance.
- [ ] Protect every grade calculation from zero or tiny horizontal distance.
- [ ] Calculate provisional maximum grade only across intervals meeting a documented minimum horizontal-distance guard.
- [ ] Mark elevation statistics unavailable when endpoints/data are insufficient.
- [ ] Document that Phase 8 replaces raw elevation and adjacent-point grade with filtered analysis.

## Application DTOs and state

- [ ] Add analysis segments to the loaded-track response.
- [ ] Regenerate bindings.
- [ ] Add `selectedAnalysisSegmentID` and focused selection actions to the Zustand interaction store.
- [ ] Reset segment selection on track change.
- [ ] Keep persistent segment selection when the pointer briefly hovers a point.
- [ ] Define escape/clear behavior for segment selection.

## Segment list

- [ ] Add a `Segments` panel ordered by route distance.
- [ ] Display start/end distance, distance, gain/loss, average grade, and maximum grade when available.
- [ ] Make each row a real accessible button or selection control.
- [ ] Visually and semantically identify the selected row.
- [ ] Scroll the selected row into view only when selection originated elsewhere and it is offscreen.

## Map and profile highlighting

- [ ] Build one route feature per analysis segment, retaining source GPX breaks where necessary.
- [ ] Add `segmentID` to feature properties.
- [ ] Style the selected segment above the normal route.
- [ ] Allow map segment click to select the segment.
- [ ] Shade or bracket the segment's cumulative-distance range on the elevation chart.
- [ ] Keep point hover visible above segment highlighting.
- [ ] Selecting from list, map, or chart must dispatch the same segment ID.

## Tests

- [ ] Test tracks shorter than 2 km.
- [ ] Test exact 2 km and exact multiple boundaries.
- [ ] Test a short final segment.
- [ ] Test points that skip across a boundary.
- [ ] Test GPX source-segment breaks.
- [ ] Test flat, climbing, descending, and missing-elevation segments.
- [ ] Test zero/tiny horizontal distance does not yield invalid grade.
- [ ] Test selection precedence between point hover and persistent segment selection.
- [ ] Test the selected map feature and profile range share the DTO's start/end distance.

## Learning checks

- [ ] Explain the difference between GPX source segments and analysis segments.
- [ ] Explain why average grade uses net elevation change.
- [ ] Identify where Phase 5's raw-elevation results are intentionally provisional.
- [ ] State the complexity of segmentation.

## Phase 5 gate

- [ ] Phase 4 gate still passes.
- [ ] Segment distances cover the route once, within the stated floating-point tolerance.
- [ ] The final partial segment is correct.
- [ ] No `NaN` or infinity appears for flat or duplicate-coordinate data.
- [ ] Selection from any view highlights the same segment in all views.
- [ ] Point hover and segment selection remain visually distinguishable.

Suggested commits:

```text
feat: calculate fixed-distance segments
feat: add synchronized segment selection
```

---

# Phase 6 — Local SQLite track library

## Outcome

Imported tracks appear in a local library, survive application restarts, and remain available after the user moves or deletes the original GPX file.

## Storage layout

- [ ] Resolve the base directory with `os.UserConfigDir` or the pinned Wails v3 platform-path facility.
- [ ] Use a stable Summit directory beneath that base.
- [ ] Store the database as `summit.db`.
- [ ] Store managed originals under `sources/`.
- [ ] Create directories with user-only write permissions where supported.
- [ ] Keep path construction in Go and test it independently from the real home directory.
- [ ] Never expose managed absolute paths to React.

Expected logical layout:

```text
Summit/
  summit.db
  sources/
    <content-hash>.gpx
```

## SQLite setup

- [ ] Add `modernc.org/sqlite` and commit module changes.
- [ ] Open through `database/sql` with driver name `sqlite`.
- [ ] Enable foreign keys for every connection.
- [ ] Configure a busy timeout.
- [ ] Choose a conservative connection policy suitable for one desktop process and serialized writes.
- [ ] Ping and migrate the database during application startup.
- [ ] Close the database during application shutdown.
- [ ] Surface startup/migration failures clearly rather than silently creating an empty library.

## Migrations

- [ ] Store ordered SQL migrations as embedded files.
- [ ] Record applied migration versions in a dedicated schema table or a consistently managed SQLite user version.
- [ ] Run each migration transactionally where SQLite permits.
- [ ] Refuse to open a database with a schema version newer than the application understands.
- [ ] Make a failed migration leave the prior schema usable.
- [ ] Add a migration test that starts from an empty temporary database.

## Initial schema

Create only tables used through Phase 8:

- [ ] `tracks`
  - Stable text ID generated from cryptographically secure random bytes
  - Name
  - Point count
  - Capability flags
  - Created/imported timestamps
- [ ] `track_points`
  - Track foreign key
  - Stable sequence number
  - Source segment index
  - Latitude/longitude
  - Nullable elevation
  - Nullable timestamp
- [ ] `waypoints`
  - Track foreign key
  - Sequence number
  - Name
  - Coordinates
  - Nullable elevation/time
- [ ] `imports`
  - Track foreign key
  - SHA-256 source hash with uniqueness constraint
  - Original path for local provenance/debugging
  - Managed relative path
  - Import timestamp
  - Parser/schema version
- [ ] Add indexes for ordered point lookup, track foreign keys, hash lookup, and library sort order.
- [ ] Add foreign-key cascade behavior and test it.
- [ ] Do not add tags, folders, or precomputed profile/segment tables yet.

## Managed import transaction

- [ ] Open the selected source read-only.
- [ ] Stream it through SHA-256 calculation; avoid trusting filename or modification time for identity.
- [ ] Check the hash for an existing import before copying/inserting.
- [ ] If it exists, return the existing track with a duplicate result rather than creating another track.
- [ ] Parse and analyze before committing database state.
- [ ] Copy source bytes to a temporary file inside `sources/`.
- [ ] Flush/close the temporary file and atomically rename it to `<hash>.gpx`.
- [ ] Insert track, point, waypoint, and import rows in one database transaction.
- [ ] Roll back database rows on any insert failure.
- [ ] Remove only the just-created temporary/managed file if the database transaction fails and no existing import owns it.
- [ ] Ensure an application crash cannot leave a database record pointing at a partial file.
- [ ] Preserve the original source file unchanged.

## Repository boundary

- [ ] Define the repository interface in the application/domain-owning layer rather than the SQLite package.
- [ ] Include only behavior needed now: create/import metadata, list summaries, load a track, find by hash, and remove a track.
- [ ] Implement SQLite as the concrete repository.
- [ ] Keep SQL rows private to the persistence package.
- [ ] Map persistence rows back into domain `Track` values before analysis.
- [ ] Recompute cheap statistics, profiles, and segments rather than persisting all derived results.
- [ ] Preserve the managed source as the authoritative original input.

## Wails service API

- [ ] Add `ImportTrack() (*ImportResult, error)` using a native single-file picker.
- [ ] Add `ListTracks() ([]TrackListItem, error)`.
- [ ] Add `GetTrack(id string) (*TrackDetails, error)`.
- [ ] Add `RemoveTrack(id string) error`.
- [ ] Validate IDs at the service/application boundary.
- [ ] Return a duplicate status and existing ID from imports.
- [ ] Make removal an explicit UI action with confirmation.
- [ ] Remove database data and its managed source coherently; report partial filesystem cleanup without resurrecting deleted rows.
- [ ] Regenerate bindings.

## Library UI

- [ ] Add a persistent library sidebar or panel.
- [ ] Load the track list at application startup.
- [ ] Add single-file import.
- [ ] Show library loading, empty, duplicate, and error states.
- [ ] Sort tracks predictably by most recently imported, then name.
- [ ] Select a library item to load its full details into the existing workspace.
- [ ] Preserve map/profile behavior for database-loaded tracks.
- [ ] Add explicit removal with track name in the confirmation.
- [ ] Keep the selected track stable when refreshing the list if it still exists.
- [ ] Select a sensible neighboring item or empty state after removal.

## Tests

- [ ] Use a fresh temporary directory/database per persistence test.
- [ ] Test empty-database migration.
- [ ] Test reopening an already migrated database.
- [ ] Test unsupported future schema version.
- [ ] Test track/point/waypoint round trips including nullable values and source segments.
- [ ] Test transaction rollback on an injected insert failure.
- [ ] Test duplicate hash behavior.
- [ ] Test two different files with the same filename.
- [ ] Test missing/corrupted managed source behavior.
- [ ] Test foreign-key cascade behavior.
- [ ] Test repository list ordering.
- [ ] Test persistence after closing and reopening the database.
- [ ] Test service DTO mapping from a stored track.
- [ ] Test library UI loading, selection, duplicate messaging, removal confirmation, and errors.

## Learning checks

- [ ] Explain why the original source is managed outside the database.
- [ ] Explain what the import transaction protects and what the filesystem rename protects.
- [ ] Explain the purpose of each index.
- [ ] Explain nullable SQL values versus optional domain values.
- [ ] Explain why derived statistics are recomputed.

## Phase 6 gate

- [ ] Phase 5 gate still passes for newly opened and stored tracks.
- [ ] Import a track, quit Summit, reopen, and load it.
- [ ] Move or delete the original file and confirm the library copy still loads.
- [ ] Import the same bytes from another path and confirm no duplicate track is created.
- [ ] A failed import leaves neither partial rows nor a partial managed file.
- [ ] Removing a track removes its rows and managed source after confirmation.
- [ ] All migration/repository tests pass against temporary storage.

Suggested commits:

```text
feat: add SQLite track repository
feat: manage imported GPX sources
feat: add local track library UI
```

---

# Phase 7 — Concurrent bulk import

## Outcome

Users can select many GPX files, see responsive per-job progress, cancel the job, and retain every fully completed import. One bad file does not discard unrelated successes.

## Internal API

- [ ] Keep the core importer usable without Wails:

  ```go
  func (i *Importer) ImportTracks(
      ctx context.Context,
      paths []string,
      progress chan<- ImportProgress,
  ) error
  ```

- [ ] Make progress delivery respect cancellation and avoid blocking workers forever when the receiver exits.
- [ ] Close the progress channel from the owning orchestration goroutine, never from workers.
- [ ] Define whether the returned error represents job-level failure; per-file failures belong in results/progress and should not stop unrelated files.

## Job and progress model

- [ ] Give each job a cryptographically random stable ID.
- [ ] Define file states: queued, parsing, storing, completed, duplicate, failed, and cancelled.
- [ ] Define `ImportProgress` with:
  - Job ID
  - Total files
  - Completed/duplicate/failed/cancelled counts
  - Current filename without an unnecessary absolute path
  - Current file state
  - Optional sanitized error message
  - Job completion flag
- [ ] Ensure progress counters are monotonic and never exceed total files.
- [ ] Emit one terminal update even when cancelled.

## Worker pool

- [ ] Default worker count to `min(runtime.NumCPU(), 4)`, with a minimum of one.
- [ ] Cap workers at the number of input files.
- [ ] Feed paths through a jobs channel.
- [ ] Hash, parse, and analyze separate files concurrently.
- [ ] Keep SQLite writes safe through the repository's transaction/connection policy.
- [ ] Check `ctx.Err()` before expensive work and between hash, parse, analyze, and store steps.
- [ ] Use `sync.WaitGroup` to close the results stream after all workers exit.
- [ ] Aggregate results and progress in one goroutine to avoid scattered counter mutexes.
- [ ] Let an individual file failure become a result rather than cancel the full job.
- [ ] Make cancellation stop accepting new work and allow in-flight operations to end at defined checkpoints.
- [ ] Do not leave partially inserted tracks or partial managed copies.

## Wails job bridge

- [ ] Add `StartBulkImport() (string, error)`.
- [ ] Open a native multi-file dialog filtered to GPX.
- [ ] Return an empty job ID for user cancellation without emitting an error.
- [ ] Start import in a goroutine only after selection succeeds.
- [ ] Add `CancelImport(jobID string) error`.
- [ ] Keep active job cancel functions in a mutex-protected map.
- [ ] Remove jobs from the map after their terminal event is emitted.
- [ ] Make repeated/late cancellation safe and return a useful not-active result.
- [ ] Do not expose a Go channel through Wails bindings.
- [ ] Register and emit a typed `summit:import-progress` event using the pinned Wails v3 event API.
- [ ] Regenerate bindings and typed event definitions.

## Progress UI

- [ ] Add `Import Multiple GPX Files` to the library.
- [ ] Subscribe to the typed progress event before starting the job when the API ordering permits.
- [ ] Track the active job ID and ignore stale events from earlier jobs.
- [ ] Display total and completed counts plus current filename/state.
- [ ] Provide a cancel button while the job is active.
- [ ] Disable duplicate bulk starts unless multiple concurrent jobs are intentionally supported; through Phase 7, support one active UI job.
- [ ] Show a completion summary with imported, duplicate, failed, and cancelled totals.
- [ ] Show individual file failures without replacing the whole library screen.
- [ ] Refresh the library after terminal completion or cancellation.
- [ ] Unsubscribe the event listener when the component unmounts.

## Tests

- [ ] Test that maximum concurrent parsers never exceeds the configured worker count.
- [ ] Test worker count for zero, one, and many inputs.
- [ ] Test all-success, all-duplicate, all-failure, and mixed-result batches.
- [ ] Test cancellation before workers start.
- [ ] Test cancellation during hashing/parsing.
- [ ] Test cancellation during/around persistence without partial rows.
- [ ] Test that one malformed GPX does not cancel valid imports.
- [ ] Test progress counter monotonicity and the terminal event.
- [ ] Test a progress consumer that exits early does not leak workers.
- [ ] Test repeated cancellation.
- [ ] Run importer/repository tests with `go test -race ./...`.
- [ ] Test frontend stale-job filtering, event cleanup, cancel state, and completion summaries.

## Learning checks

- [ ] Draw the ownership of jobs, worker channels, result channel, progress channel, and cancellation context.
- [ ] Explain which goroutine closes each channel.
- [ ] Explain why database correctness still relies on transactions rather than a mutex alone.
- [ ] Identify every shared map/counter and how it is protected or confined.
- [ ] Explain how cancellation differs from failure.

## Phase 7 gate

- [ ] Phase 6 gate still passes for single imports.
- [ ] Import at least 50 fixture copies or generated distinct fixtures.
- [ ] The UI remains responsive and progress reaches a terminal state.
- [ ] Cancel midway and confirm queued work stops.
- [ ] Confirm every database track is fully committed and loadable.
- [ ] Confirm mixed batches retain successful files and report failures.
- [ ] `go test -race ./...` passes repeatedly.
- [ ] No event listener receives duplicate updates after navigating/unmounting/remounting the library UI.

Suggested commits:

```text
feat: add concurrent track importer
feat: stream Wails import progress
test: cover cancellation and race safety
```

---

# Phase 8 — Hiking-focused GPS analysis

## Outcome

Summit reports useful movement, speed, pace, filtered elevation, grade, and outlier-aware statistics while preserving raw source data and explaining unavailable results.

## Analysis options

- [ ] Create one `AnalysisOptions` value owned by Go.
- [ ] Use documented defaults rather than scattered numeric constants.
- [ ] Keep options internal/application-configured through Phase 8; do not build a settings screen yet.
- [ ] Include at least:
  - Moving speed threshold: `0.5 km/h`
  - Long recording gap threshold
  - Likely impossible hiking speed threshold
  - Local spike distance/ratio thresholds
  - Elevation median window distance/sample count
  - Elevation low-pass strength/window
  - Ascent/descent reversal deadband: `3 m`
  - Grade window distance
  - Minimum horizontal distance for grade
- [ ] Put units in field names or documentation.
- [ ] Validate options at analyzer construction and reject nonsensical negative/zero values.

## Timestamp validation and duration model

- [ ] Validate time only within each GPX source segment.
- [ ] Treat an interval as timed only when both endpoints have timestamps and delta is positive.
- [ ] Flag nonmonotonic or duplicate timestamps instead of inventing durations.
- [ ] Identify long recording gaps separately from ordinary stopped time.
- [ ] Define and document:
  - Elapsed time: first valid timestamp to last valid timestamp where meaningful
  - Recorded time: sum of accepted positive intervals, excluding long gaps
  - Moving time: accepted intervals at or above movement threshold
  - Stopped time: accepted intervals below movement threshold
- [ ] Make unavailable durations nullable/flagged when timed coverage is insufficient.
- [ ] Ensure moving time plus stopped time does not exceed recorded time.

## Speed and pace

- [ ] Calculate interval speed from accepted horizontal distance and time.
- [ ] Exclude source-segment crossings, invalid time intervals, long gaps, and flagged outlier intervals.
- [ ] Calculate average moving speed from accepted moving distance divided by moving time.
- [ ] Calculate moving pace as moving time per kilometre.
- [ ] Keep calculations in base SI units; format km/h and min/km at the UI boundary.
- [ ] Never emit infinite pace for zero distance.
- [ ] Return availability and coverage information so partial timestamp data is not presented as complete.

## Coordinate outlier detection

- [ ] Preserve every original point and never rewrite the managed GPX.
- [ ] Continue rejecting structurally invalid coordinate ranges during parsing.
- [ ] Add diagnostics for likely GPS jumps.
- [ ] Use speed-based detection when reliable timestamps exist.
- [ ] Add a local geometric spike check: an isolated point with very long incoming/outgoing edges and a much shorter neighbor-to-neighbor path is suspicious.
- [ ] Avoid flagging a legitimate sparse straight section solely because point spacing is large.
- [ ] Record reason codes and affected point/interval indices.
- [ ] Exclude only flagged edges/points from affected derived statistics.
- [ ] Return outlier count and concise diagnostics to the UI.
- [ ] Make thresholds explicit enough to retune later for cycling activity profiles.

## Distance-based interpolation

- [ ] Implement lookup/interpolation at a requested cumulative distance.
- [ ] Clamp or explicitly reject distances outside route bounds; choose one behavior and test it.
- [ ] Locate surrounding points with binary search.
- [ ] Interpolate latitude, longitude, elevation, and timestamp only when both surrounding values are available.
- [ ] Return surrounding source indices and interpolation fraction.
- [ ] Never interpolate across a GPX source-segment break.
- [ ] Handle duplicate cumulative distances without division by zero.
- [ ] Reuse interpolation for exact segment boundaries and consistent map/profile lookup where appropriate.

## Elevation smoothing

- [ ] Preserve both raw and smoothed elevation series.
- [ ] Split filtering at source-segment boundaries and missing-elevation gaps.
- [ ] Apply a short median filter first to reject isolated vertical spikes.
- [ ] Apply a distance-aware low-pass filter second so uneven GPS sampling does not change smoothing strength arbitrarily.
- [ ] Preserve original point indices and cumulative distances.
- [ ] Do not fabricate elevation across long missing runs.
- [ ] Document endpoint behavior for each filter window.
- [ ] Confirm flat input remains flat and a monotonic climb remains monotonic within tolerance.

## Improved gain and loss

- [ ] Calculate ascent/descent from the smoothed series.
- [ ] Apply a 3 m reversal/deadband rule so small oscillations do not accumulate as repeated climb/loss.
- [ ] Reset accumulation across source segment breaks and missing-elevation gaps.
- [ ] Return raw and filtered gain/loss for diagnostics, but make filtered values the default summary.
- [ ] Recalculate analysis-segment gain/loss with the same filtered pipeline.
- [ ] Document the algorithm with a small worked example in code comments or package documentation.

## Grade

- [ ] Calculate grade over a centered distance window rather than adjacent raw points.
- [ ] Use interpolated/surrounding smoothed elevation at window endpoints.
- [ ] Require a minimum horizontal distance and sufficient elevation coverage.
- [ ] Return unavailable grade near gaps or when coverage is insufficient.
- [ ] Keep signed grade: climbing positive, descending negative.
- [ ] Recalculate analysis-segment average and maximum grade from the improved grade series.
- [ ] Prevent `NaN`, infinity, and extreme values caused by duplicate coordinates.

## DTO and UI changes

- [ ] Extend summary DTOs with timed coverage, elapsed/recorded/moving/stopped time, moving speed, moving pace, raw/filtered gain/loss, and outlier count.
- [ ] Extend profile points with raw elevation, smoothed elevation, grade, speed, and availability as appropriate.
- [ ] Keep payload size in mind; avoid duplicating full coordinate arrays unnecessarily.
- [ ] Regenerate bindings.
- [ ] Add summary cards for movement and filtered elevation statistics.
- [ ] Format durations consistently, including tracks over 24 hours.
- [ ] Format pace as `min/km` without displaying invalid zero/infinite values.
- [ ] Add a raw/smoothed elevation toggle to the profile.
- [ ] Show grade and available speed in the profile tooltip.
- [ ] Add an outlier indicator and a details view/list with point index and reason.
- [ ] Explain partial/unavailable time coverage instead of displaying zero.
- [ ] Use filtered gain/loss and grade in segment UI while retaining raw values only where they help comparison.
- [ ] Keep activity presets and user-adjustable analysis settings out of Phase 8.

## Golden fixtures

- [ ] Add a steady walking track with regular timestamps.
- [ ] Add a walking track with an intentional stopped interval.
- [ ] Add a track with a long timestamp gap.
- [ ] Add duplicate and nonmonotonic timestamp cases.
- [ ] Add a flat track with small sawtooth elevation noise.
- [ ] Add a known climb/descent with one large elevation spike.
- [ ] Add an isolated coordinate spike with good neighboring geometry.
- [ ] Add a legitimate sparse track that should not be marked as an outlier.
- [ ] Add missing-time, partial-time, missing-elevation, and multi-segment cases.
- [ ] Record expected values and tolerances beside each golden fixture.

## Tests and invariants

- [ ] Test moving/stopped classification at, above, and below the threshold.
- [ ] Test long-gap exclusion.
- [ ] Test average speed and pace from known time/distance.
- [ ] Test partial timed coverage and unavailable results.
- [ ] Test speed-based and geometric outlier detection independently.
- [ ] Test legitimate sparse data is retained.
- [ ] Test interpolation at start, exact points, between points, end, duplicate distance, gaps, and source breaks.
- [ ] Test median/low-pass filters with flat, impulse, monotonic, and missing data.
- [ ] Test deadband gain/loss against sawtooth noise and a real climb.
- [ ] Test positive, negative, zero, and unavailable grade.
- [ ] Test Phase 5 segment statistics now use the filtered pipeline.
- [ ] Assert moving time + stopped time <= recorded time.
- [ ] Assert gain/loss are nonnegative.
- [ ] Assert point ordering and original indices never change.
- [ ] Assert no DTO contains `NaN` or infinity.
- [ ] Assert interpolation endpoints match source values.
- [ ] Benchmark the full pipeline on a large generated track.
- [ ] Run `go test -race ./...`.

## Learning checks

- [ ] Explain why distance-window filters suit irregular GPS sampling better than fixed point counts alone.
- [ ] Explain the tradeoff between smoothing noise and erasing real terrain changes.
- [ ] Explain why raw data is preserved even when outliers are excluded from statistics.
- [ ] Explain why hiking thresholds should become activity profiles before adding cycling.
- [ ] Identify which algorithms are linear and which use binary search.

## Phase 8 gate

- [ ] Phase 7 gate still passes.
- [ ] A known stopped period appears in stopped time and not moving time.
- [ ] A long recording gap does not inflate moving or stopped time.
- [ ] A coordinate spike no longer dominates distance, speed, pace, or segment statistics.
- [ ] Sawtooth elevation noise produces materially less false gain after filtering.
- [ ] Raw and smoothed profiles remain visually comparable and keep the same source alignment.
- [ ] Missing timestamps/elevations yield explicit unavailable values rather than zeroes.
- [ ] Map, profile, point selection, segment selection, library loading, and bulk import still work.
- [ ] Go tests, race tests, frontend tests, type checking, frontend build, and `wails3 build` all pass.

Suggested commits:

```text
feat: add movement and pace analysis
feat: smooth elevation and grade
feat: detect GPS outliers
feat: expose advanced analysis UI
```

---

# Recurring definition of done

Run this checklist at the end of every phase, using the exact scripts created by the Wails template/project:

- [ ] Inspect `git status --short` and confirm every changed file belongs to the phase.
- [ ] Format Go code and verify no accidental generated/manual edits.
- [ ] Run `go vet ./...`.
- [ ] Run `go test ./...`.
- [ ] Run `go test -race ./...` once concurrency or shared mutable state exists.
- [ ] Regenerate Wails bindings after public service/DTO changes.
- [ ] Confirm binding regeneration is idempotent.
- [ ] Run frontend unit tests.
- [ ] Run TypeScript type checking.
- [ ] Run the production frontend build.
- [ ] Confirm new application UI uses Mantine components/theme tokens and adds no avoidable custom CSS.
- [ ] Confirm Zustand selectors do not cause unrelated map/chart panels to rerender on every pointer update.
- [ ] Run `wails3 dev` and execute the phase's manual acceptance flow.
- [ ] Run `wails3 build` before marking the phase complete.
- [ ] Update the progress table and completion date.
- [ ] Commit at a reviewable boundary with the suggested message or a clearer equivalent.

## Testing principles

- Prefer small deterministic fixtures.
- Compare geospatial floating-point results with explicit tolerances.
- Test observable behavior and invariants rather than private helper structure.
- Do not reproduce MapLibre, SQLite, or Wails internals in mocks.
- Keep automated tests independent of network tiles and the real user home directory.
- Use temporary directories and databases for filesystem/persistence tests.
- Keep at least one anonymized realistic GPX as an end-to-end regression fixture.

## Stop conditions

Pause progression and fix the current phase when any of these occurs:

- A test or production build is red.
- Generated bindings do not match exported Go methods/types.
- A domain formula appears in TypeScript.
- Missing values are silently converted to zero.
- A source GPX file is modified.
- A failed/cancelled import leaves partial database state.
- The race detector finds shared-state access.
- A new dependency has no concrete current-phase use.

Once Phase 8 passes, return to `summit-roadmap.md` and plan Phase 9 from the working application rather than extending this checklist speculatively.
