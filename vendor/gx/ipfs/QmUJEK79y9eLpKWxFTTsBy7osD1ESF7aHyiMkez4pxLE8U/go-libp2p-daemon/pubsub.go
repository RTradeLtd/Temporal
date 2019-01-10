package p2pd

import (
	pb "gx/ipfs/QmUJEK79y9eLpKWxFTTsBy7osD1ESF7aHyiMkez4pxLE8U/go-libp2p-daemon/pb"

	ps "gx/ipfs/QmVRxA4J3UPQpw74dLrQ6NJkfysCA1H4GU28gVpXQt9zMU/go-libp2p-pubsub"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	ggio "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/io"
)

func (d *Daemon) doPubsub(req *pb.Request) (*pb.Response, *ps.Subscription) {
	if d.pubsub == nil {
		return errorResponseString("PubSub not enabled"), nil
	}

	if req.Pubsub == nil {
		return errorResponseString("Malformed request; missing parameters"), nil
	}

	switch *req.Pubsub.Type {
	case pb.PSRequest_GET_TOPICS:
		return d.doPubsubGetTopics(req.Pubsub)

	case pb.PSRequest_LIST_PEERS:
		return d.doPubsubListPeers(req.Pubsub)

	case pb.PSRequest_PUBLISH:
		return d.doPubsubPublish(req.Pubsub)

	case pb.PSRequest_SUBSCRIBE:
		return d.doPubsubSubscribe(req.Pubsub)

	default:
		log.Debugf("Unexpected pubsub request type: %d", *req.Pubsub.Type)
		return errorResponseString("Unexpected request"), nil
	}
}

func (d *Daemon) doPubsubGetTopics(req *pb.PSRequest) (*pb.Response, *ps.Subscription) {
	topics := d.pubsub.GetTopics()
	return psOkResponse(psResponseTopics(topics)), nil
}

func (d *Daemon) doPubsubListPeers(req *pb.PSRequest) (*pb.Response, *ps.Subscription) {
	if req.Topic == nil {
		return errorResponseString("Malformed request; missing topic parameter"), nil
	}

	peers := d.pubsub.ListPeers(*req.Topic)
	return psOkResponse(psResponsePeers(peers)), nil
}

func (d *Daemon) doPubsubPublish(req *pb.PSRequest) (*pb.Response, *ps.Subscription) {
	if req.Topic == nil {
		return errorResponseString("Malformed request; missing topic parameter"), nil
	}

	err := d.pubsub.Publish(*req.Topic, req.Data)
	if err != nil {
		return errorResponse(err), nil
	}

	return okResponse(), nil
}

func (d *Daemon) doPubsubSubscribe(req *pb.PSRequest) (*pb.Response, *ps.Subscription) {
	if req.Topic == nil {
		return errorResponseString("Malformed request; missing topic parameter"), nil
	}

	sub, err := d.pubsub.Subscribe(*req.Topic)
	if err != nil {
		return errorResponse(err), nil
	}

	return okResponse(), sub
}

func (d *Daemon) doPubsubPipe(sub *ps.Subscription, r ggio.ReadCloser, w ggio.WriteCloser) {
	go func() {
		// read something until the client closes the connection
		// at which point we cancel the subscription
		for {
			var req pb.Request
			err := r.ReadMsg(&req)
			if err != nil {
				sub.Cancel()
				return
			}

			log.Warningf("unexpected message (%s)", req.GetType().String())
		}
	}()

	for {
		msg, err := sub.Next(d.ctx)
		if err != nil {
			log.Warningf("subscription error: %s", err.Error())
			// goroutine will cancel the subscription once the connection is closed on return
			return
		}

		psmsg := psMessage(msg)
		err = w.WriteMsg(psmsg)
		if err != nil {
			log.Warningf("error writing pubsub message: %s", err.Error())
			// goroutine will cancel the subscription once the connection is closed on return
			return
		}
	}
}

func psResponseTopics(topics []string) *pb.PSResponse {
	return &pb.PSResponse{Topics: topics}
}

func psResponsePeers(peers []peer.ID) *pb.PSResponse {
	xpeers := make([][]byte, len(peers))
	for x, p := range peers {
		xpeers[x] = []byte(p)
	}
	return &pb.PSResponse{PeerIDs: xpeers}
}

func psMessage(msg *ps.Message) *pb.PSMessage {
	return &pb.PSMessage{
		From:      msg.From,
		Data:      msg.Data,
		Seqno:     msg.Seqno,
		TopicIDs:  msg.TopicIDs,
		Signature: msg.Signature,
		Key:       msg.Key,
	}
}

func psOkResponse(r *pb.PSResponse) *pb.Response {
	res := okResponse()
	res.Pubsub = r
	return res
}
