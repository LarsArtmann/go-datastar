// Package static holds the embedded DataStar JavaScript client bundle and its
// version. It is a dedicated asset module so the [github.com/larsartmann/go-datastar]
// protocol package can serve the client without owning the bytes directly.
//
// The returned byte slice is the shared embedded value and must not be modified.
//
// # Provenance
//
// datastar.js is the upstream minified client bundle from
// github.com/starfederation/datastar at the release matching [Version],
// fetched from the project's release assets and committed verbatim.
// Renovate's custom manager proposes Version bumps from upstream releases;
// the upgrade process (replace the file, bump Version, run the wire-format
// goldens) is documented in docs/static-js.md.
package static

import _ "embed"

// Version is the release version of the embedded DataStar JavaScript client.
const Version = "1.0.2"

//go:embed datastar.js
var datastarJS []byte

// Bytes returns the embedded DataStar JavaScript client bundle.
//
// The returned slice is shared and must not be modified by the caller.
func Bytes() []byte { return datastarJS }
