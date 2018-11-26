package json

import (
	"testing"

	"gx/ipfs/QmfWqohMtbivn5NRJvtrLzCW3EU4QmoLvVNtmvo9vbdtVA/refmt/tok/fixtures"
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
