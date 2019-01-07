package p2pd

import (
	"context"
	"time"

	pb "gx/ipfs/QmdoKBy1K7Dm59KyuFE6Gsmcq73zS9dscrzVwaoSGDyxF7/go-libp2p-daemon/pb"

	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
)

func (d *Daemon) doConnManager(req *pb.Request) *pb.Response {
	if req.ConnManager == nil {
		return errorResponseString("Malformed request; missing parameters")
	}

	switch *req.ConnManager.Type {
	case pb.ConnManagerRequest_TAG_PEER:
		p, err := peer.IDFromBytes(req.ConnManager.GetPeer())
		if err != nil {
			return errorResponse(err)
		}

		tag := req.ConnManager.GetTag()
		if tag == "" {
			return errorResponseString("Malformed request; missing tag parameter")
		}
		weight := req.ConnManager.GetWeight()

		d.host.ConnManager().TagPeer(p, tag, int(weight))
		return okResponse()

	case pb.ConnManagerRequest_UNTAG_PEER:
		p, err := peer.IDFromBytes(req.ConnManager.GetPeer())
		if err != nil {
			return errorResponse(err)
		}

		tag := req.ConnManager.GetTag()
		if tag == "" {
			return errorResponseString("Malformed request; missing tag parameter")
		}

		d.host.ConnManager().UntagPeer(p, tag)
		return okResponse()

	case pb.ConnManagerRequest_TRIM:
		ctx, cancel := context.WithTimeout(d.ctx, 60*time.Second)
		defer cancel()

		d.host.ConnManager().TrimOpenConns(ctx)
		return okResponse()

	default:
		log.Debugf("Unexpected ConnManager request type: %d", *req.ConnManager.Type)
		return errorResponseString("Unexpected request")
	}
}
