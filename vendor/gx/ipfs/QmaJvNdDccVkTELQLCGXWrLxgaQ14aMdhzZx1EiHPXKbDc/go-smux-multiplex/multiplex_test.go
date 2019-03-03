package peerstream_multiplex

import (
	"testing"

	test "gx/ipfs/QmVtV1y2e8W4eQgzsP6qfSpCCZ6zWYE4m6NzJjB7iswwrT/go-stream-muxer/test"
)

func TestMultiplexTransport(t *testing.T) {
	test.SubtestAll(t, DefaultTransport)
}
