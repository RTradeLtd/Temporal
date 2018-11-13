package tns

import (
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/database/models"
	ci "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
)

const (
	// CommandEcho is a test command used to test if we have successfully connected to a tns daemon
	CommandEcho = "/echo/1.0.0"
	// CommandRecordRequest is a command used to request a record from tns
	CommandRecordRequest = "/recordRequest/1.0.0"
	// CommandZoneRequest is a command used to request a zone from tns
	CommandZoneRequest = "/zoneRequest/1.0.0"
)

var (
	// Commands are all the commands that TNS supports via the libp2p interface
	Commands = []string{CommandEcho, CommandRecordRequest, CommandZoneRequest}
)

// RecordRequest is a message sent when requeting a record form TNS, the response is simply Record
type RecordRequest struct {
	RecordName string `json:"record_name"`
	UserName   string `json:"user_name"`
}

// ZoneRequest is a message sent when requesting a reccord from TNS.
type ZoneRequest struct {
	UserName           string `json:"user_name"`
	ZoneName           string `json:"zone_name"`
	ZoneManagerKeyName string `json:"zone_manager_key_name"`
}

// Zone is a mapping of human readable names, mapped to a public key. In order to retrieve the latest
type Zone struct {
	Manager   *ZoneManager `json:"zone_manager"`
	PublicKey string       `json:"zone_public_key"`
	// A human readable name for this zone
	Name string `json:"name"`
	// A map of records managed by this zone
	Records                 map[string]*Record `json:"records"`
	RecordNamesToPublicKeys map[string]string  `json:"record_names_to_public_keys"`
}

// Record is a particular name entry managed by a zone
type Record struct {
	PublicKey string `json:"public_key"`
	// A human readable name for this record
	Name string `json:"name"`
	// User configurable meta data for this record
	MetaData map[string]interface{} `json:"meta_data"`
}

// ZoneManager is the authorized manager of a zone
type ZoneManager struct {
	PublicKey string `json:"public_key"`
}

// HostOpts is our options for when we create our libp2p host
type HostOpts struct {
	IPAddress string `json:"ip_address"`
	Port      string `json:"port"`
	IPVersion string `json:"ip_version"`
	Protocol  string `json:"protocol"`
}

// Manager is used to manipulate a zone on TNS and run a daemon
type Manager struct {
	PrivateKey        ci.PrivKey
	ZonePrivateKey    ci.PrivKey
	RecordPrivateKeys map[string]ci.PrivKey
	Zone              *Zone
	Host              host.Host
	ZM                *models.ZoneManager
	RM                *models.RecordManager
	l                 *log.Logger
	service           string
}

// Client is used to query a TNS daemon
type Client struct {
	PrivateKey ci.PrivKey
	Host       host.Host
	IPFSAPI    string
}

// Host is an interface used by a TNS client or daemon
type Host interface {
	MakeHost(pk ci.PrivKey, opts *HostOpts) error
}
