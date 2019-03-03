package cbor

import (
	"testing"

	"gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt/tok/fixtures"
)

func testBool(t *testing.T) {
	t.Run("bool true", func(t *testing.T) {
		seq := fixtures.SequenceMap["true"]
		canon := b(0xf5)
		t.Run("encode canonical", func(t *testing.T) {
			checkEncoding(t, seq, canon, nil)
		})
		t.Run("decode canonical", func(t *testing.T) {
			checkDecoding(t, seq, canon, nil)
		})
	})
	t.Run("bool false", func(t *testing.T) {
		seq := fixtures.SequenceMap["false"]
		canon := b(0xf4)
		t.Run("encode canonical", func(t *testing.T) {
			checkEncoding(t, seq, canon, nil)
		})
		t.Run("decode canonical", func(t *testing.T) {
			checkDecoding(t, seq, canon, nil)
		})
	})
}
