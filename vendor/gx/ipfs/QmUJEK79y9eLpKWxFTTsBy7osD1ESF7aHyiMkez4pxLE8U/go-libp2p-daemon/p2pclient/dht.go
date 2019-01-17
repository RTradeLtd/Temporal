package p2pclient

import (
	"context"
	"errors"
	"fmt"
	"net"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	crypto "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	pb "gx/ipfs/QmUJEK79y9eLpKWxFTTsBy7osD1ESF7aHyiMkez4pxLE8U/go-libp2p-daemon/pb"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	ggio "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/io"
)

// PeerInfo wraps the PeerInfo message from our protobuf with richer types.
type PeerInfo struct {
	// Id is the peer's ID
	ID peer.ID
	// Addrs are the peer's listen addresses.
	Addrs []ma.Multiaddr
}

func convertPbPeerInfo(pbi *pb.PeerInfo) (PeerInfo, error) {
	if pbi == nil {
		return PeerInfo{}, errors.New("null peerinfo")
	}

	id, err := peer.IDFromBytes(pbi.GetId())
	if err != nil {
		return PeerInfo{}, err
	}

	addrs := make([]ma.Multiaddr, 0, len(pbi.Addrs))
	for _, addrbytes := range pbi.Addrs {
		addr, err := ma.NewMultiaddrBytes(addrbytes)
		if err != nil {
			log.Errorf("error parsing multiaddr in peerinfo: %s", err)
			continue
		}
		addrs = append(addrs, addr)
	}

	pi := PeerInfo{
		ID:    id,
		Addrs: addrs,
	}

	return pi, nil
}

func newDHTReq(req *pb.DHTRequest) *pb.Request {
	return &pb.Request{
		Type: pb.Request_DHT.Enum(),
		Dht:  req,
	}
}

func readDhtResponseStream(ctx context.Context, control net.Conn) (<-chan *pb.DHTResponse, error) {
	r := ggio.NewDelimitedReader(control, MessageSizeMax)
	msg := &pb.Response{}
	if err := r.ReadMsg(msg); err != nil {
		return nil, err
	}
	if msg.GetType() != pb.Response_OK {
		return nil, errors.New(msg.GetError().GetMsg())
	}
	if msg.Dht.GetType() != pb.DHTResponse_BEGIN {
		return nil, fmt.Errorf("expected a stream BEGIN message but got %s", msg.Dht.GetType().String())
	}

	out := make(chan *pb.DHTResponse)
	go func() {
		defer close(out)
		defer control.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg := &pb.DHTResponse{}
				if err := r.ReadMsg(msg); err != nil {
					log.Errorf("reading daemon response: %s", err)
					return
				}

				if msg.GetType() == pb.DHTResponse_END {
					return
				}

				out <- msg
			}
		}
	}()

	return out, nil
}

// doDHT issues a request to the daemon and returns its DHTResponse.
func (c *Client) doDHT(dhtReq *pb.DHTRequest) (*pb.DHTResponse, error) {
	control, err := c.newControlConn()
	if err != nil {
		return nil, err
	}
	defer control.Close()

	w := ggio.NewDelimitedWriter(control)
	req := newDHTReq(dhtReq)
	if err = w.WriteMsg(req); err != nil {
		return nil, err
	}

	r := ggio.NewDelimitedReader(control, MessageSizeMax)
	msg := &pb.Response{}
	if err = r.ReadMsg(msg); err != nil {
		return nil, err
	}

	if msg.GetType() == pb.Response_ERROR {
		err := fmt.Errorf("error from daemon in %s response: %s", req.GetType().String(), msg.GetError())
		log.Errorf(err.Error())
		return nil, err
	}

	return msg.GetDht(), nil
}

// doDHTNonNil issues a request to the daemon and returns its DHTResponse, ensuring
// it's not nil and returning an error when it is. This is a convenience
// function.
func (c *Client) doDHTNonNil(req *pb.DHTRequest) (*pb.DHTResponse, error) {
	resp, err := c.doDHT(req)
	if err == nil && resp == nil {
		return nil, fmt.Errorf("dht response was not populated in %s request", req.GetType().String())
	}
	return resp, err
}

// FindPeer queries the daemon for a peer's address.
func (c *Client) FindPeer(peer peer.ID) (PeerInfo, error) {
	req := &pb.DHTRequest{
		Type: pb.DHTRequest_FIND_PEER.Enum(),
		Peer: []byte(peer),
	}

	msg, err := c.doDHTNonNil(req)
	if err != nil {
		return PeerInfo{}, err
	}

	info, err := convertPbPeerInfo(msg.GetPeer())
	if err != nil {
		return PeerInfo{}, err
	}

	return info, nil
}

// GetPublicKey queries the daemon for a peer's address.
func (c *Client) GetPublicKey(peer peer.ID) (crypto.PubKey, error) {
	req := &pb.DHTRequest{
		Type: pb.DHTRequest_GET_PUBLIC_KEY.Enum(),
		Peer: []byte(peer),
	}

	msg, err := c.doDHTNonNil(req)
	if err != nil {
		return nil, err
	}

	key, err := crypto.UnmarshalPublicKey(msg.GetValue())
	if err != nil {
		return nil, err
	}

	return key, nil
}

// GetValue queries the daemon for a value stored at a key.
func (c *Client) GetValue(key []byte) ([]byte, error) {
	req := &pb.DHTRequest{
		Type: pb.DHTRequest_GET_VALUE.Enum(),
		Key:  key,
	}

	msg, err := c.doDHTNonNil(req)
	if err != nil {
		return nil, err
	}

	return msg.GetValue(), nil
}

// PutValue sets the value stored at a given key in the DHT to a given value.
func (c *Client) PutValue(key []byte, value []byte) error {
	req := &pb.DHTRequest{
		Type:  pb.DHTRequest_PUT_VALUE.Enum(),
		Key:   key,
		Value: value,
	}

	_, err := c.doDHT(req)
	return err
}

// Provide announces that our peer provides content described by a CID.
func (c *Client) Provide(id cid.Cid) error {
	req := &pb.DHTRequest{
		Type: pb.DHTRequest_PROVIDE.Enum(),
		Cid:  id.Bytes(),
	}

	_, err := c.doDHT(req)
	return err
}

func convertResponseToPeerInfo(respc <-chan *pb.DHTResponse) <-chan PeerInfo {
	out := make(chan PeerInfo, 10)

	go func() {
		defer close(out)

		for resp := range respc {
			info, err := convertPbPeerInfo(resp.GetPeer())
			if err != nil {
				log.Errorf("error converting peerinfo: %s", err.Error())
				continue
			}

			out <- info
		}
	}()

	return out
}

func convertResponseToPeerID(respc <-chan *pb.DHTResponse) <-chan peer.ID {
	out := make(chan peer.ID, 10)

	go func() {
		defer close(out)

		for resp := range respc {
			id, err := peer.IDFromBytes(resp.GetValue())
			if err != nil {
				log.Errorf("error parsing peer id: %s", err.Error())
				continue
			}

			out <- id
		}
	}()

	return out
}

func convertResponseToValue(respc <-chan *pb.DHTResponse) <-chan []byte {
	out := make(chan []byte, 10)

	go func() {
		defer close(out)

		for resp := range respc {
			out <- resp.GetValue()
		}
	}()

	return out
}

func (c *Client) streamRequest(ctx context.Context, req *pb.Request) (<-chan *pb.DHTResponse, error) {
	control, err := c.newControlConn()
	if err != nil {
		return nil, err
	}

	w := ggio.NewDelimitedWriter(control)
	if err = w.WriteMsg(req); err != nil {
		return nil, err
	}

	return readDhtResponseStream(ctx, control)
}

func (c *Client) streamRequestPeerInfo(ctx context.Context, req *pb.Request) (<-chan PeerInfo, error) {
	respc, err := c.streamRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	out := convertResponseToPeerInfo(respc)
	return out, nil
}

func (c *Client) streamRequestPeerID(ctx context.Context, req *pb.Request) (<-chan peer.ID, error) {
	respc, err := c.streamRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	out := convertResponseToPeerID(respc)
	return out, nil
}

func (c *Client) streamRequestValue(ctx context.Context, req *pb.Request) (<-chan []byte, error) {
	respc, err := c.streamRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	out := convertResponseToValue(respc)
	return out, nil
}

// FindPeersConnectedToPeer queries the DHT for peers that have an active
// connection to a given peer.
func (c *Client) FindPeersConnectedToPeer(ctx context.Context, peer peer.ID) (<-chan PeerInfo, error) {
	req := newDHTReq(&pb.DHTRequest{
		Type: pb.DHTRequest_FIND_PEERS_CONNECTED_TO_PEER.Enum(),
		Peer: []byte(peer),
	})

	return c.streamRequestPeerInfo(ctx, req)
}

// FindProviders queries the DHT for peers that provide a piece of content
// identified by a CID.
func (c *Client) FindProviders(ctx context.Context, cid cid.Cid) (<-chan PeerInfo, error) {
	req := newDHTReq(&pb.DHTRequest{
		Type: pb.DHTRequest_FIND_PROVIDERS.Enum(),
		Cid:  cid.Bytes(),
	})

	return c.streamRequestPeerInfo(ctx, req)
}

// GetClosestPeers queries the DHT routing table for peers that are closest
// to a provided key.
func (c *Client) GetClosestPeers(ctx context.Context, key []byte) (<-chan peer.ID, error) {
	req := newDHTReq(&pb.DHTRequest{
		Type: pb.DHTRequest_GET_CLOSEST_PEERS.Enum(),
		Key:  key,
	})

	return c.streamRequestPeerID(ctx, req)
}

// SearchValue queries the DHT for the best/most valid value stored at a key.
// Later responses are better.
func (c *Client) SearchValue(ctx context.Context, key []byte) (<-chan []byte, error) {
	req := newDHTReq(&pb.DHTRequest{
		Type: pb.DHTRequest_SEARCH_VALUE.Enum(),
		Key:  key,
	})

	return c.streamRequestValue(ctx, req)
}
