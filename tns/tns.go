package tns

import (
	ci "github.com/libp2p/go-libp2p-crypto"
)

// Manager is used to manipulate a zone on TNS
type Manager struct {
	PrivateKey        ci.PrivKey
	ZonePrivateKey    ci.PrivKey
	RecordPrivateKeys map[string]ci.PrivKey
	Zone              *Zone
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
