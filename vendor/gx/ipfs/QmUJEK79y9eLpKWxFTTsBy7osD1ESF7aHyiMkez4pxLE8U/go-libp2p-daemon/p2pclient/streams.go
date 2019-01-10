package p2pclient

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	pb "gx/ipfs/QmUJEK79y9eLpKWxFTTsBy7osD1ESF7aHyiMkez4pxLE8U/go-libp2p-daemon/pb"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	manet "gx/ipfs/QmZcLBXKaFe8ND5YHPkJRAwmhJGrVsi1JqDZNyJ4nRK5Mj/go-multiaddr-net"
	ggio "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/io"
	proto "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/proto"
)

// StreamInfo wraps the protobuf structure with friendlier types.
type StreamInfo struct {
	Peer  peer.ID
	Addr  ma.Multiaddr
	Proto string
}

func convertStreamInfo(info *pb.StreamInfo) (*StreamInfo, error) {
	id, err := peer.IDFromBytes(info.Peer)
	if err != nil {
		return nil, err
	}
	addr, err := ma.NewMultiaddrBytes(info.Addr)
	if err != nil {
		return nil, err
	}
	streamInfo := &StreamInfo{
		Peer:  id,
		Addr:  addr,
		Proto: info.GetProto(),
	}
	return streamInfo, nil
}

type byteReaderConn struct {
	net.Conn
}

func (c *byteReaderConn) ReadByte() (byte, error) {
	b := make([]byte, 1)
	_, err := c.Read(b)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func readMsgBytesSafe(r *byteReaderConn) (*bytes.Buffer, error) {
	length, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}
	n, err := io.CopyN(out, r, int64(length))
	if err != nil {
		return nil, err
	}
	if n != int64(length) {
		return nil, fmt.Errorf("read incorrect number of bytes in header: expected %d, got %d", length, n)
	}
	return out, nil
}

func readMsgSafe(c *byteReaderConn, msg proto.Message) error {
	header, err := readMsgBytesSafe(c)
	if err != nil {
		return err
	}

	r := ggio.NewFullReader(header, MessageSizeMax)
	if err = r.ReadMsg(msg); err != nil {
		return err
	}

	return nil
}

// NewStream initializes a new stream on one of the protocols in protos with
// the specified peer.
func (c *Client) NewStream(peer peer.ID, protos []string) (*StreamInfo, io.ReadWriteCloser, error) {
	controlconn, err := c.newControlConn()
	if err != nil {
		return nil, nil, err
	}
	control := &byteReaderConn{controlconn}
	w := ggio.NewDelimitedWriter(control)

	req := &pb.Request{
		Type: pb.Request_STREAM_OPEN.Enum(),
		StreamOpen: &pb.StreamOpenRequest{
			Peer:  []byte(peer),
			Proto: protos,
		},
	}

	if err = w.WriteMsg(req); err != nil {
		control.Close()
		return nil, nil, err
	}

	resp := &pb.Response{}
	err = readMsgSafe(control, resp)
	if err != nil {
		control.Close()
		return nil, nil, err
	}
	if err := resp.GetError(); err != nil {
		return nil, nil, fmt.Errorf("error from daemon: %s", err.GetMsg())
	}
	info, err := convertStreamInfo(resp.GetStreamInfo())
	if err != nil {
		return nil, nil, fmt.Errorf("parsing stream info: %s", err)
	}

	return info, control, nil
}

// Close stops the listener address.
func (c *Client) Close() error {
	if c.listener != nil {
		err := c.listener.Close()
		return err
	}
	return nil
}

func (c *Client) streamDispatcher() {
	for {
		rawconn, err := c.listener.Accept()
		if err != nil {
			log.Warningf("accepting incoming connection: %s", err)
			return
		}
		conn := &byteReaderConn{rawconn}

		info := &pb.StreamInfo{}
		err = readMsgSafe(conn, info)
		if err != nil {
			log.Errorf("error reading stream info: %s", err)
			conn.Close()
			continue
		}
		streamInfo, err := convertStreamInfo(info)
		if err != nil {
			log.Errorf("error parsing stream info: %s", err)
			conn.Close()
			continue
		}

		c.mhandlers.Lock()
		defer c.mhandlers.Unlock()
		handler, ok := c.handlers[streamInfo.Proto]
		if !ok {
			conn.Close()
			continue
		}

		go handler(streamInfo, conn)
	}
}

func (c *Client) listen() error {
	l, err := manet.Listen(c.listenMaddr)
	if err != nil {
		return err
	}

	c.listener = l
	go c.streamDispatcher()

	return nil
}

// StreamHandlerFunc is the type of callbacks executed upon receiving a new stream
// on a given protocol.
type StreamHandlerFunc func(*StreamInfo, io.ReadWriteCloser)

// NewStreamHandler establishes an inbound multi-address and starts a listener.
// All inbound connections to the listener are delegated to the provided
// handler.
func (c *Client) NewStreamHandler(protos []string, handler StreamHandlerFunc) error {
	control, err := c.newControlConn()
	if err != nil {
		return err
	}

	c.mhandlers.Lock()
	defer c.mhandlers.Unlock()

	w := ggio.NewDelimitedWriter(control)
	req := &pb.Request{
		Type: pb.Request_STREAM_HANDLER.Enum(),
		StreamHandler: &pb.StreamHandlerRequest{
			Addr:  c.listenMaddr.Bytes(),
			Proto: protos,
		},
	}
	if err := w.WriteMsg(req); err != nil {
		return err
	}

	for _, proto := range protos {
		c.handlers[proto] = handler
	}

	return nil
}
