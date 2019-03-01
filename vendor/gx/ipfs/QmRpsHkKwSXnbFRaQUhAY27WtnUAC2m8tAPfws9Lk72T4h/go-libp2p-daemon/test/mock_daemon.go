package test

import (
	"net"
	"testing"
	"time"

	"gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/p2pclient"
	pb "gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/pb"
	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	manet "gx/ipfs/Qmc85NSvmSG4Frn9Vb2cBc1rMyULH6D3TNVEfCzSKoUpip/go-multiaddr-net"
	ggio "gx/ipfs/QmddjPSGZb3ieihSseFeCfVRpZzcqczPNsD2DvarSwnjJB/gogo-protobuf/io"
	"gx/ipfs/QmddjPSGZb3ieihSseFeCfVRpZzcqczPNsD2DvarSwnjJB/gogo-protobuf/proto"
)

type mockDaemon struct {
	clientMaddr ma.Multiaddr
	listener    manet.Listener
}

func newMockDaemon(t testing.TB, listenMaddr, clientMaddr ma.Multiaddr) *mockDaemon {
	listener, err := manet.Listen(listenMaddr)
	if err != nil {
		t.Fatalf("listening on maddr in mock daemon: %s", err)
	}
	return &mockDaemon{
		clientMaddr: clientMaddr,
		listener:    listener,
	}
}

func (d *mockDaemon) Close() error {
	return d.listener.Close()
}

const testTimeout = time.Second

type mockConn struct {
	net.Conn
	r ggio.ReadCloser
	w ggio.WriteCloser
}

func (d *mockDaemon) ExpectConn(t testing.TB) *mockConn {
	timeoutc := make(chan struct{}, 1)
	go func() {
		select {
		case <-time.After(testTimeout):
			d.Close()
			t.Fatalf("timeout waiting for connection")
		case <-timeoutc:
			return
		}
	}()

	conn, err := d.listener.Accept()
	timeoutc <- struct{}{}

	if err != nil {
		t.Fatalf("accepting incoming connection: %s", err)
	}

	r := ggio.NewDelimitedReader(conn, p2pclient.MessageSizeMax)
	w := ggio.NewDelimitedWriter(conn)

	return &mockConn{
		Conn: conn,
		r:    r,
		w:    w,
	}
}

func (c *mockConn) ExpectRequestType(t testing.TB, typ pb.Request_Type) *pb.Request {
	req := &pb.Request{}
	if err := c.r.ReadMsg(req); err != nil {
		t.Fatalf("reading message: %s", err)
	}

	if req.GetType() != typ {
		t.Fatalf("request type %s did not match expected %s", req.GetType().String(), typ.String())
	}

	return req
}

func (c *mockConn) ExpectDHTRequestType(t testing.TB, typ pb.DHTRequest_Type) *pb.DHTRequest {
	req := &pb.Request{}
	if err := c.r.ReadMsg(req); err != nil {
		t.Fatalf("reading message: %s", err)
	}

	if req.GetType() != pb.Request_DHT {
		t.Fatalf("request type %s did not match expected %s", req.GetType().String(), typ.String())
	}

	if req.Dht == nil {
		t.Fatal("dht request was not populated")
	}

	return req.Dht
}

func (c *mockConn) SendMessage(t testing.TB, mes proto.Message) {
	err := c.w.WriteMsg(mes)
	if err != nil {
		t.Fatalf("sending message: %s", err)
	}
}

func streamStart() *pb.Response {
	return &pb.Response{
		Type: pb.Response_OK.Enum(),
		Dht: &pb.DHTResponse{
			Type: pb.DHTResponse_BEGIN.Enum(),
		},
	}
}

func streamEnd() *pb.DHTResponse {
	return &pb.DHTResponse{
		Type: pb.DHTResponse_END.Enum(),
	}
}

func wrapResponseStream(resps []*pb.DHTResponse) []proto.Message {
	respStream := make([]proto.Message, len(resps)+2)
	respStream[0] = streamStart()
	i := 1
	for _, resp := range resps {
		respStream[i] = resp
		i++
	}
	respStream[len(resps)+1] = streamEnd()
	return respStream
}

func (c *mockConn) SendStreamAsync(t testing.TB, resps []*pb.DHTResponse) {
	go func() {
		messages := wrapResponseStream(resps)
		for _, mes := range messages {
			if err := c.w.WriteMsg(mes); err != nil {
				t.Fatalf("writing stream: %s", err)
			}
		}
	}()
}
