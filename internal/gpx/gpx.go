// Package gpx reads GPX documents into the track domain model.
//
// It owns the XML vocabulary, tolerance of the GPX 1.0 and 1.1 namespaces, and
// validation of coordinate attributes. Parsing stops at the domain boundary: it
// performs no analysis, derives no statistics, and never silently repairs
// malformed input. A document that cannot be understood becomes a contextual
// error, so bad input is visible instead of quietly distorted.
//
// Filesystem concerns belong to callers. Parse consumes an io.Reader so the
// same code serves the desktop application, the CLI, and tests without knowing
// where bytes came from.
package gpx

import (
	"io"

	"github.com/carlospcalvo/summit/internal/track"
)

var ErrInvalidCoordinate, ErrInvalidElevation, ErrInvalidTimestamp error

func Parse(r io.Reader) (track.Track, error)
