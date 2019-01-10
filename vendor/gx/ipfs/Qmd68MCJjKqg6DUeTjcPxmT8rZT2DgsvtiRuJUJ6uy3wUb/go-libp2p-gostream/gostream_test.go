package gostream

import (
	"bufio"
	"context"
	"testing"
	"time"

	multiaddr "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	peerstore "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore"
	libp2p "gx/ipfs/QmYxivS34F2M2n44WQQnRHGAKS8aoRUxwGpi9wk4Cdn4Jf/go-libp2p"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	host "gx/ipfs/QmaoXrM4Z41PD48JY36YqQGKQpLGjyLA2cKcLsES7YddAq/go-libp2p-host"
)

// newHost illustrates how to build a libp2p host with secio using
// a randomly generated key-pair
func newHost(t *testing.T, listen multiaddr.Multiaddr) host.Host {
	h, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(listen),
	)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestServerClient(t *testing.T) {
	m1, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10000")
	m2, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10001")
	srvHost := newHost(t, m1)
	clientHost := newHost(t, m2)
	defer srvHost.Close()
	defer clientHost.Close()

	srvHost.Peerstore().AddAddrs(clientHost.ID(), clientHost.Addrs(), peerstore.PermanentAddrTTL)
	clientHost.Peerstore().AddAddrs(srvHost.ID(), srvHost.Addrs(), peerstore.PermanentAddrTTL)

	var tag protocol.ID = "/testitytest"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		listener, err := Listen(srvHost, tag)
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()

		if listener.Addr().String() != srvHost.ID().Pretty() {
			t.Fatal("bad listener address")
		}

		servConn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer servConn.Close()

		reader := bufio.NewReader(servConn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				t.Fatal(err)
			}
			if string(msg) != "is libp2p awesome?\n" {
				t.Fatalf("Bad incoming message: %s", msg)
			}

			_, err = servConn.Write([]byte("yes it is\n"))
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	clientConn, err := Dial(clientHost, srvHost.ID(), tag)
	if err != nil {
		t.Fatal(err)
	}

	if clientConn.LocalAddr().String() != clientHost.ID().Pretty() {
		t.Fatal("Bad LocalAddr")
	}

	if clientConn.RemoteAddr().String() != srvHost.ID().Pretty() {
		t.Fatal("Bad RemoteAddr")
	}

	if clientConn.LocalAddr().Network() != Network {
		t.Fatal("Bad Network()")
	}

	err = clientConn.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	err = clientConn.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	err = clientConn.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, err = clientConn.Write([]byte("is libp2p awesome?\n"))
	if err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(clientConn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	if string(resp) != "yes it is\n" {
		t.Errorf("Bad response: %s", resp)
	}

	err = clientConn.Close()
	if err != nil {
		t.Fatal(err)
	}
}
