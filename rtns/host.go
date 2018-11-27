package rtns

import (
	"context"
	"encoding/base64"

	path "gx/ipfs/QmVi2uUygezqaMTqs3Yzt5FcZFHJoYD4B7jQ2BELjj7ZuY/go-path"

	ci "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	config "gx/ipfs/QmbK4EmM2Xx5fmbqK38TGP3PpY66r3tkXLZTcc7dF9mFwM/go-ipfs-config"
	peer "gx/ipfs/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
	ds "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore"

	"github.com/ipfs/go-ipfs/core"
	repo "github.com/ipfs/go-ipfs/repo"
)

// Publisher provides a helper to publish IPNS records
type Publisher struct {
	host *core.IpfsNode
}

// NewPublisher is used to generate our IPNS publisher
func NewPublisher(swarmAddrs ...string) (*Publisher, error) {
	pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		return nil, err
	}
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
		Permanent: true,
		Repo:      &repoMock,
	})
	if err != nil {
		return nil, err
	}
	return &Publisher{
		host: host,
	}, nil
}

// Publish is used to publish an IPNS record
func (p *Publisher) Publish(ctx context.Context, pk ci.PrivKey, content string) error {
	return p.host.Namesys.Publish(ctx, pk, path.FromString(content))
}
