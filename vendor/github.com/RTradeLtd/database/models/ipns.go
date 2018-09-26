package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// IPNS will hold all of the IPNS entries in our system
type IPNS struct {
	gorm.Model
	Sequence int64 `gorm:"type:integer;not null;default:0" json:"sequence"`
	// the ipns hash, is the peer id of the peer used to sign the entry
	IPNSHash string `gorm:"type:varchar(255);unique;column:ipns_hash" json:"ipns_hash"`
	// List of content hashes this IPNS entry has pointed to
	IPFSHashes      pq.StringArray `gorm:"type:text[];column:ipfs_hash" json:"ipfs_hashes"`
	CurrentIPFSHash string         `gorm:"type:varchar(255);column:current_ipfs_hash" json:"current_ipfs_hash"`
	LifeTime        string         `gorm:"type:varchar(255)" json:"life_time"`
	TTL             string         `gorm:"type:varchar(255)" json:"ttl"`
	Key             string         `gorm:"type:varchar(255)" json:"key"`
	NetworkName     string         `gorm:"type:varchar(255)" json:"network_name"`
}

type IpnsManager struct {
	DB *gorm.DB
}

var nilEntry IPNS

func NewIPNSManager(db *gorm.DB) *IpnsManager {
	return &IpnsManager{DB: db}
}

func (im *IpnsManager) FindByIPNSHash(ipnsHash string) (*IPNS, error) {
	var entry IPNS
	im.DB.Table("ip_ns").Where("ipns_hash = ?", ipnsHash).First(&entry)
	if entry.CreatedAt == nilTime {
		return nil, errors.New("ipns hash does not exist")
	}
	return &entry, nil
}

func (im *IpnsManager) UpdateIPNSEntry(ipnsHash, ipfsHash, key, networkName string, lifetime, ttl time.Duration) (*IPNS, error) {
	var entry IPNS
	// search for an IPNS entry that matches the given ipns hash
	if check := im.DB.Where("ipns_hash = ? AND network_name = ?", ipnsHash, networkName).First(&entry); check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	// if the returned model does not exist create it
	if entry.CreatedAt == nilTime {
		// Create the record
		entry, err := im.CreateEntry(ipnsHash, ipfsHash, key, networkName, lifetime, ttl)
		if err != nil {
			return nil, err
		}
		return entry, nil
	}
	// increase the sequence number
	entry.Sequence++
	// update the hashes it has pointed to
	entry.IPFSHashes = append(entry.IPFSHashes, ipfsHash)
	// update the current hash this record points to
	entry.CurrentIPFSHash = ipfsHash
	// update the lifetime
	entry.LifeTime = lifetime.String()
	// update the ttl
	entry.TTL = ttl.String()
	// update the key used to sign
	entry.Key = key
	// only update  changed fields
	check := im.DB.Table("ip_ns").Model(&entry).Updates(map[string]interface{}{
		"sequence":          &entry.Sequence,
		"ipfs_hashes":       &entry.IPFSHashes,
		"current_ipfs_hash": &entry.CurrentIPFSHash,
		"lifeTime":          &entry.LifeTime,
		"ttl":               &entry.TTL,
		"key":               &entry.Key,
	})

	if check.Error != nil {
		return nil, check.Error
	}
	return &entry, nil
}

func (im *IpnsManager) CreateEntry(ipnsHash, ipfsHash, key, networkName string, lifetime, ttl time.Duration) (*IPNS, error) {
	// See above UpdateEntry function for an explanation
	var entry IPNS
	err := im.DB.Where("ipns_hash = ? AND network_name = ?", ipnsHash, networkName)
	if err == nil {
		return nil, errors.New("ipns hash already exists")
	}
	entry.Sequence = 1
	entry.IPNSHash = ipnsHash
	entry.IPFSHashes = pq.StringArray{ipfsHash}
	entry.LifeTime = lifetime.String()
	entry.TTL = ttl.String()
	entry.Key = key
	entry.NetworkName = networkName
	if check := im.DB.Create(&entry); check.Error != nil {
		return nil, check.Error
	}
	return &entry, nil
}
