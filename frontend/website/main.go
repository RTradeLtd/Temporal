package main

import (
	"bufio"
	"context"
	"fmt"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"io/ioutil"
	"log"
	"os"

	host "gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"

	net "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"

	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"

	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"

	libp2p "github.com/libp2p/go-libp2p"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

var (
	listenAddress       = "/ip4/127.0.0.1/tcp/9090"
	listenAddressClient = "/ip4/127.0.0.1/tcp/9091"
)

func main() {
	runMode := os.Getenv("RUN_MODE")
	switch runMode {
	case "server":
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pk, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
		if err != nil {
			log.Fatal(err)
		}

		host, err := libp2p.New(ctx, libp2p.Identity(pk), libp2p.ListenAddrStrings(listenAddress))
		if err != nil {
			log.Fatal(err)
		}

		listenAddressFormatted := fmt.Sprintf("/ipfs/%s", host.ID().Pretty())
		hostAddr, err := ma.NewMultiaddr(listenAddressFormatted)
		if err != nil {
			log.Fatal(err)
		}
		address := host.Addrs()[0]
		fullAddress := address.Encapsulate(hostAddr)
		regStream(host)
		fmt.Println(fullAddress)
		select {}
	case "client":
		if len(os.Args) < 4 || len(os.Args) > 4 {
			log.Fatal("not enough arguments")
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		target := os.Args[1]
		port := os.Args[2]
		peerString := os.Args[3]

		targetAddress := fmt.Sprintf("/ip4/%s/tcp/%s/ipfs/%s", target, port, peerString)
		pk, _, err := crypto.GenerateKeyPair(crypto.RSA, 1024)
		if err != nil {
			log.Fatal(err)
		}
		host, err := libp2p.New(ctx, libp2p.Identity(pk), libp2p.ListenAddrStrings(listenAddressClient))
		if err != nil {
			log.Fatal(err)
		}
		regStream(host)
		//listenAddressClientFormatted := fmt.Sprintf("/ipfs/%s", host.ID().Pretty())
		ipfsAddr, err := ma.NewMultiaddr(targetAddress)
		pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatal(err)
		}
		peerID, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatal(err)
		}
		targetPeerAddr, err := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerID)),
		)
		targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)
		host.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)
		stream, err := host.NewStream(context.Background(), peerID, "/echo/1.0.0")
		if err != nil {
			log.Fatal(err)
		}

		_, err = stream.Write([]byte("Hello, world!\n"))
		if err != nil {
			log.Fatalln(err)
		}

		out, err := ioutil.ReadAll(stream)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("read reply: %q\n", out)
	default:
		fmt.Println("bork")
	}
}

func regStream(ha host.Host) {
	ha.SetStreamHandler("/echo/1.0.0", func(s net.Stream) {
		log.Println("new stream")
		if err := doEcho(s); err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})
}

// doEcho reads a line of data a stream and writes it back
func doEcho(s net.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("read: %s\n", str)
	_, err = s.Write([]byte(str))
	return err
}
