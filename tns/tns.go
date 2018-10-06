package tns

import (
	"context"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	ci "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
)

// Manager is used to manipulate a zone on TNS
type Manager struct {
	PrivateKey        ci.PrivKey
	ZonePrivateKey    ci.PrivKey
	RecordPrivateKeys map[string]ci.PrivKey
	Zone              *Zone
	Host              host.Host
}

// GenerateTNSManager is used to generate a TNS manager for a particular PKI space
func GenerateTNSManager(zoneName string) (*Manager, error) {
	managerPK, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		return nil, err
	}
	zonePK, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		return nil, err
	}
	zoneManager := ZoneManager{
		PublicKey: managerPK.GetPublic(),
	}
	zone := Zone{
		Name:      zoneName,
		PublicKey: zonePK.GetPublic(),
		Manager:   &zoneManager,
	}
	manager := Manager{
		PrivateKey:        managerPK,
		ZonePrivateKey:    zonePK,
		RecordPrivateKeys: nil,
		Zone:              &zone,
	}
	return &manager, nil
}

// MakeHost is used to generate the libp2p connection for our TNS daemon
func (m *Manager) MakeHost(opts *HostOpts) error {
	if opts == nil {
		opts = &HostOpts{
			IPAddress: "0.0.0.0",
			Port:      "9999",
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
		libp2p.Identity(m.PrivateKey),
		libp2p.ListenAddrStrings(url),
	)
	if err != nil {
		return err
	}
	m.Host = host
	return nil
}

// HostAddress is used to generate a formatted libp2p host address
func (m *Manager) HostAddress() string {
	return fmt.Sprintf("/p2p/%s", m.Host.ID().Pretty())
}
