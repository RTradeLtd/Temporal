package tns

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/RTradeLtd/database/models"
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
	LogFile   string     `json:"log_file"`
	DB        *gorm.DB   `json:"db"`
}

// GenerateTNSManager is used to generate a TNS manager for a particular PKI space
func GenerateTNSManager(opts *ManagerOpts, db *gorm.DB) (*Manager, error) {
	var (
		logger    = log.New()
		managerPK ci.PrivKey
		zonePK    ci.PrivKey
		err       error
	)
	// if opts is nil, generate a new identity
	// this is only particularly useful for running the echo test
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
			LogFile:   "./templogs.log",
		}
	}
	// extract a peer id for the zone manager
	managerPKID, err := peer.IDFromPublicKey(opts.ManagerPK.GetPublic())
	if err != nil {
		return nil, err
	}
	// format our zone manager
	zoneManager := ZoneManager{
		PublicKey: managerPKID.String(),
	}
	// extract a peer id for the zone
	zonePKID, err := peer.IDFromPublicKey(opts.ZonePK.GetPublic())
	if err != nil {
		return nil, err
	}
	// format our zone
	zone := Zone{
		Name:      opts.ZoneName,
		PublicKey: zonePKID.String(),
		Manager:   &zoneManager,
	}
	// create our manager struct which serves as the basis for the TNS manager daemon
	manager := Manager{
		PrivateKey:        opts.ManagerPK,
		ZonePrivateKey:    opts.ZonePK,
		RecordPrivateKeys: nil,
		Zone:              &zone,
		service:           "tns-manager",
	}
	// while a DB connection isn't necessary, it can allow for lower-latency answers
	if db != nil {
		manager.ZM = models.NewZoneManager(db)
		manager.RM = models.NewRecordManager(db)
	}
	// open log file
	logfile, err := os.OpenFile(opts.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %s", err)
	}
	logger.Out = logfile
	logger.Info("logger initialized")
	manager.l = logger
	return &manager, nil
}

// RunTNSDaemon is used to run our TNS daemon, and setup the available stream handlers
func (m *Manager) RunTNSDaemon() {
	m.LogInfo("generating echo stream")
	// our echo stream is a basic test used to determine whether or not a tns manager daemon is functioning properly
	m.Host.SetStreamHandler(
		CommandEcho, func(s net.Stream) {
			m.LogInfo("new stream detected")
			if err := m.HandleQuery(s, "echo"); err != nil {
				log.Warn(err.Error())
				s.Reset()
			} else {
				s.Close()
			}
		})
	m.LogInfo("generating record request stream")
	// our record request stream allows clients to request a record from the tns manager daemon
	m.Host.SetStreamHandler(
		CommandRecordRequest, func(s net.Stream) {
			m.LogInfo("new stream detected")
			if err := m.HandleQuery(s, "record-request"); err != nil {
				log.Warn(err.Error())
				s.Reset()
			} else {
				s.Close()
			}
		})
	m.LogInfo("generating zone request stream")
	// our zone request stream allows clients to request a zone from the tns manager daemon
	m.Host.SetStreamHandler(
		CommandZoneRequest, func(s net.Stream) {
			m.LogInfo("new stream detected")
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
		// read the message being sent by the client
		// it must end with a new line
		bodyBytes, err := responseBuffer.ReadString('\n')
		if err != nil {
			return err
		}
		// format a response
		msg := fmt.Sprintf("echo test...\nyou sent: %s\n", string(bodyBytes))
		// send a response to th eclient
		_, err = s.Write([]byte(msg))
		return err
	case "record-request":
		// read the message being sent by the client
		// it must end wit ha new line
		bodyBytes, err := responseBuffer.ReadBytes('\n')
		if err != nil {
			return err
		}
		// unmarshal the message into a record request type
		req := RecordRequest{}
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			return err
		}
		// search for the record  in the database
		// this is temporary, and will be expanded to allow the client to specify the source of information
		r, err := m.RM.FindRecordByNameAndUser(req.UserName, req.RecordName)
		if err != nil {
			return err
		}
		// send the latest ipfs hash for this record to the client, allowing them to extract information from ipfs
		_, err = s.Write([]byte(r.LatestIPFSHash))
		return err
	case "zone-request":
		// read the message being sent by the client
		// it must end wit ha new line
		bodyBytes, err := responseBuffer.ReadBytes('\n')
		if err != nil {
			return err
		}
		// unmarshal the message into a zone request type
		req := ZoneRequest{}
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			return err
		}
		// search for the zone in the database
		// this is temporary, and will be expanded to allow the client to specify the source of information
		z, err := m.ZM.FindZoneByNameAndUser(req.ZoneName, req.UserName)
		if err != nil {
			return err
		}
		// send the latest ipfs hash for this zone to the client, allowing them to extract information from ipfs
		_, err = s.Write([]byte(z.LatestIPFSHash))
		return err
	default:
		// basic handler for a generic stream
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
