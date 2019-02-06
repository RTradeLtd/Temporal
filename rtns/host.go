package rtns

import (
	"context"
	"encoding/base64"
	"time"

	config "gx/ipfs/QmPEpj17FDRpc7K1aArKZp3RsHtzRMKykeK9GVgn4WQGPR/go-ipfs-config"
	ci "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	path "gx/ipfs/QmT3rzed1ppXefourpmoZ7tyVQfsGPQZ1pHDngLmCvXxd3/go-path"
	peer "gx/ipfs/QmTRhk7cgjUf2gfQ3p2M9KPECNZEW9XUrmHcFCgog4cPgB/go-libp2p-peer"
	ds "gx/ipfs/QmaRb5yNXKonhbkpNxNawoydk4N6es6b4fPj19sjEKsh5D/go-datastore"

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
	return &Publisher{
		host: host,
	}, nil
}

// PublishWithEOL is used to publish an IPNS record with non default lifetime values
func (p *Publisher) PublishWithEOL(ctx context.Context, pk ci.PrivKey, content string, eol time.Time) error {
	return p.host.Namesys.PublishWithEOL(ctx, pk, path.FromString(content), eol)
}
