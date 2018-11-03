package peerstream_multiplex

import (
	"testing"

	test "github.com/libp2p/go-stream-muxer/test"
)

func TestMultiplexTransport(t *testing.T) {
	test.SubtestAll(t, DefaultTransport)
}
