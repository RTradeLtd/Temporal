package sm_yamux

import (
	"testing"

	test "github.com/libp2p/go-stream-muxer/test"
)

func TestYamuxTransport(t *testing.T) {
	test.SubtestAll(t, DefaultTransport)
}
