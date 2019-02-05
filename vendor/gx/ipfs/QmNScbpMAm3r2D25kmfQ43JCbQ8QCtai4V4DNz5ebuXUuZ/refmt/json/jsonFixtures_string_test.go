package json

import (
	"testing"

	"gx/ipfs/QmNScbpMAm3r2D25kmfQ43JCbQ8QCtai4V4DNz5ebuXUuZ/refmt/tok/fixtures"
)

func testString(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		seq := fixtures.SequenceMap["empty string"]
		checkCanonical(t, seq, `""`)
		t.Run("decode with extra whitespace", func(t *testing.T) {
			checkDecoding(t, seq, `  "" `, nil)
		})
	})
	t.Run("flat string", func(t *testing.T) {
		seq := fixtures.SequenceMap["flat string"]
		checkCanonical(t, seq, `"value"`)
	})
	t.Run("strings needing escape", func(t *testing.T) {
		seq := fixtures.SequenceMap["strings needing escape"]
		checkCanonical(t, seq, `"str\nbroken\ttabbed"`)
	})
}
