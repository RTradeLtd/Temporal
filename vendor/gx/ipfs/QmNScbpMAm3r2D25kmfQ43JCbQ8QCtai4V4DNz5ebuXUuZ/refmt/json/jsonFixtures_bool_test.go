package json

import (
	"testing"

	"gx/ipfs/QmNScbpMAm3r2D25kmfQ43JCbQ8QCtai4V4DNz5ebuXUuZ/refmt/tok/fixtures"
)

func testBool(t *testing.T) {
	t.Run("bool true", func(t *testing.T) {
		checkCanonical(t, fixtures.SequenceMap["true"], `true`)
	})
	t.Run("bool false", func(t *testing.T) {
		checkCanonical(t, fixtures.SequenceMap["false"], `false`)
	})
}
