package test

import (
	"net"
	"os"
	"testing"
	"time"

	"gx/ipfs/QmdoKBy1K7Dm59KyuFE6Gsmcq73zS9dscrzVwaoSGDyxF7/go-libp2p-daemon/p2pclient"
	pb "gx/ipfs/QmdoKBy1K7Dm59KyuFE6Gsmcq73zS9dscrzVwaoSGDyxF7/go-libp2p-daemon/pb"
	ggio "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/io"
	"gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/proto"
)

type mockdaemon struct {
	clientPath string
	listener   net.Listener
}

func newMockDaemon(t testing.TB, listenPath, clientPath string) *mockdaemon {
	_, err := os.Stat(clientPath)
	if err != nil {
		t.Fatalf("searching for client socket in mock daemon: %s", err)
	}

	listener, err := net.Listen("unix", listenPath)
	if err != nil {
		t.Fatalf("listening on unix domain socket in mock daemon: %s", err)
	}

	return &mockdaemon{
		clientPath: clientPath,
		listener:   listener,
	}
}

func (d *mockdaemon) Close() error {
	return d.listener.Close()
}

const testTimeout = time.Second

type mockconn struct {
	net.Conn
	r ggio.ReadCloser
	w ggio.WriteCloser
}

func (d *mockdaemon) ExpectConn(t testing.TB) *mockconn {
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

	return &mockconn{
		Conn: conn,
		r:    r,
		w:    w,
	}
}

func (c *mockconn) ExpectRequestType(t testing.TB, typ pb.Request_Type) *pb.Request {
	req := &pb.Request{}
	if err := c.r.ReadMsg(req); err != nil {
		t.Fatalf("reading message: %s", err)
	}

	if req.GetType() != typ {
		t.Fatalf("request type %s did not match expected %s", req.GetType().String(), typ.String())
	}

	return req
}

func (c *mockconn) ExpectDHTRequestType(t testing.TB, typ pb.DHTRequest_Type) *pb.DHTRequest {
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

func (c *mockconn) SendMessage(t testing.TB, mes proto.Message) {
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

func (c *mockconn) SendStreamAsync(t testing.TB, resps []*pb.DHTResponse) {
	go func() {
		messages := wrapResponseStream(resps)
		for _, mes := range messages {
			if err := c.w.WriteMsg(mes); err != nil {
				t.Fatalf("writing stream: %s", err)
			}
		}
	}()
}
