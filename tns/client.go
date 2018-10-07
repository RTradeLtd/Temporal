package tns

import (
	"context"
	"fmt"
	"io/ioutil"

	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// GenerateTNSClient is used to generate a TNS Client
func GenerateTNSClient(genPK bool, pk ...ci.PrivKey) (*Client, error) {
	var (
		privateKey ci.PrivKey
		err        error
	)
	if genPK {
		privateKey, _, err = ci.GenerateKeyPair(ci.Ed25519, 256)
		if err != nil {
			return nil, err
		}
	} else {
		privateKey = pk[0]
	}
	return &Client{
		PrivateKey: privateKey,
	}, nil
}

// QueryTNS is used to query a peer for TNS name resolution
func (c *Client) QueryTNS(peerID peer.ID) error {
	s, err := c.Host.NewStream(context.Background(), peerID, "/echo/1.0.0")
	if err != nil {
		fmt.Println("failed to generate new stream ", err.Error())
		return err
	}
	resp, err := ioutil.ReadAll(s)
	if err != nil {
		return err
	}
	fmt.Printf("response\n%s", string(resp))
	return nil
}

// AddPeerToPeerStore is used to add a TNS node to our peer store list
func (c *Client) AddPeerToPeerStore(peerAddr string) (peer.ID, error) {
	ipfsaddr, err := ma.NewMultiaddr(peerAddr)
	if err != nil {
		return "", err
	}
	fmt.Println("ipfsaddr ", ipfsaddr)
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return "", err
	}
	fmt.Println("pid ", pid)
	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		return "", err
	}
	fmt.Println("peerid ", peerid)
	targetPeerAddr, err := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)),
	)
	if err != nil {
		return "", err
	}
	fmt.Println("targetPeerAddr ", targetPeerAddr)
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
	c.Host.Peerstore().AddAddr(
		peerid, targetAddr, pstore.PermanentAddrTTL,
	)
	return peerid, nil
}

// MakeHost is used to generate the libp2p connection for our TNS client
func (c *Client) MakeHost(pk ci.PrivKey, opts *HostOpts) error {
	host, err := makeHost(pk, opts, true)
	if err != nil {
		return err
	}
	c.Host = host
	return nil
}
