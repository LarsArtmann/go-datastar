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
// fetched verbatim from the upstream repo's bundles/ directory at the
// release tag (upstream releases publish no assets).
// Renovate's custom manager proposes Version bumps from upstream releases;
// the upgrade process (replace the file, bump Version, run the wire-format
// goldens) is documented in docs/static-js.md.
package static

import _ "embed"

// Version is the release version of the embedded DataStar JavaScript client.
const Version = "1.0.4"

// Named version constants for every DataStar runtime release this module has
// shipped, so consumers reference compile-time names instead of each
// redeclaring their own per-version constant lists (the templ-components
// datastar package maintained such a local copy until 2026-11-17 — this is
// the single source of truth now).
//
// The constants are untyped string constants: they stay assignable to any
// consumer's branded version type. Version1_0_4 always equals [Version] and
// moves with every pin bump; the older names are frozen literals.
const (
	// Version1_0_4 is the embedded (pinned) runtime version.
	Version1_0_4 = Version

	// Version1_0_3 is a compile-compatibility alias for consumers still
	// referencing the pre-1.0.4 name; it names the 1.0.3 CDN runtime,
	// deliberately NOT the pinned bytes. Deprecated: follow [Version] or
	// reference the current Version1_0_X constant.
	Version1_0_3 = "1.0.3"

	// Version1_0_2 is a compile-compatibility alias for consumers still
	// referencing the pre-1.0.3 name; it names the 1.0.2 CDN runtime.
	// Deprecated: follow [Version] or reference the current Version1_0_X
	// constant.
	Version1_0_2 = "1.0.2"
)

// VersionIsValid reports whether v is one of the defined version constants
// (the pinned version or a deprecated alias).
func VersionIsValid(v string) bool {
	switch v {
	case Version1_0_4, Version1_0_3, Version1_0_2:
		return true
	default:
		return false
	}
}

//go:embed datastar.js
var datastarJS []byte

// Bytes returns the embedded DataStar JavaScript client bundle.
//
// The returned slice is shared and must not be modified by the caller.
func Bytes() []byte { return datastarJS }
