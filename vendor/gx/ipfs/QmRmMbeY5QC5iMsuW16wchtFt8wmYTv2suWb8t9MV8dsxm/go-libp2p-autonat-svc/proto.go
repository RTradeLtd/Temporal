package autonat

import (
	pb "gx/ipfs/QmXmZtMdQokSodDNvPdhDyaVRAjgybvR8dQtuMsNoWv4Lq/go-libp2p-autonat/pb"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"
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
