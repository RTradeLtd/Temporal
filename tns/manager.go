package tns

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	ci "github.com/libp2p/go-libp2p-crypto"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

// ManagerOpts defines options for controlling our TNS Manager daemon
type ManagerOpts struct {
	ManagerPK ci.PrivKey `json:"manager_pk"`
	ZonePK    ci.PrivKey `json:"zone_pk"`
	ZoneName  string     `json:"zone_name"`
	DB        *gorm.DB   `json:"db"`
}

// GenerateTNSManager is used to generate a TNS manager for a particular PKI space
func GenerateTNSManager(opts *ManagerOpts, db *gorm.DB) (*Manager, error) {
	var (
		managerPK ci.PrivKey
		zonePK    ci.PrivKey
		err       error
	)
	if opts == nil {
		managerPK, _, err = ci.GenerateKeyPair(ci.Ed25519, 256)
		if err != nil {
			return nil, err
		}
		zonePK, _, err = ci.GenerateKeyPair(ci.Ed25519, 256)
		if err != nil {
			return nil, err
		}
		opts = &ManagerOpts{
			ManagerPK: managerPK,
			ZonePK:    zonePK,
			ZoneName:  "default",
		}
	}
	managerPKID, err := peer.IDFromPublicKey(opts.ManagerPK.GetPublic())
	if err != nil {
		return nil, err
	}
	zoneManager := ZoneManager{
		PublicKey: managerPKID.String(),
	}
	zonePKID, err := peer.IDFromPublicKey(opts.ZonePK.GetPublic())
	if err != nil {
		return nil, err
	}
	zone := Zone{
		Name:      opts.ZoneName,
		PublicKey: zonePKID.String(),
		Manager:   &zoneManager,
	}
	manager := Manager{
		PrivateKey:        opts.ManagerPK,
		ZonePrivateKey:    opts.ZonePK,
		RecordPrivateKeys: nil,
		Zone:              &zone,
	}
	if db != nil {
		manager.ZM = models.NewZoneManager(db)
		manager.RM = models.NewRecordManager(db)
	}
	return &manager, nil
}

// RunTNSDaemon is used to run our TNS daemon, and setup the available stream handlers
func (m *Manager) RunTNSDaemon() {
	fmt.Println("generating echo stream")
	m.Host.SetStreamHandler(
		CommandEcho, func(s net.Stream) {
			log.Info("new stream detected")
			if err := m.HandleQuery(s, "echo"); err != nil {
				log.Warn(err.Error())
				s.Reset()
			} else {
				s.Close()
			}
		})
	fmt.Println("generating record request stream")
	m.Host.SetStreamHandler(
		CommandRecordRequest, func(s net.Stream) {
			log.Info("new stream detected")
			if err := m.HandleQuery(s, "record-request"); err != nil {
				log.Warn(err.Error())
				s.Reset()
			} else {
				s.Close()
			}
		})
	fmt.Println("generating zone request stream")
	m.Host.SetStreamHandler(
		CommandZoneRequest, func(s net.Stream) {
			log.Info("new stream detected")
			if err := m.HandleQuery(s, "zone-request"); err != nil {
				log.Warn(err.Error())
				s.Reset()
			} else {
				s.Close()
			}
		})
}

// HandleQuery is used to handle a query sent to tns
func (m *Manager) HandleQuery(s net.Stream, cmd string) error {
	responseBuffer := bufio.NewReader(s)
	switch cmd {
	case "echo":
		bodyBytes, err := responseBuffer.ReadString('\n')
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("echo test...\nyou sent: %s\n", string(bodyBytes))
		fmt.Println(msg)
		_, err = s.Write([]byte(msg))
		return err
	case "record-request":
		bodyBytes, err := responseBuffer.ReadBytes('\n')
		if err != nil {
			return err
		}
		req := RecordRequest{}
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			return err
		}
		r, err := m.RM.FindRecordByNameAndUser(req.UserName, req.RecordName)
		if err != nil {
			return err
		}
		fmt.Printf("record request\n%+v\n", req)
		_, err = s.Write([]byte(r.LatestIPFSHash))
		return err
	case "zone-request":
		bodyBytes, err := responseBuffer.ReadBytes('\n')
		if err != nil {
			return err
		}
		req := ZoneRequest{}
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			return err
		}
		fmt.Printf("zone request\n%+v\n", req)
		z, err := m.ZM.FindZoneByNameAndUser(req.ZoneName, req.UserName)
		if err != nil {
			return err
		}
		fmt.Printf("zone file recovered from database\n%+v\n", z)
		_, err = s.Write([]byte(z.LatestIPFSHash))
		return err
	default:
		fmt.Println("unsupported command type")
		_, err := s.Write([]byte("message received thanks"))
		return err
	}
}

// MakeHost is used to generate the libp2p connection for our TNS daemon
func (m *Manager) MakeHost(pk ci.PrivKey, opts *HostOpts) error {
	host, err := makeHost(pk, opts, false)
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
