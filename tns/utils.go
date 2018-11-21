package tns

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/RTradeLtd/rtfs"
	libp2p "github.com/libp2p/go-libp2p"
	ci "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
)

// MakeHost is used to generate the libp2p connection for our TNS daemon
func makeHost(pk ci.PrivKey, opts *HostOpts, client bool) (host.Host, error) {
	if opts == nil && !client {
		opts = &HostOpts{
			IPAddress: "0.0.0.0",
			Port:      "9999",
			IPVersion: "ip4",
			Protocol:  "tcp",
		}
	} else if opts == nil && client {
		opts = &HostOpts{
			IPAddress: "0.0.0.0",
			Port:      "9998",
			IPVersion: "ip4",
			Protocol:  "tcp",
		}
	}
	url := fmt.Sprintf(
		"/%s/%s/%s/%s",
		opts.IPVersion,
		opts.IPAddress,
		opts.Protocol,
		opts.Port,
	)
	host, err := libp2p.New(
		context.Background(),
		libp2p.Identity(pk),
		libp2p.ListenAddrStrings(url),
	)
	if err != nil {
		return nil, err
	}
	return host, nil
}

// GenerateStreamAndWrite is a helper function used to generate, and interact with a stream
func (c *Client) GenerateStreamAndWrite(ctx context.Context, peerID peer.ID, cmd, ipfsAPI string, reqBytes []byte) (interface{}, error) {
	var (
		s    inet.Stream
		intf interface{}
		err  error
	)
	switch cmd {
	case "record-record":
		s, err = c.Host.NewStream(ctx, peerID, CommandRecordRequest)
	case "zone-request":
		s, err = c.Host.NewStream(ctx, peerID, CommandZoneRequest)
	case "echo":
		s, err = c.Host.NewStream(ctx, peerID, CommandEcho)
	default:
		return nil, errors.New("unsupported command")
	}
	if err != nil {
		return nil, err
	}
	// send a message
	_, err = s.Write(append(reqBytes, '\n'))
	if err != nil {
		return nil, err
	}
	// read the message
	resp, err := ioutil.ReadAll(s)
	if err != nil {
		return nil, err
	}
	if cmd == "echo" {
		return string(resp), nil
	}
	rtfsManager, err := rtfs.NewManager(ipfsAPI, nil, time.Minute*10)
	if err != nil {
		return nil, err
	}
	if err = rtfsManager.DagGet(string(resp), &intf); err != nil {
		return nil, err
	}
	return intf, nil
}
