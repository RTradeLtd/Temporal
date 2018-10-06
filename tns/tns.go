package tns

import (
	"context"
	"errors"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	ci "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	ma "github.com/multiformats/go-multiaddr"
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

// HostMultiAddress is used to get a formatted libp2p host multi address
func (m *Manager) HostMultiAddress() (ma.Multiaddr, error) {
	return ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", m.Host.ID().Pretty()))
}

// ReachableAddress is used to get a reachable address for this host
func (m *Manager) ReachableAddress(addressIndex int) (string, error) {
	if addressIndex > len(m.Host.Addrs()) {
		return "", errors.New("invalid index")
	}
	ipAddr := m.Host.Addrs()[addressIndex]
	multiAddr, err := m.HostMultiAddress()
	if err != nil {
		return "", err
	}
	reachableAddr := ipAddr.Encapsulate(multiAddr)
	return reachableAddr.String(), nil
}
