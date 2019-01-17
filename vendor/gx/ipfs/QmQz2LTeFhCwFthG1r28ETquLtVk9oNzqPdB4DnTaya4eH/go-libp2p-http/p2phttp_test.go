package p2phttp

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	multiaddr "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	peerstore "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore"
	libp2p "gx/ipfs/QmYxivS34F2M2n44WQQnRHGAKS8aoRUxwGpi9wk4Cdn4Jf/go-libp2p"
	host "gx/ipfs/QmaoXrM4Z41PD48JY36YqQGKQpLGjyLA2cKcLsES7YddAq/go-libp2p-host"
	gostream "gx/ipfs/Qmd68MCJjKqg6DUeTjcPxmT8rZT2DgsvtiRuJUJ6uy3wUb/go-libp2p-gostream"
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

	listener, err := gostream.Listen(srvHost, "/testiti-test")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			resp := fmt.Sprintf("Hi %s!", body)
			w.Write([]byte(resp))
		})
		server := &http.Server{}
		server.Serve(listener)
	}()

	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", NewTransport(clientHost, ProtocolOption("/testiti-test")))
	client := &http.Client{Transport: tr}

	buf := bytes.NewBufferString("Hector")
	res, err := client.Post(fmt.Sprintf("libp2p://%s/hello", srvHost.ID().Pretty()), "text/plain", buf)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != "Hi Hector!" {
		t.Errorf("expected Hi Hector! but got %s", text)
	}

	t.Log(string(text))
}
