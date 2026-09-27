# Summit — Phase 9–16 Implementation Checklist

This checklist continues [`summit-phases-0-8-todo.md`](./summit-phases-0-8-todo.md). Start it only after the Phase 8 gate passes. Complete phases in order and do not advance while the current phase has a red test, broken build, data-loss risk, or unresolved acceptance item.

The product direction remains in [`summit-roadmap.md`](./summit-roadmap.md). This guide turns its later phases into concrete work, tests, learning checks, and reviewable commit boundaries.

## Progress

| Phase | Outcome | Status | Completed |
| --- | --- | --- | --- |
| 9 | Local terrain elevation from SRTM DEM tiles | Not started | — |
| 10 | A–B terrain cross-sections independent of tracks | Not started | — |
| 11 | Signed slope analysis and map visualization | Not started | — |
| 12 | Persisted, fully editable waypoints | Not started | — |
| 13 | Track-focused GeoJSON import | Not started | — |
| 14 | Locally imported PMTiles basemaps | Not started | — |
| 15 | GPX, GeoJSON, and CSV export | Not started | — |
| 16 | Native macOS workflow and release preparation | Not started | — |

Use `In progress`, `Blocked`, or `Complete` in the status column. Add a date only after the phase gate passes.

---

## Decisions carried forward

- [ ] Use module path `github.com/carlospcalvo/summit`.
- [ ] Continue with the pinned Wails v3 beta version; do not update it incidentally during a feature phase.
- [ ] Keep geospatial, filesystem, persistence, import/export, caching, and concurrency logic in Go.
- [ ] Keep MapLibre/SVG rendering, forms, and visual selection state in React/TypeScript.
- [ ] Use Mantine for application layout, forms, dialogs, notifications, tables, menus, loading states, spacing, and theming.
- [ ] Use Zustand for shared track, map, chart, segment, terrain-profile, and basemap interaction state.
- [ ] Keep isolated component state in React and substantial form state in Mantine Form.
- [ ] Limit hand-written CSS to MapLibre integration and specialized SVG chart behavior.
- [ ] Use metric base units and metric display units.
- [ ] Keep original imported files immutable.
- [ ] Store large managed source files outside SQLite and store their metadata in SQLite.
- [ ] Use additive, versioned database migrations.
- [ ] Keep generated Wails bindings generated and never hand-edit them.
- [ ] Target Apple Silicon macOS for required acceptance checks.
- [ ] Preserve every Phase 0–8 behavior and regression test.

## New decisions for Phases 9–16

- [ ] Use user-provided NASA SRTM 1 arc-second `.hgt` files for terrain elevation.
- [ ] Copy imported HGT files into Summit-managed storage.
- [ ] Generate terrain profiles along a straight A–B geodesic line.
- [ ] Limit a terrain profile to 200 km and at most 10,000 samples.
- [ ] Use signed slope and direction-aware colors.
- [ ] Provide complete local waypoint creation, editing, moving, and deletion.
- [ ] Support GeoJSON points and lines; reject polygonal/collection geometry clearly.
- [ ] Import user-provided PMTiles archives rather than bulk-downloading public tiles.
- [ ] Support raster PMTiles and Protomaps-compatible vector basemaps initially.
- [ ] Export GPX, track-focused GeoJSON, and segment-analysis CSV; defer PNG.
- [ ] Prepare GitHub Releases and secure updates, but keep updater delivery disabled until artifacts are signed and notarized.

## Plain-language GIS glossary

- **DEM:** a rectangular grid of terrain heights. Given latitude/longitude, Summit finds nearby grid cells and estimates the height.
- **SRTM HGT:** a simple DEM file containing one degree of terrain as signed 16-bit height samples. It has no embedded map projection or rich metadata.
- **Bilinear interpolation:** an estimate made by blending the four nearest grid-cell heights.
- **Geodesic:** the shortest path over the earth model between two coordinates.
- **Slope/grade:** elevation change divided by horizontal distance, expressed as a percentage.
- **PMTiles:** a single file containing many map tiles at multiple zoom levels.
- **HTTP byte range:** a request for only a selected portion of a file; PMTiles uses this to avoid loading the full archive.

## Reference documentation

- [NASA SRTM Global 1 arc-second dataset](https://data.nasa.gov/dataset/nasa-shuttle-radar-topography-mission-global-1-arc-second-v003-e47e1)
- [SRTM Collection User Guide](https://lpdaac.usgs.gov/documents/179/SRTM_User_Guide_V3.pdf)
- [MapLibre custom protocols](https://maplibre.org/maplibre-gl-js/docs/API/functions/addProtocol/)
- [PMTiles specification and implementations](https://github.com/protomaps/PMTiles)
- [PMTiles with MapLibre](https://docs.protomaps.com/pmtiles/maplibre)
- [OpenStreetMap tile usage policy](https://operations.osmfoundation.org/policies/tiles/)
- [Wails v3 HTTP service routes](https://v3.wails.io/features/bindings/services/)
- [Wails v3 native file drop](https://v3.wails.io/features/drag-and-drop/files/)
- [Wails v3 application menus](https://v3.wails.io/features/menus/application/)
- [Wails v3 macOS packaging](https://v3.wails.io/guides/build/macos/)
- [Wails v3 updater tutorial](https://v3.wails.io/tutorials/04-self-update-a-wails-app/)
- [Mantine](https://mantine.dev/)
- [Mantine Form](https://mantine.dev/form/package/)
- [Mantine notifications](https://mantine.dev/x/notifications/)
- [Zustand](https://zustand.docs.pmnd.rs/)

## Frontend implementation rule

- [ ] Build ordinary UI from Mantine components before considering a custom component.
- [ ] Use the centralized Summit theme and Mantine spacing/color/typography tokens rather than page-specific CSS.
- [ ] Use Mantine Form for waypoint and import/export forms, Mantine modals for confirmation, and Mantine notifications/alerts for operation results.
- [ ] Use Zustand selectors/actions for state shared between the map, SVG profile, lists, and tool panels.
- [ ] Keep Wails/Go results authoritative; do not turn Zustand into a second domain database.
- [ ] Retain the custom SVG elevation renderer and MapLibre layers because their indexed interaction and rendering requirements are application-specific.

---

# Phase 9 — DEM terrain data

## Outcome

Users can import local SRTM `.hgt` tiles and compare terrain elevation with raw and smoothed GPX elevation. Summit never changes the original track or DEM source.

## Understand the file format first

- [ ] Read the SRTM format section of the official user guide before implementing the reader.
- [ ] Document these assumptions beside the reader:
  - One tile covers one degree of latitude and longitude.
  - The filename identifies its southwest corner, for example `S40W072.hgt`.
  - A 1 arc-second tile contains 3,601 × 3,601 samples.
  - Each sample is a signed 16-bit big-endian integer in metres.
  - Rows are stored north to south; columns west to east.
  - `-32768` represents unavailable/nodata elevation.
- [ ] Reject `.hgt` files whose names or exact byte size do not match supported SRTM1 data.
- [ ] Do not guess a different resolution from an arbitrary file size in this phase.

## Models and interfaces

- [ ] Add a reusable `TerrainProvider` interface because both track correction and free terrain profiles need the same source:

  ```go
  type TerrainProvider interface {
      ElevationAt(ctx context.Context, c Coordinate) (ElevationSample, error)
      Coverage(ctx context.Context, bounds Bounds) (CoverageReport, error)
  }
  ```

- [ ] Define `ElevationSample` with coordinate, nullable elevation, source tile ID/name, approximate resolution, and availability/nodata reason.
- [ ] Define typed errors or reason codes for missing tile, nodata, invalid coordinate, corrupt tile, and cancelled lookup.
- [ ] Distinguish missing coverage from a genuine elevation near/below sea level.
- [ ] Keep the interface independent of HGT filenames and SQLite.

## Coordinate-to-raster lookup

- [ ] Derive the tile key from `floor(latitude)` and `floor(longitude)`, including negative coordinates.
- [ ] Format north/south and east/west filenames with correct zero padding.
- [ ] Convert coordinate position within the tile to fractional row/column.
- [ ] Account for file rows running north-to-south.
- [ ] Clamp an exact north/east tile boundary consistently to the adjacent tile rather than reading past a file edge.
- [ ] Read the four surrounding signed samples using explicit big-endian decoding.
- [ ] Apply bilinear interpolation only when all required corner samples are valid.
- [ ] Return nodata when interpolation touches nodata; do not silently turn nodata into zero.
- [ ] Return exact sample values at exact grid intersections.
- [ ] Check context cancellation before opening and before/after multiple disk reads.

## Managed DEM storage

- [ ] Add a database migration for `dem_tiles` with stable ID, tile key, managed relative path, hash, bounds, resolution, byte size, source label, and import timestamp.
- [ ] Add a unique constraint for tile key and content hash behavior.
- [ ] Create a managed `dem/` directory beneath Summit's application-data root.
- [ ] Import through a temporary file, SHA-256 hash, validation, atomic rename, then database transaction.
- [ ] Replace an existing tile key only through an explicit confirmed action; keep the prior tile if the new import fails.
- [ ] Treat identical content as a duplicate and return the existing tile.
- [ ] Never delete the user's original HGT file.
- [ ] Remove the managed file only after a successful database removal decision.

## Reader and cache

- [ ] Use `os.File.ReadAt` for small random reads instead of loading each roughly 25 MB tile entirely into memory.
- [ ] Add a bounded, concurrency-safe LRU cache of open tile readers.
- [ ] Give the cache a small explicit default capacity, such as eight files.
- [ ] Close a file when it is evicted.
- [ ] Avoid holding the cache mutex during disk reads.
- [ ] Prevent two concurrent misses from leaking duplicate open file handles.
- [ ] Close all cached readers during application shutdown.
- [ ] Detect a managed file that disappears or changes size after import and return a useful corruption/missing-file error.

## Track terrain analysis

- [ ] Add terrain elevation lookup for every accepted track point.
- [ ] Keep raw GPS, smoothed GPS, and terrain elevation as distinct series.
- [ ] Preserve source point indices and source GPX segment breaks.
- [ ] Do not persist a modified GPX track.
- [ ] Cache derived terrain results only if profiling proves repeated lookup too slow; key any cache by track data version and DEM tile hashes.
- [ ] Calculate terrain-based gain/loss and grade through the Phase 8 analysis pipeline.
- [ ] Return a coverage report with covered, missing-tile, and nodata point counts.
- [ ] Permit partial coverage and render gaps rather than failing the whole track.

## Wails API and UI

- [ ] Add `ImportDEMTiles() ([]DEMTileImportResult, error)` with multi-file selection.
- [ ] Add `ListDEMTiles() ([]DEMTileSummary, error)`.
- [ ] Add `RemoveDEMTile(id string) error`.
- [ ] Add `GetTerrainAnalysis(trackID string) (*TerrainTrackAnalysis, error)` or an equivalent focused call.
- [ ] Regenerate Wails bindings.
- [ ] Add a terrain-data management panel with import, coverage, duplicate/replacement, remove, and error states.
- [ ] Explain HGT filenames and missing-tile messages in plain language.
- [ ] Add raw GPS / smoothed GPS / terrain elevation choices to the profile.
- [ ] Show coverage percentage and missing tile names.
- [ ] Keep the existing smoothed GPS view usable when no DEM is installed.

## Fixtures and tests

- [ ] Generate tiny synthetic raster readers for most unit tests rather than committing full SRTM tiles.
- [ ] Keep one small supported-format fixture only if its repository size is acceptable.
- [ ] Test filename parsing in all hemispheres and at zero latitude/longitude.
- [ ] Test negative-coordinate flooring.
- [ ] Test north-to-south row orientation and byte offsets.
- [ ] Test signed big-endian decoding, including negative terrain.
- [ ] Test exact-cell and bilinear interpolation values.
- [ ] Test nodata in each of the four corners.
- [ ] Test tile north/east/south/west edges and transitions.
- [ ] Test invalid name, size, truncation, missing managed file, and cancellation.
- [ ] Test LRU hit, miss, eviction, close, concurrent lookup, and shutdown.
- [ ] Test import duplicates, replacement rollback, and filesystem/database cleanup.
- [ ] Test partial track coverage and preservation of raw elevation.
- [ ] Benchmark repeated sampling across one tile and across many tiles.
- [ ] Run the race detector against cache tests.

## Learning checks

- [ ] Explain why HGT needs filename metadata to locate its samples.
- [ ] Calculate the byte offset for a chosen row/column by hand.
- [ ] Explain endianness and nodata handling.
- [ ] Explain bilinear interpolation using four cell values.
- [ ] Explain why an open-file LRU is different from caching whole rasters.

## Phase 9 gate

- [ ] Phase 8 gate still passes.
- [ ] Import a real SRTM tile and query several coordinates against an external reference.
- [ ] A track crossing two installed tiles has a continuous terrain profile within expected DEM error.
- [ ] Missing tiles and nodata render gaps and identify the cause.
- [ ] Removing a tile closes cached handles and removes only Summit's managed copy.
- [ ] Raw GPX data remains byte-for-byte unchanged.
- [ ] Tests, race tests, frontend checks, and `wails3 build` pass.

Suggested commits:

```text
feat: import and read SRTM terrain tiles
feat: add terrain elevation sampling
feat: compare GPS and terrain elevation
```

---

# Phase 10 — Terrain cross-sections

## Outcome

Users can place points A and B on the map and see the terrain elevation between them without loading or creating a GPX track.

## Interaction model

- [ ] Add a `Terrain Profile` map mode that is visually distinct from ordinary track selection.
- [ ] First map click sets A; second click sets B and starts generation.
- [ ] A third click starts a new profile by replacing A and clearing B.
- [ ] Provide explicit clear/cancel controls and Escape-key behavior.
- [ ] Allow A and B to be dragged; regenerate after drag end, not every pointer frame.
- [ ] Keep the generated profile ephemeral through this phase.
- [ ] Preserve the loaded track/library selection while terrain-profile mode is active.

## Request and result models

- [ ] Define:

  ```go
  type TerrainProfileRequest struct {
      Start Coordinate `json:"start"`
      End   Coordinate `json:"end"`
  }
  ```

- [ ] Return total distance, actual sample spacing, samples, covered/missing counts, minimum/maximum elevation, gain/loss, and missing tile names.
- [ ] Give every sample a stable index, cumulative distance, coordinate, nullable elevation, and terrain source metadata.
- [ ] Keep unavailable statistics nullable.

## Geodesic sampling

- [ ] Calculate A–B distance with the existing geospatial model.
- [ ] Reject distances above 200 km with a message that explains the limit.
- [ ] Use nominal spacing near 30 m, matching SRTM1 resolution.
- [ ] Increase spacing when necessary to cap the result at 10,000 samples.
- [ ] Include A and B exactly.
- [ ] Generate intermediate coordinates along a geodesic rather than linear latitude/longitude interpolation.
- [ ] Handle antimeridian crossing and near-polar inputs even though SRTM coverage is limited by latitude.
- [ ] Check cancellation throughout sampling.
- [ ] Use the Phase 9 provider for every elevation lookup.

## Service and cancellation

- [ ] Add `GenerateTerrainProfile(request TerrainProfileRequest) (*TerrainProfile, error)`.
- [ ] Cancel an in-flight request when A/B changes or mode is cleared.
- [ ] Prevent a stale response from replacing a newer profile.
- [ ] Regenerate bindings.

## Map and chart

- [ ] Render A and B markers and their connecting line above the basemap.
- [ ] Fit or pad the line only on first completion; do not fight later user navigation.
- [ ] Reuse the elevation-profile component through an explicit generic series input.
- [ ] Show distance, terrain elevation, grade when available, and missing coverage in the tooltip.
- [ ] Synchronize chart hover with a highlighted point on the A–B line.
- [ ] Synchronize map sample hit testing with chart selection.
- [ ] Render missing DEM coverage as chart gaps and map-line styling.

## Tests

- [ ] Test zero-distance A/B.
- [ ] Test short profiles with fewer than two nominal intervals.
- [ ] Test exact inclusion of endpoints.
- [ ] Test nominal and adaptive sample spacing.
- [ ] Test the 10,000-sample and 200 km limits.
- [ ] Test antimeridian geodesics.
- [ ] Test tile transitions, nodata, missing tiles, and partial statistics.
- [ ] Test cancellation and stale-response suppression.
- [ ] Test A/B state transitions, clear, Escape, and drag completion.
- [ ] Test chart/map sample-index synchronization.

## Learning checks

- [ ] Explain why a straight line on screen is not necessarily linear latitude/longitude interpolation.
- [ ] Explain the sample cap and its memory/UI benefit.
- [ ] Explain how cancellation prevents stale UI.
- [ ] Identify which Phase 3 and Phase 9 components were reused.

## Phase 10 gate

- [ ] Phase 9 gate still passes.
- [ ] Generate a fully covered real terrain profile and compare key heights with source data.
- [ ] Generate a profile crossing a missing tile and verify a clear partial result.
- [ ] Drag A/B repeatedly and confirm only the newest result remains.
- [ ] The feature works without a GPX track loaded.
- [ ] Existing track selection and profiles remain intact after leaving the mode.

Suggested commits:

```text
feat: sample geodesic terrain profiles
feat: add map terrain-profile mode
```

---

# Phase 11 — Slope visualization

## Outcome

Tracks and A–B terrain profiles are divided into meaningful signed-slope sections and colored consistently on the map and chart.

## Slope model

- [ ] Define `SlopeDirection` as ascent, descent, or flat.
- [ ] Define `SlopeBucket` for `<5`, `5–10`, `10–15`, `15–20`, and `>20` percent magnitude.
- [ ] Define `SlopeSection` with start/end coordinate, start/end distance, source indices, signed grade, direction, bucket, and elevation source.
- [ ] Keep bucket membership separate from presentation colors.
- [ ] Use signed grade while classifying magnitude by absolute value.

## Calculation

- [ ] Reuse Phase 8 distance-window grade rather than adjacent raw points.
- [ ] Support smoothed GPS elevation and DEM terrain elevation.
- [ ] Never mix elevation sources inside one section.
- [ ] Split at GPX source breaks, missing elevation, terrain gaps, and outliers.
- [ ] Apply exact bucket boundaries consistently: 5% belongs to `5–10`, 10% to `10–15`, and so on.
- [ ] Treat values close to zero as flat using one documented epsilon.
- [ ] Merge adjacent intervals only when source, direction, and bucket match.
- [ ] Preserve enough indices/coordinates to draw exact geometry.
- [ ] Ensure merging does not cross a gap or source segment.

## Rendering

- [ ] Add a slope-visualization toggle and elevation-source selector.
- [ ] Use gray for flat/gentle sections, blue progression for descents, and yellow/orange/red progression for ascents.
- [ ] Provide a legend with direction signs and numeric ranges.
- [ ] Add line patterns or textual labels so meaning is not color-only.
- [ ] Render sections as MapLibre features keyed by stable ID.
- [ ] Show signed grade, bucket, direction, and distance range on hover/selection.
- [ ] Apply matching colors/regions to the elevation chart.
- [ ] Keep selected point/segment styling visible above slope colors.

## Tests

- [ ] Test every exact bucket boundary and values immediately around it.
- [ ] Test positive, negative, zero, and unavailable grade.
- [ ] Test gap/source-break splitting.
- [ ] Test adjacent-section merging and non-merging.
- [ ] Test no invalid grade from duplicate coordinates.
- [ ] Test both elevation sources.
- [ ] Test DTO-to-MapLibre feature mapping and legend labels.
- [ ] Test selection remains visible in slope mode.

## Learning checks

- [ ] Explain why grade is signed but buckets use magnitude.
- [ ] Explain why intervals are merged for rendering.
- [ ] Explain why source/gap boundaries prevent merging.

## Phase 11 gate

- [ ] Phase 10 gate still passes.
- [ ] A known climb and descent receive correct direction and buckets.
- [ ] Track map and chart use the same section IDs and ranges.
- [ ] Terrain-profile slope visualization behaves identically.
- [ ] The legend remains understandable without relying only on color.
- [ ] Large tracks do not create excessive MapLibre feature counts.

Suggested commits:

```text
feat: calculate signed slope sections
feat: visualize track and terrain slope
```

---

# Phase 12 — Full waypoint editing

## Outcome

Users can import, create, edit, move, filter, associate, and delete geographic waypoints without turning Summit into a general notes application.

## Domain and persistence

- [ ] Add stable waypoint IDs.
- [ ] Make track association nullable so standalone waypoints are supported.
- [ ] Define waypoint types: Camp, Water, Summit, Pass, Parking, Shelter, Danger, and Custom.
- [ ] Store name, notes, type, coordinate, nullable elevation, elevation source, created/updated timestamps, and optional track ID.
- [ ] Keep notes plain text with a reasonable length limit.
- [ ] Use fixed built-in icons/colors; do not support uploaded icon files.
- [ ] Add a migration from the Phase 8 waypoint table without losing imported waypoints.
- [ ] Map unknown imported types to Custom while preserving the original value in source properties where available.

## Application API

- [ ] Add `CreateWaypoint(input) (*WaypointDetails, error)`.
- [ ] Add `UpdateWaypoint(id, input) (*WaypointDetails, error)`.
- [ ] Add `MoveWaypoint(id, coordinate) (*WaypointDetails, error)`.
- [ ] Add `DeleteWaypoint(id) error`.
- [ ] Add `ListWaypoints(filter) ([]WaypointSummary, error)`.
- [ ] Validate coordinate range, type, note/name length, and referenced track existence in Go.
- [ ] Use optimistic UI only for drag preview; make the saved Go result authoritative.
- [ ] Regenerate bindings.

## Map and form interaction

- [ ] Add `Create Waypoint` mode and cursor state.
- [ ] Create at the clicked coordinate, then open a compact editor.
- [ ] Build the editor with Mantine Form and Mantine inputs/selects/textarea inside a Drawer or Modal.
- [ ] Default type to Custom and leave elevation unavailable unless deliberately filled.
- [ ] Add name, type, notes, track association, coordinate, and elevation display fields.
- [ ] Add an explicit `Use terrain elevation` action using Phase 9 coverage.
- [ ] Allow marker dragging; save on drag end and revert visually on failure.
- [ ] Require confirmation before deletion.
- [ ] Distinguish imported, user-entered, GPS, and terrain elevation sources where applicable.

## List and selection

- [ ] Add a waypoint list with type and track filters.
- [ ] Keep standalone waypoints visible when no track is selected.
- [ ] Selecting a list item centers/highlights its map marker without excessive zoom.
- [ ] Selecting a marker opens the same details/editor state.
- [ ] Add keyboard-accessible create, edit, save, cancel, and delete flows.
- [ ] Keep marker type distinguishable by icon/label as well as color.

## Tests

- [ ] Test migration of existing GPX waypoints.
- [ ] Test standalone and track-associated CRUD.
- [ ] Test invalid coordinates, missing tracks, bad types, and text limits.
- [ ] Test drag success, failure rollback, and overlapping updates.
- [ ] Test terrain elevation fill with coverage, nodata, and missing tile.
- [ ] Test deletion cascade when a track is removed according to the chosen database rule.
- [ ] Test type/filter behavior and selection synchronization.
- [ ] Test keyboard and accessible labels.

## Learning checks

- [ ] Explain why waypoint track association is optional.
- [ ] Explain why drag preview lives in React while validation/persistence lives in Go.
- [ ] Explain how migration preserves earlier imported waypoints.

## Phase 12 gate

- [ ] Phase 11 gate still passes.
- [ ] Create and edit each waypoint type.
- [ ] Drag a marker, restart Summit, and verify its coordinate persisted.
- [ ] Create a standalone waypoint and associate it with a track later.
- [ ] Terrain fill never overwrites an elevation without explicit action.
- [ ] Deletion requires confirmation and removes the correct record only.
- [ ] GPX-imported waypoints still render and edit correctly.

Suggested commits:

```text
feat: add persisted waypoint model
feat: add waypoint editing and map interaction
```

---

# Phase 13 — GeoJSON import

## Outcome

Summit imports track-like GeoJSON points and lines into the same library and analysis pipeline as GPX, while reporting unsupported geometry precisely.

## Supported input

- [ ] Accept top-level `FeatureCollection`, `Feature`, Point, MultiPoint, LineString, and MultiLineString.
- [ ] Reject Polygon, MultiPolygon, and GeometryCollection with typed unsupported-geometry results.
- [ ] Accept standard `[longitude, latitude]` and optional third elevation coordinate.
- [ ] Reject positions with fewer than two numbers or invalid coordinate ranges.
- [ ] Ignore coordinate elements beyond elevation but preserve the original feature properties.
- [ ] Treat each LineString feature as a track.
- [ ] Treat a MultiLineString as one track with multiple source segments.
- [ ] Treat points as waypoints; associate them only when an explicit supported property identifies an imported track.

## Models and parser

- [ ] Introduce typed geometry variants rather than passing arbitrary maps through analysis code.
- [ ] Use `map[string]any` only for GeoJSON properties at the interchange boundary.
- [ ] Add `Feature` with typed geometry, properties, and optional source feature ID.
- [ ] Parse with `encoding/json.Decoder` and enforce a sensible maximum input size/depth policy.
- [ ] Map known properties: name, time, waypoint type, notes, and track association.
- [ ] Preserve unknown source properties as JSON for export.
- [ ] Reject non-finite numeric values.
- [ ] Include feature index/ID in parse errors.

## Import and persistence

- [ ] Extend imports with source format (`gpx` or `geojson`).
- [ ] Add retained source-properties JSON to tracks/waypoints where needed.
- [ ] Copy the original GeoJSON into managed source storage through hash, temporary file, and atomic rename.
- [ ] Apply duplicate detection by source bytes while allowing distinct features within one source to create distinct entities.
- [ ] Import supported features even when other features are unsupported or invalid.
- [ ] Return a report listing created tracks/waypoints, duplicates, skipped features, and errors.
- [ ] Wrap all database inserts for one source file in a policy that preserves supported successes only if their managed source/import record can be committed coherently.

## UI

- [ ] Add GeoJSON filters to native open/import dialogs.
- [ ] Route `.geojson` and `.json` drops to GeoJSON parsing.
- [ ] Show a Mantine modal or drawer preview/import summary for multi-feature files.
- [ ] Clearly identify unsupported polygon/collection features.
- [ ] Load imported line tracks through the existing map/profile/analysis UI.
- [ ] Load imported point features through the waypoint UI.
- [ ] Avoid implying that arbitrary `.json` files are valid GeoJSON.

## Tests

- [ ] Test every supported top-level and geometry type.
- [ ] Test longitude/latitude ordering and optional elevation.
- [ ] Test multiple source segments from MultiLineString.
- [ ] Test known and retained unknown properties.
- [ ] Test malformed JSON, excessive input, invalid coordinates, null geometry, and unsupported geometry.
- [ ] Test mixed valid/invalid collections and import reports.
- [ ] Test duplicate source imports.
- [ ] Test managed-source persistence after original deletion.
- [ ] Test imported tracks in Phase 8 analysis and Phase 11 slope mode.
- [ ] Test points in Phase 12 waypoint editing.

## Learning checks

- [ ] Explain GeoJSON coordinate order.
- [ ] Explain why geometry is typed but properties remain flexible.
- [ ] Explain partial-success behavior for feature collections.
- [ ] Explain how this moves Summit beyond GPX without becoming a general GIS viewer.

## Phase 13 gate

- [ ] Phase 12 gate still passes.
- [ ] Import a FeatureCollection containing lines and points.
- [ ] MultiLineString breaks remain separate during distance analysis.
- [ ] Unsupported polygons are reported and never interpreted as tracks.
- [ ] Imported source properties survive a database restart.
- [ ] Removing the original GeoJSON does not break the managed library item.

Suggested commits:

```text
feat: parse track-focused GeoJSON
feat: import GeoJSON tracks and waypoints
```

---

# Phase 14 — Offline PMTiles maps

## Outcome

Users can import a lawful PMTiles archive, choose it as a basemap, and use Summit with networking disabled.

## Policy boundary

- [ ] Do not bulk-download or prefetch `tile.openstreetmap.org`; its policy prohibits offline downloading.
- [ ] Accept only user-provided/local PMTiles archives in this phase.
- [ ] Show attribution stored in archive metadata/style whenever the map is visible.
- [ ] Tell users they are responsible for using an archive whose license permits their use.

## Archive support

- [ ] Support PMTiles specification version 3.
- [ ] Read and validate header, directory metadata, bounds, zoom range, tile type, compression, and attribution.
- [ ] Support raster archives without a custom vector style.
- [ ] Support Protomaps-compatible vector basemap schema with bundled offline style, sprite, and glyph assets.
- [ ] Detect unsupported vector schemas and report them rather than rendering a blank map.
- [ ] Do not claim arbitrary thematic vector PMTiles support.

## Managed storage and database

- [ ] Add `offline_maps` migration with ID, name, managed path, SHA-256, size, bounds, zoom range, tile/compression type, schema/style kind, attribution, and imported timestamp.
- [ ] Create a managed `maps/` directory.
- [ ] Stream hash/copy through a temporary file and atomic rename.
- [ ] Reject invalid headers before committing database state.
- [ ] Treat matching hashes as duplicates.
- [ ] Prevent removal of the currently active archive until another basemap is selected or the user confirms fallback.
- [ ] Remove only Summit's managed copy.

## Local byte-range route

- [ ] Register a Wails HTTP service route such as `/offline-maps/`.
- [ ] Resolve archive IDs through the repository; never accept raw filesystem paths.
- [ ] Support `HEAD` and single-range `GET` requests.
- [ ] Return correct `Accept-Ranges`, `Content-Length`, `Content-Range`, content type, status codes, and cache headers.
- [ ] Reject malformed/multiple/out-of-bounds ranges safely.
- [ ] Support request cancellation and close files promptly.
- [ ] Prevent path traversal and ID guessing from exposing unrelated files.
- [ ] Do not load full archives into memory.

## MapLibre integration

- [ ] Add the official `pmtiles` JavaScript package and pin it.
- [ ] Register its MapLibre custom protocol once and clean it up only when the app lifecycle requires it.
- [ ] Point archive requests at the Wails local range route.
- [ ] Build raster or supported vector MapLibre style definitions from validated metadata.
- [ ] Bundle every font glyph/sprite needed by the offline vector style.
- [ ] Add online/offline basemap selection and persist the selected basemap ID.
- [ ] Fall back gracefully when a selected archive is missing/corrupt.
- [ ] Preserve Summit track, slope, waypoint, and selection layers across basemap switches.

## UI

- [ ] Add import, list, select, inspect, rename-display-name, and remove controls.
- [ ] Use Mantine tables/lists, modals, alerts, and notifications rather than custom-styled equivalents.
- [ ] Display archive bounds, zoom range, size, format, attribution, and support status.
- [ ] Fit to archive bounds only through an explicit action.
- [ ] Indicate clearly when Summit is using a fully local basemap.
- [ ] Explain unsupported vector-schema errors in plain language.

## Tests

- [ ] Use tiny generated raster and supported-vector PMTiles fixtures.
- [ ] Test header/metadata validation and unsupported versions/types.
- [ ] Test hash duplicate handling and atomic import rollback.
- [ ] Test `HEAD`, full `GET`, prefix/suffix/open-ended ranges, invalid ranges, cancellation, and concurrent requests.
- [ ] Test that routes cannot escape managed storage.
- [ ] Test missing/corrupt files and active-map removal rules.
- [ ] Test raster and supported-vector style creation.
- [ ] Test basemap switching preserves overlay sources/layers.
- [ ] Run a packaged application with networking disabled as a manual test.

## Learning checks

- [ ] Explain why PMTiles needs byte-range reads.
- [ ] Explain why a vector archive also needs a compatible style, sprites, and glyphs.
- [ ] Explain the boundary between Go file serving and MapLibre rendering.
- [ ] Explain why public OSM tiles cannot back an offline downloader.

## Phase 14 gate

- [ ] Phase 13 gate still passes.
- [ ] Import and render one real raster or Protomaps-compatible vector archive.
- [ ] Disable networking and confirm basemap, tracks, slopes, and waypoints still render.
- [ ] Attribution remains visible offline.
- [ ] Range requests never expose arbitrary local files.
- [ ] Switching basemaps does not duplicate or lose Summit overlay layers.
- [ ] Removing an inactive archive cleans its metadata and managed copy.

Suggested commits:

```text
feat: import and serve PMTiles archives
feat: add offline MapLibre basemaps
```

---

# Phase 15 — Export

## Outcome

Users can export a selected library track and its waypoints as GPX or GeoJSON, or export its analysis segments as CSV, without altering source data.

## Export architecture

- [ ] Introduce the justified interface:

  ```go
  type Exporter interface {
      Export(ctx context.Context, w io.Writer, data ExportData) error
  }
  ```

- [ ] Implement one exporter per format.
- [ ] Keep path selection/atomic filesystem writing outside format encoders.
- [ ] Define `ExportFormat` and `ElevationMode` enums.
- [ ] Support raw GPS, smoothed GPS, and DEM terrain elevation choices.
- [ ] Default to raw GPS elevation.
- [ ] Reject DEM export when coverage is insufficient unless the user explicitly accepts gaps.
- [ ] Never modify stored tracks, managed originals, or waypoints.

## GPX 1.1

- [ ] Emit valid UTF-8 GPX 1.1 XML with Summit creator/version metadata.
- [ ] Preserve source GPX segment boundaries.
- [ ] Export track name, coordinates, selected elevation source, and timestamps where available.
- [ ] Export associated waypoints with name, notes/description, elevation, and Summit waypoint type in a namespaced extension if needed.
- [ ] Escape text correctly.
- [ ] Omit unavailable optional elements rather than writing zero values.

## GeoJSON

- [ ] Export one LineString or MultiLineString feature for the track.
- [ ] Export associated waypoints as Point features.
- [ ] Use `[longitude, latitude]` and optional elevation.
- [ ] Emit known canonical properties from current Summit data.
- [ ] Merge retained source properties only when they do not replace canonical geometry/identity fields.
- [ ] Produce a FeatureCollection with stable ordering.

## CSV analysis

- [ ] Export UTF-8 RFC 4180-style CSV with a header row.
- [ ] Write one row per analysis segment.
- [ ] Include track ID/name, segment index, start/end/distance, gain/loss, average/maximum grade, moving time, speed/pace when available, elevation mode, and outlier count.
- [ ] Use base or clearly labelled metric units in column names.
- [ ] Leave unavailable cells empty rather than writing zero.
- [ ] Quote commas, quotes, and newlines correctly.

## Safe save flow

- [ ] Add `ExportTrack(trackID, format, elevationMode) error` at the Wails boundary.
- [ ] Choose the target with a native save dialog and format-appropriate default extension.
- [ ] Treat cancellation as success/no-op.
- [ ] Confirm or use platform behavior for overwrite.
- [ ] Write a temporary file in the destination directory.
- [ ] Flush, close, and atomically rename on success.
- [ ] Remove the temporary file after failure/cancellation.
- [ ] Preserve an existing target if encoding fails.
- [ ] Return contextual errors without exposing unrelated managed paths.
- [ ] Regenerate bindings.

## UI

- [ ] Add export from the selected track and native File menu.
- [ ] Use Mantine Form controls for format description and elevation-source choice.
- [ ] Disable unavailable elevation choices with an explanation.
- [ ] Show success with target filename and failure with retry.
- [ ] Keep PNG snapshot visibly out of scope rather than presenting a disabled promise.

## Tests

- [ ] Use golden exports for a small multi-segment track with waypoints and special characters.
- [ ] Parse exported GPX back through Summit and compare relevant domain values.
- [ ] Parse exported GeoJSON back through Summit and compare geometry/properties.
- [ ] Parse CSV with Go's CSV reader and compare columns/rows.
- [ ] Test all elevation modes, missing values, partial DEM, no waypoints, and multiple segments.
- [ ] Test property conflicts and coordinate order.
- [ ] Test writer failure, context cancellation, target overwrite, temporary cleanup, and atomic replacement.
- [ ] Test save-dialog cancellation and frontend availability states.

## Learning checks

- [ ] Explain why exporters accept `io.Writer`.
- [ ] Explain why safe file output needs a temporary file and rename.
- [ ] Explain the round-trip guarantees and deliberate information loss for each format.
- [ ] Identify why an exporter interface is justified here.

## Phase 15 gate

- [ ] Phase 14 gate still passes.
- [ ] GPX and GeoJSON round-trip through Summit within numeric tolerances.
- [ ] CSV opens correctly in a common spreadsheet application.
- [ ] Raw, smoothed, and terrain export choices produce the expected elevations.
- [ ] A forced encoder failure leaves an existing target unchanged.
- [ ] Original managed sources remain unchanged.

Suggested commits:

```text
feat: add GPX and GeoJSON exporters
feat: export segment analysis CSV
feat: add safe desktop export flow
```

---

# Phase 16 — Desktop polish and release preparation

## Outcome

Summit behaves like a native macOS application: files open and drop naturally, menus and shortcuts work, recent tracks and window state persist, packaged builds are recognizable, and secure GitHub releases are prepared without enabling unsigned automatic updates.

## Native file opening

- [ ] Register `.gpx` and `.geojson`/supported `.json` file associations in macOS bundle metadata.
- [ ] Handle files passed at cold launch and while the app is already running.
- [ ] Route associated files through the managed import pipeline, then select the imported/existing track.
- [ ] Deduplicate repeated OS open events.
- [ ] Queue early open events until database, migrations, services, and frontend are ready.
- [ ] Report unsupported or malformed associated files without blocking application startup.

## Native file drop

- [ ] Enable Wails file drop in window options.
- [ ] Add explicit drop targets and full-window visual feedback.
- [ ] Route GPX/GeoJSON to track import, HGT to DEM import, and PMTiles to offline-map import.
- [ ] Support mixed drops and return one grouped result summary.
- [ ] Ignore directories and unsupported extensions with a clear report.
- [ ] Prevent HTML/webview default navigation on dropped files.
- [ ] Reuse existing import validation; do not create a parallel parser path.

## Recent tracks

- [ ] Add/update `last_opened_at` for library tracks.
- [ ] Query the most recent ten existing track IDs.
- [ ] Build a dynamic `File → Open Recent` native menu with `Clear Menu`.
- [ ] Remove stale IDs automatically when tracks are deleted.
- [ ] Display track names, adding a disambiguator only for duplicates.
- [ ] Keep recents based on managed library IDs rather than fragile original paths.

## Native menus and shortcuts

- [ ] Add standard macOS application, File, Edit where meaningful, View, Window, and Help menus.
- [ ] Add Open/Import (`Cmd+O`), bulk import, export (`Cmd+Shift+E`), create waypoint, clear selection (`Esc` in-app), map/profile focus, basemap choice, and quit actions.
- [ ] Disable context-sensitive commands when no track/waypoint is selected.
- [ ] Keep native menu action and React button behavior routed through the same application action.
- [ ] Avoid intercepting standard text-editing shortcuts inside form fields.
- [ ] Add accessible shortcut hints in relevant UI controls.

## Window and UI state

- [ ] Store versioned desktop preferences in an atomic JSON config file, separate from track data.
- [ ] Persist window size, position, maximized state, panel visibility/width, selected basemap, and last library track ID.
- [ ] Debounce frequent resize/move writes.
- [ ] Validate minimum/maximum dimensions.
- [ ] Detect saved positions outside currently connected displays and recenter safely.
- [ ] Recover from corrupt/unknown-version config by preserving or renaming it and using defaults.
- [ ] Do not persist transient errors, hover, active import jobs, or unsaved forms.

## Visual and bundle polish

- [ ] Create a source app icon at sufficient resolution and generate Wails/macOS assets.
- [ ] Set product name, bundle identifier, semantic version, copyright, and description.
- [ ] Add polished loading, empty, offline, missing-data, and fatal-startup screens.
- [ ] Audit keyboard focus order, contrast, non-color status cues, and screen-reader names.
- [ ] Ensure no development menus/logging/devtools ship in production builds.
- [ ] Add an About window with version, commit, licenses, data attribution, and diagnostics-copy action.

## GitHub release preparation

- [ ] Adopt semantic versions and inject version/commit/build date at build time.
- [ ] Add GitHub Actions that run tests, race tests where supported, frontend checks, and Wails production build.
- [ ] Build a macOS `.app`/archive for tagged releases.
- [ ] Generate `SHA256SUMS` in the same release job as artifacts.
- [ ] Generate/update release notes from an explicit changelog or curated tag notes.
- [ ] Keep release jobs reproducible and pin action versions.
- [ ] Do not place Apple credentials, private signing keys, or updater signing keys in the repository.

## Signing and notarization — deliberately pending

- [ ] Document Apple Developer Program membership as a prerequisite.
- [ ] Document Developer ID Application certificate setup.
- [ ] Document hardened runtime and the minimum entitlements Summit needs.
- [ ] Document keychain/notarytool credential profile setup.
- [ ] Add Wails signing and notarization tasks without embedding credentials.
- [ ] Add stapling and Gatekeeper verification commands.
- [ ] Leave actual production signing/notarization unchecked until credentials exist.
- [ ] Do not label unsigned artifacts as production-ready macOS releases.

## Secure updater

- [ ] Integrate the Wails v3 updater against `carlospcalvo/summit` GitHub Releases.
- [ ] Check updates manually from the application menu before considering background checks.
- [ ] Require SHA-256 verification.
- [ ] Generate a dedicated Ed25519 update-signing key outside the repository and embed only its public key.
- [ ] Require signed update metadata/artifacts before installation.
- [ ] Publish release asset and checksum/signature metadata atomically.
- [ ] Handle no update, network failure, invalid checksum/signature, user cancellation, download progress, relaunch, and failed swap.
- [ ] Keep updater installation disabled until the downloaded `.app` is Developer ID signed and notarized.
- [ ] Allow unsigned development builds to exercise a mock/local updater path only.

## Tests and manual matrix

- [ ] Test cold-launch and warm-launch file-open events.
- [ ] Test duplicate and malformed open events before/after frontend readiness.
- [ ] Test every drop type plus mixed, directory, and unsupported drops.
- [ ] Test recent ordering, limits, stale deletion, clearing, and duplicate names.
- [ ] Test native menu enablement and shortcuts with/without form focus.
- [ ] Test config round-trip, corrupt config, unknown version, rapid resize, and off-screen restoration.
- [ ] Test production bundle metadata, icon, and file associations.
- [ ] Test updater no-update/update paths against a controlled fixture release.
- [ ] Test checksum/signature rejection and interrupted download cleanup.
- [ ] Test relaunch/rollback behavior without risking the development binary.
- [ ] Manually run a packaged build from Finder, with network on and off.
- [ ] When credentials become available, manually verify codesign, notarization, stapling, Gatekeeper launch, and update from the prior signed version.

## Learning checks

- [ ] Explain cold versus warm file-open events.
- [ ] Explain why recent items use library IDs rather than original paths.
- [ ] Explain atomic preference writes and off-screen window recovery.
- [ ] Explain the distinct roles of code signing, notarization, checksum, and update signature.
- [ ] Explain why the updater remains disabled for unsigned releases.

## Phase 16 gate

- [ ] Phase 15 gate still passes.
- [ ] GPX/GeoJSON double-click opens or imports the correct track in a packaged app.
- [ ] Drag/drop routes every supported file type through its existing importer.
- [ ] Recent tracks, menus, shortcuts, and window state survive restarts.
- [ ] A corrupt preference file cannot prevent Summit from opening.
- [ ] Tagged CI produces versioned artifacts and checksums.
- [ ] Updater rejects altered artifacts.
- [ ] Automatic installation remains disabled while signing/notarization is incomplete.
- [ ] Tests, frontend checks, production build, and packaged-app smoke tests pass.

Suggested commits:

```text
feat: add native file open and drop workflows
feat: add macOS menus and persistent window state
chore: add application branding and packaging metadata
ci: prepare GitHub release artifacts
feat: prepare verified application updates
```

---

# Database migration sequence

Keep exact migration numbers based on the Phase 8 repository. The logical sequence is:

- [ ] Phase 9: add `dem_tiles`.
- [ ] Phase 12: expand waypoint IDs, nullable track association, type, notes, elevation source, and timestamps.
- [ ] Phase 13: add import source format and retained GeoJSON source properties.
- [ ] Phase 14: add `offline_maps`.
- [ ] Phase 16: add `tracks.last_opened_at` if not already present.

For every migration:

- [ ] Upgrade a fixture database representing the previous phase.
- [ ] Preserve all existing track, point, waypoint, and import rows.
- [ ] Verify indexes, constraints, foreign keys, and nullability.
- [ ] Reopen the migrated database and exercise the affected repository.
- [ ] Refuse a schema newer than the application supports.
- [ ] Keep migrations additive unless a rebuild is required and tested transactionally.

# Recurring definition of done

- [ ] Inspect `git status --short` and confirm every change belongs to the phase.
- [ ] Format Go and verify generated files were not manually edited.
- [ ] Run `go vet ./...`.
- [ ] Run `go test ./...`.
- [ ] Run `go test -race ./...` for caches, HTTP serving, jobs, persistence, and shared state.
- [ ] Run focused benchmarks where a phase handles large rasters, profiles, archives, or tracks.
- [ ] Regenerate Wails bindings after public service/DTO changes.
- [ ] Confirm regeneration is idempotent.
- [ ] Run frontend tests and TypeScript type checking.
- [ ] Run the production frontend build and `wails3 build`.
- [ ] Confirm ordinary UI uses Mantine/theme tokens and introduces no avoidable custom CSS.
- [ ] Confirm Zustand selectors keep high-frequency map/chart interaction scoped to relevant subscribers.
- [ ] Exercise the phase's manual workflow in `wails3 dev` and a packaged build where required.
- [ ] Update progress and completion date.
- [ ] Commit at a reviewable boundary.

## Testing principles

- Generate small deterministic binary fixtures instead of committing large DEM/PMTiles files.
- Keep tests independent of internet access, the real home directory, GitHub, and Apple credentials.
- Use temporary app-data directories and databases.
- Compare geospatial floats with documented tolerances.
- Test gaps/nodata as first-class states, never as zero.
- Test cancellation and cleanup around every large-file operation.
- Keep at least one manual real-data acceptance case per terrain/offline phase.

## Stop conditions

Pause and fix the current phase when:

- A migration risks losing existing library data.
- A missing/nodata elevation becomes zero.
- A managed-file route can expose an arbitrary path.
- A failed import/export leaves a partial committed file or row.
- A renderer silently accepts unsupported geometry or vector-tile schema.
- An offline feature depends on remote fonts, sprites, styles, or tiles.
- A source GPX, GeoJSON, HGT, or PMTiles file is modified.
- The race detector reports shared-state access.
- An unsigned/unnotarized build can install itself as an automatic update.

When Phase 16 passes, Summit has completed this roadmap. Future work should start from observed product use rather than extending the phase list automatically.
