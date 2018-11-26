package autonat

import (
	pb "gx/ipfs/QmYdJ7pcu5fbwBeiue9TPaNie3QwPaGsh6sPzZiz9gMgLQ/go-libp2p-autonat/pb"

	pstore "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore"
	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"
)

const AutoNATProto = "/libp2p/autonat/1.0.0"

var log = logging.Logger("autonat")

func newDialMessage(pi pstore.PeerInfo) *pb.Message {
	msg := new(pb.Message)
	msg.Type = pb.Message_DIAL.Enum()
	msg.Dial = new(pb.Message_Dial)
	msg.Dial.Peer = new(pb.Message_PeerInfo)
	msg.Dial.Peer.Id = []byte(pi.ID)
	msg.Dial.Peer.Addrs = make([][]byte, len(pi.Addrs))
	for i, addr := range pi.Addrs {
		msg.Dial.Peer.Addrs[i] = addr.Bytes()
	}

	return msg
}
