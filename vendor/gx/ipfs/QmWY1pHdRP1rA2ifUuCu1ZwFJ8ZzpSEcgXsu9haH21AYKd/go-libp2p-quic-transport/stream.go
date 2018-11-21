package libp2pquic

import (
	smux "github.com/libp2p/go-stream-muxer"
	quic "github.com/lucas-clemente/quic-go"
)

type stream struct {
	quic.Stream
}

var _ smux.Stream = &stream{}

func (s *stream) Reset() error {
	if err := s.Stream.CancelRead(0); err != nil {
		return err
	}
	return s.Stream.CancelWrite(0)
}
