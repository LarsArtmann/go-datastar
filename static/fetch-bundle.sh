#!/usr/bin/env bash
# Fetch the canonical minified DataStar client bundle at a given release tag
# and print everything the upgrade needs: the sha256 for
# static/checksum_test.go and the provenance line for the static.go package
# doc. Upstream releases publish no downloadable assets — the bundle lives in
# the repo's bundles/ directory at the tag.
#
# Usage (from anywhere; writes next to itself):
#   ./fetch-bundle.sh 1.0.3
#
# The downloaded file is NOT byte-modified: the committed bundle must stay
# identical to upstream so the checksum pin means something.
set -euo pipefail

version="${1:?usage: fetch-bundle.sh <version> (e.g. 1.0.3)}"
cd "$(dirname "$0")"

url="https://raw.githubusercontent.com/starfederation/datastar/v${version}/bundles/datastar.js"
out="datastar-${version}.js"

curl -fsSL "${url}" -o "${out}"

sum=$(sha256sum "${out}" | cut -d' ' -f1)

echo "fetched ${out} (${url})"
echo "sha256: ${sum}"
echo
echo "provenance line for static/static.go (package doc):"
echo "  datastar.js is the upstream minified client bundle from"
echo "  github.com/starfederation/datastar at v${version}, fetched verbatim"
echo "  from bundles/datastar.js at the tag (sha256 ${sum})."
echo
echo "Next steps — one commit (docs/static-js.md upgrade process):"
echo "  1. mv ${out} datastar.js"
echo "  2. update const Version in static.go to \"${version}\""
echo "  3. update bundleSHA256 in checksum_test.go to ${sum}"
echo "  4. run the full gate; if wire-format goldens change, that is a"
echo "     protocol change — verify against upstream and record it"
