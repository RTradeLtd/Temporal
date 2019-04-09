package rtns

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	path "gx/ipfs/QmQAgv6Gaoe2tQpcabqwKXKChp2MZ7i3UXv9DqTTaxCaTR/go-path"
	ci "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto"
	config "gx/ipfs/QmUAuYuiafnJRZxDDX7MuruMNsicYNuyub5vUeAcupUBNs/go-ipfs-config"
	ds "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"

	"github.com/ipfs/go-ipfs/core"
	repo "github.com/ipfs/go-ipfs/repo"
)

// Publisher provides a helper to publish IPNS records
type Publisher struct {
	host *core.IpfsNode
}

// Opts is used to configure our connection
type Opts struct {
	PK ci.PrivKey
}

// NewPublisher is used to generate our IPNS publisher
func NewPublisher(pk ci.PrivKey, permanent bool, swarmAddrs ...string) (*Publisher, error) {
	pid, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		return nil, err
	}
	pkBytes, err := pk.Bytes()
	if err != nil {
		return nil, err
	}
	// generate a blank config
	c := config.Config{}
	// popular config with necessary defaults
	c.Bootstrap = config.DefaultBootstrapAddresses
	c.Addresses.Swarm = swarmAddrs
	c.Identity.PeerID = pid.Pretty()
	c.Identity.PrivKey = base64.StdEncoding.EncodeToString(pkBytes)
	// generate a null datastore, as we just want to publish records
	d := ds.NewNullDatastore()
	// create a mock repo to feed into our node
	repoMock := repo.Mock{
		C: c,
		D: d,
	}
	// create a new node
	host, err := core.NewNode(context.Background(), &core.BuildCfg{
		Online:    true,
		Permanent: permanent,
		Repo:      &repoMock,
		// this is used to enable ipns pubsub
		ExtraOpts: map[string]bool{
			"ipnsps": true,
		},
	})
	if err != nil {
		return nil, err
	}
	// print connection information
	fmt.Printf("listening on %v, with pid %s\n", swarmAddrs, pid.Pretty())
	return &Publisher{
		host: host,
	}, nil
}

// PublishWithEOL is used to publish an IPNS record with non default lifetime values
func (p *Publisher) PublishWithEOL(ctx context.Context, pk ci.PrivKey, content string, eol time.Time) error {
	return p.host.Namesys.PublishWithEOL(ctx, pk, path.FromString(content), eol)
}
