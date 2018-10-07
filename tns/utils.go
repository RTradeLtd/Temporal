package tns

import (
	"context"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	ci "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
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
