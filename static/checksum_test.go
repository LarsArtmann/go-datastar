package static_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/larsartmann/go-datastar/static"
)

// bundleSHA256 pins the exact committed bundle. Update it ONLY together with
// static/datastar.js and static.Version — a mismatch means the embedded asset
// drifted from the reviewed, committed artifact.
const bundleSHA256 = "4df1f98ac52c6ec8986375b8394c90ddd739a381cda05011edd1c18bf33a1625"

func TestBytes_Checksum(t *testing.T) {
	t.Parallel()

	sum := sha256.Sum256(static.Bytes())

	if got := hex.EncodeToString(sum[:]); got != bundleSHA256 {
		t.Errorf("bundle checksum drifted: got sha256 %s, want %s — "+
			"replace the checksum constant in the same commit as the bundle", got, bundleSHA256)
	}
}
