package tcp

import (
	"testing"

	utils "gx/ipfs/QmNQWMWWBmkAcaVEspSNwYB95axzKFhYTdqZtABA2zXoPu/go-libp2p-transport/test"
	insecure "gx/ipfs/QmPRoHqULmP4MuKAN5EFaJ64MLpeMY8cny2318xDBDmmkp/go-conn-security/insecure"
	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	mplex "gx/ipfs/QmaJvNdDccVkTELQLCGXWrLxgaQ14aMdhzZx1EiHPXKbDc/go-smux-multiplex"
	tptu "gx/ipfs/QmeqC5shQjEBRG9B8roZqQCJ9xb7Pq6AbWxJFMyLgqBBWh/go-libp2p-transport-upgrader"
)

func TestTcpTransport(t *testing.T) {
	for i := 0; i < 2; i++ {
		ta := NewTCPTransport(&tptu.Upgrader{
			Secure: insecure.New("peerA"),
			Muxer:  new(mplex.Transport),
		})
		tb := NewTCPTransport(&tptu.Upgrader{
			Secure: insecure.New("peerB"),
			Muxer:  new(mplex.Transport),
		})

		zero := "/ip4/127.0.0.1/tcp/0"
		utils.SubtestTransport(t, ta, tb, zero, "peerA")

		envReuseportVal = false
	}
	envReuseportVal = true
}

func TestTcpTransportCantListenUtp(t *testing.T) {
	for i := 0; i < 2; i++ {
		utpa, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/0/utp")
		if err != nil {
			t.Fatal(err)
		}

		tpt := NewTCPTransport(&tptu.Upgrader{
			Secure: insecure.New("peerB"),
			Muxer:  new(mplex.Transport),
		})

		_, err = tpt.Listen(utpa)
		if err == nil {
			t.Fatal("shouldnt be able to listen on utp addr with tcp transport")
		}

		envReuseportVal = false
	}
	envReuseportVal = true
}
