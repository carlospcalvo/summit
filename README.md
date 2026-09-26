# Summit

Summit is a local desktop application for exploring GPS tracks and understanding the terrain around them.

The project has one guiding idea:

> A fast desktop tool for understanding GPS tracks and terrain.

Summit is also a structured Go and Wails learning project. Each feature introduces a practical language, desktop, database, concurrency, or geospatial concept while contributing to a useful application.

## Project status

Summit is currently in the planning and bootstrap stage. The repository contains the product roadmap and detailed implementation checklists; the application scaffold has not been created yet.

The first usable milestone is:

```text
GPX inspection CLI
        ↓
Wails file opening
        ↓
Track statistics and elevation profile
```

Follow the [Phase 0–8 implementation checklist](docs/summit-phases-0-8-todo.md) to begin development.

## Planned capabilities

Summit will grow in deliberate stages:

- Open and inspect GPX tracks.
- Calculate distance, elevation gain/loss, bounds, speed, pace, and moving time.
- Render interactive elevation profiles and MapLibre maps.
- Synchronize point and segment selection across the map and chart.
- Organize tracks in a local SQLite library.
- Import many tracks concurrently with progress and cancellation.
- Detect GPS outliers and reduce elevation noise.
- Compare GPS elevation with local terrain data.
- Generate terrain cross-sections between two map points.
- Visualize slope along tracks and terrain profiles.
- Create and edit geographic waypoints.
- Import and export GPX and track-focused GeoJSON.
- Use locally imported PMTiles basemaps offline.
- Behave like a native macOS application with file associations, drag and drop, menus, and keyboard shortcuts.

## Architecture

Summit keeps domain and geospatial logic in Go. React and TypeScript handle presentation and interaction.

```text
React / TypeScript
  UI · MapLibre · SVG charts · forms · selection state
                       │
                       │ generated Wails bindings
                       ▼
Go application services
  files · parsing · analysis · persistence · concurrency
                       │
                       ▼
Local data
  GPX · GeoJSON · SQLite · SRTM HGT · PMTiles
```

### Go owns

- Filesystem access and managed local storage
- GPX and GeoJSON parsing
- Distance, elevation, grade, bounds, and interpolation
- Track segmentation and GPS analysis
- DEM terrain sampling
- SQLite persistence and migrations
- Import/export, caching, and concurrency

### React and TypeScript own

- Application rendering and controls
- MapLibre integration
- Responsive SVG charts
- Forms and accessible interaction
- Hover and selection state
- Purely visual transformations

When logic could reasonably live in either layer, prefer Go unless the behavior is purely visual.

## Planned technology

- [Go](https://go.dev/) for the domain, application, storage, and CLI layers
- [Wails v3](https://v3.wails.io/) for the desktop application and Go/TypeScript bridge
- [React](https://react.dev/) and TypeScript for the frontend
- [MapLibre GL JS](https://maplibre.org/maplibre-gl-js/docs/) for map rendering
- [SQLite](https://www.sqlite.org/) through `database/sql` for the local track library
- [SRTM HGT](https://lpdaac.usgs.gov/documents/179/SRTM_User_Guide_V3.pdf) for local terrain elevation
- [PMTiles](https://github.com/protomaps/PMTiles) for imported offline basemaps

Dependencies will be pinned when the project is scaffolded. Wails v3 is currently a deliberate beta dependency.

## Development roadmap

The work is split into 17 gated phases. Each phase includes implementation tasks, tests, manual acceptance checks, learning prompts, and suggested commit boundaries.

| Phases | Focus |
| --- | --- |
| 0–2 | GPX CLI, first Wails slice, and package structure |
| 3–5 | Elevation profile, MapLibre, and track segments |
| 6–8 | SQLite library, concurrent imports, and robust GPS analysis |
| 9–11 | Terrain elevation, cross-sections, and slope visualization |
| 12–13 | Waypoints and GeoJSON |
| 14–16 | Offline maps, export, and desktop polish |

Read the documents in this order:

1. [Product and learning roadmap](docs/summit-roadmap.md)
2. [Phase 0–8 implementation checklist](docs/summit-phases-0-8-todo.md)
3. [Phase 9–16 implementation checklist](docs/summit-phases-9-16-todo.md)

## Getting started

The repository is not runnable yet. To start Phase 0, prepare:

- macOS on Apple Silicon for the primary development environment
- Go 1.25 or newer
- Node.js and npm
- Xcode command-line tools
- Git
- The Wails v3 CLI

Install the Wails CLI:

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

Verify the desktop toolchain:

```bash
wails3 doctor
```

Then follow the bootstrap and Phase 0 sections in the [first implementation checklist](docs/summit-phases-0-8-todo.md). That checklist is the source of truth for scaffold commands, module naming, required fixtures, tests, and the phase completion gate.

The first planned CLI flow is:

```bash
summit inspect my-track.gpx
```

Expected output will resemble:

```text
Track: Base Sur Lanín
Points: 13,482
Distance: 21.42 km
Elevation gain: 1,147 m
Elevation loss: 1,153 m
Minimum: 812 m
Maximum: 1,694 m
```

## Development principles

- Build one complete vertical slice before generalizing it.
- Keep framework code at the edges.
- Preserve original imported files.
- Represent missing GPS and terrain values explicitly; never turn missing data into zero.
- Recompute cheap derived statistics instead of persisting every result.
- Introduce concurrency only for real workloads and always support cancellation.
- Add interfaces when there is a concrete boundary or second implementation.
- Keep fixtures small, deterministic, and anonymized.
- Require each phase gate to pass before moving forward.

## Current non-goals

These are intentionally outside the roadmap:

- Accounts and authentication
- Cloud sync and collaboration
- A server backend or microservices
- Subscription or billing systems
- A custom map renderer
- 3D terrain
- Mobile applications

The application is local-first. Geography and track analysis remain the center of the product.

## Repository layout

The current repository contains planning documents only:

```text
.
├── README.md
└── docs/
    ├── summit-roadmap.md
    ├── summit-phases-0-8-todo.md
    └── summit-phases-9-16-todo.md
```

The application, CLI, Go packages, frontend, fixtures, and build configuration will be introduced as the implementation checklists progress.
