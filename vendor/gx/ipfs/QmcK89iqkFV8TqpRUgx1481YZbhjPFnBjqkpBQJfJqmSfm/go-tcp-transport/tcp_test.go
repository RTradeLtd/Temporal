package tcp

import (
	"testing"

	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	utils "gx/ipfs/QmYr9RHifaqHTFZdAsUPLmiMAi2oNeEqA48AFKxXJAsLpJ/go-libp2p-transport/test"
	insecure "gx/ipfs/QmcGgFLHMFLcNMMvxsBC5LeLqubLR5djxjShVU3koVMtVq/go-conn-security/insecure"
	mplex "gx/ipfs/QmdiBZzwGtN2yHJrWD9ojQ7ASS48nv7BcojWLkYd1ZtrV2/go-smux-multiplex"
	tptu "gx/ipfs/QmfNvpHX396fhMeauERV6eFnSJg78rUjhjpFf1JvbjxaYM/go-libp2p-transport-upgrader"
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
