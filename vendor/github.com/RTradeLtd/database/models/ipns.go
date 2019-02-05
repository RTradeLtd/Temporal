package models

import (
	"errors"
	"time"

	"github.com/RTradeLtd/gorm"
	"github.com/lib/pq"
)

// IPNS will hold all of the IPNS entries in our system
type IPNS struct {
	gorm.Model
	Sequence int64 `gorm:"type:integer"`
	// the ipns hash, is the peer id of the peer used to sign the entry
	IPNSHash string `gorm:"type:varchar(255);unique"`
	// List of content hashes this IPNS entry has pointed to
	IPFSHashes      pq.StringArray `gorm:"type:text[]"`
	CurrentIPFSHash string         `gorm:"type:varchar(255)"`
	LifeTime        string         `gorm:"type:varchar(255)"`
	TTL             string         `gorm:"type:varchar(255)"`
	Key             string         `gorm:"type:varchar(255)"`
	NetworkName     string         `gorm:"type:varchar(255)"`
	UserName        string         `gorm:"type:varchar(255)"`
}

// IpnsManager is used for manipulating IPNS records in our database
type IpnsManager struct {
	DB *gorm.DB
}

var nilEntry IPNS

// NewIPNSManager is used to generate our model manager
func NewIPNSManager(db *gorm.DB) *IpnsManager {
	return &IpnsManager{DB: db}
}

// FindByUserName is used to find all IPNS entries published by a given user
func (im *IpnsManager) FindByUserName(username string) (*[]IPNS, error) {
	entries := []IPNS{}
	if check := im.DB.Where("user_name = ?", username).Find(&entries); check.Error != nil {
		return nil, check.Error
	}
	return &entries, nil
}

// FindAll is used to find all IPNS records
func (im *IpnsManager) FindAll() ([]IPNS, error) {
	entries := []IPNS{}
	if err := im.DB.Model(&IPNS{}).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// FindByIPNSHash is used to find an IPNS record from our database searching for
// the public key hash of the key that was used to pulish a record
func (im *IpnsManager) FindByIPNSHash(ipnsHash string) (*IPNS, error) {
	var entry IPNS
	if check := im.DB.Where("ip_ns_hash = ?", ipnsHash).First(&entry); check.Error != nil {
		return nil, check.Error
	}
	return &entry, nil
}

// UpdateIPNSEntry is used to update an already existing IPNS entry, creating a no record matching the hash exists
func (im *IpnsManager) UpdateIPNSEntry(ipnsHash, ipfsHash, networkName, username string, lifetime, ttl time.Duration) (*IPNS, error) {
	var entry IPNS
	// search for an IPNS entry that matches the given ipns hash
	if check := im.DB.Where("ip_ns_hash = ? AND network_name = ?", ipnsHash, networkName).First(&entry); check.Error != nil {
		return nil, check.Error
	}
	// increase sequence
	entry.Sequence++
	// update the hashes it has pointed to
	entry.IPFSHashes = append(entry.IPFSHashes, ipfsHash)
	// update the current hash this record points to
	entry.CurrentIPFSHash = ipfsHash
	// update the lifetime
	entry.LifeTime = lifetime.String()
	// update the ttl
	entry.TTL = ttl.String()
	// only update  changed fields
	check := im.DB.Model(&entry).Updates(map[string]interface{}{
		"sequence":           entry.Sequence,
		"ip_fs_hashes":       entry.IPFSHashes,
		"current_ip_fs_hash": entry.CurrentIPFSHash,
		"life_time":          entry.LifeTime,
		"ttl":                entry.TTL,
	})

	if check.Error != nil {
		return nil, check.Error
	}
	return &entry, nil
}

// CreateEntry is used to create a brand new IPNS entry in our database
func (im *IpnsManager) CreateEntry(ipnsHash, ipfsHash, key, networkName, username string, lifetime, ttl time.Duration) (*IPNS, error) {
	// See above UpdateEntry function for an explanation
	if _, err := im.FindByIPNSHash(ipnsHash); err == nil {
		return nil, errors.New("ipns hash already exists in database")
	}
	entry := IPNS{
		Sequence:        1,
		IPNSHash:        ipnsHash,
		CurrentIPFSHash: ipfsHash,
		IPFSHashes:      pq.StringArray{ipfsHash},
		LifeTime:        lifetime.String(),
		TTL:             ttl.String(),
		Key:             key,
		NetworkName:     networkName,
		UserName:        username,
	}
	if check := im.DB.Create(&entry); check.Error != nil {
		return nil, check.Error
	}
	return &entry, nil
}
