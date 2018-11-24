// +build !linux

package singlepoll

import (
	"context"
	"errors"

	"gx/ipfs/QmXD921xzL9EDRpD6gRm1cb7Khm8VEpZ3NT3nPK7uTX6Fq/go-reuseport/poll"
)

var (
	ErrUnsupportedMode error = errors.New("only 'w' mode is supported on this arch")
)

func PollPark(ctx context.Context, fd int, mode string) error {
	if mode != "w" {
		return ErrUnsupportedMode
	}

	p, err := poll.New(fd)
	if err != nil {
		return err
	}
	defer p.Close()

	return p.WaitWriteCtx(ctx)
}
