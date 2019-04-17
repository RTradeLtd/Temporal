package autonat

import (
	pb "github.com/libp2p/go-libp2p-autonat/pb"

	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("autonat-svc")

func newDialResponseOK(addr ma.Multiaddr) *pb.Message_DialResponse {
	dr := new(pb.Message_DialResponse)
	dr.Status = pb.Message_OK.Enum()
	dr.Addr = addr.Bytes()
	return dr
}

func newDialResponseError(status pb.Message_ResponseStatus, text string) *pb.Message_DialResponse {
	dr := new(pb.Message_DialResponse)
	dr.Status = status.Enum()
	dr.StatusText = &text
	return dr
}
