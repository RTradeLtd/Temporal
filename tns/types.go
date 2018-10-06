package tns

import (
	ci "github.com/libp2p/go-libp2p-crypto"
)

// Zone is a mapping of human readable names, mapped to a public key. In order to retrieve the latest
type Zone struct {
	Manager   *ZoneManager `json:"zone_manager"`
	PublicKey ci.PubKey    `json:"zone_public_key"`
	// A human readable name for this zone
	Name string `json:"name"`
	// A map of records managed by this zone
	Records                 map[string]*Record   `json:"records"`
	RecordNamesToPublicKeys map[string]ci.PubKey `json:"record_names_to_public_keys"`
}

// Record is a particular name entry managed by a zone
type Record struct {
	PublicKey *ci.PubKey `json:"public_key"`
	// A human readable name for this record
	Name string `json:"name"`
	// User configurable meta data for this record
	MetaData map[string]interface{} `json:"meta_data"`
}

// ZoneManager is the authorized manager of a zone
type ZoneManager struct {
	PublicKey ci.PubKey `json:"public_key"`
}
