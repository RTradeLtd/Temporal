package tns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	defaultZoneName           = "myzone"
	defaultZoneManagerKeyName = "postables-3072"
	defaultZoneKeyName        = "postables-testkeydemo"
	defaultZoneUserName       = "postables"
	defaultRecordName         = "myrecord"
	defaultRecordKeyName      = "postables-testkeydemo2"
	defaultRecordUserName     = "postables"
	dev                       = true
)

// GenerateTNSClient is used to generate a TNS Client
func GenerateTNSClient(genPK bool, pk ci.PrivKey) (*Client, error) {
	var (
		privateKey ci.PrivKey
		err        error
	)
	// allow the client to provide the crytographic identity to be used, or generate one
	if genPK {
		privateKey, _, err = ci.GenerateKeyPair(ci.Ed25519, 256)
		if err != nil {
			return nil, err
		}
	} else {
		privateKey = pk
	}
	return &Client{
		PrivateKey: privateKey,
	}, nil
}

// QueryTNS is used to query a peer for TNS name resolution
func (c *Client) QueryTNS(peerID peer.ID, cmd string, requestArgs interface{}) (interface{}, error) {
	switch cmd {
	case "echo":
		// send a basic echo test
		return c.queryEcho(peerID)
	case "zone-request":
		// ensure the request argument is of type zone request
		args := requestArgs.(ZoneRequest)
		return c.ZoneRequest(peerID, &args)
	case "record-request":
		args := requestArgs.(RecordRequest)
		return c.RecordRequest(peerID, &args)
	default:
		return nil, errors.New("unsupported cmd")
	}
}

// ZoneRequest is a call used to request a zone from TNS
func (c *Client) ZoneRequest(peerID peer.ID, req *ZoneRequest) (interface{}, error) {
	if req == nil {
		req = &ZoneRequest{
			ZoneName:           defaultZoneName,
			ZoneManagerKeyName: defaultZoneManagerKeyName,
			UserName:           defaultZoneUserName,
		}
	}
	marshaledData, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	return c.GenerateStreamAndWrite(
		context.Background(), peerID, "zone-request", c.IPFSAPI, marshaledData,
	)
}

// RecordRequest is a call used to request a record from TNS
func (c *Client) RecordRequest(peerID peer.ID, req *RecordRequest) (interface{}, error) {
	if req == nil {
		req = &RecordRequest{
			RecordName: defaultRecordName,
			UserName:   defaultRecordUserName,
		}
	}
	marshaledData, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	return c.GenerateStreamAndWrite(
		context.Background(), peerID, "record-request", c.IPFSAPI, marshaledData,
	)
}

func (c *Client) queryEcho(peerID peer.ID) (interface{}, error) {
	return c.GenerateStreamAndWrite(
		context.Background(), peerID, "echo", c.IPFSAPI, []byte("test\n"),
	)
}

// AddPeerToPeerStore is used to add a TNS node to our peer store list
func (c *Client) AddPeerToPeerStore(peerAddr string) (peer.ID, error) {
	// generate a multiformat address to connect to
	// /ip4/192.168.1.101/tcp/9999/ipfs/QmbtKadk9x6s56Wh226Wu84ZUc7xEe7AFgvm9bYUbrENDM
	ipfsaddr, err := ma.NewMultiaddr(peerAddr)
	if err != nil {
		return "", err
	}
	// extract the ipfs peer id for the node
	// QmbtKadk9x6s56Wh226Wu84ZUc7xEe7AFgvm9bYUbrENDM
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return "", err
	}
	// decode the peerid
	// <peer.ID Qm*brENDM>
	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		return "", err
	}
	// generate an ipfs based peer address address that we connect to
	// /ipfs/QmbtKadk9x6s56Wh226Wu84ZUc7xEe7AFgvm9bYUbrENDM
	targetPeerAddr, err := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", pid),
	)
	if err != nil {
		return "", err
	}
	// generate a basic multiformat ip address to connect to
	// /ip4/192.168.1.101/tcp/9999
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
	// add a properly formatted libp2p address to connect to
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
