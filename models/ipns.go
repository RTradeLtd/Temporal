package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// IPNS will hold all of the IPNS entries in our system
type IPNS struct {
	gorm.Model
	Sequence int64 `gorm:"type:integer;not null;default:0" json:"sequence"`
	// the ipns hash, is the peer id of the peer used to sign the entry
	IPNSHash string `gorm:"type:varchar(255);unique;column:ipns_hash" json:"ipns_hash"`
	// List of content hashes this IPNS entry has pointed to
	IPFSHashes      []string `gorm:"type:text[];column:ipfs_hash" json:"ipfs_hashes"`
	CurrentIPFSHash string   `gorm:"type:varchar(255);column:current_ipfs_hash" json:"current_ipfs_hash"`
	LifeTime        string   `gorm:"type:varchar(255)" json:"life_time"`
	TTL             string   `gorm:"type:varchar(255)" json:"ttl"`
	Key             string   `gorm:"type:varchar(255)" json:"key"`
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
	im.DB.Table("ip_ns").Where("ipns_hash = ?").First(&entry)
	if entry.CreatedAt == nilTime {
		return nil, errors.New("ipns hash does not exist")
	}
	return &entry, nil
}

func (im *IpnsManager) UpdateIPNSEntry(ipnsHash, ipfsHash, lifetime, ttl, key string) (*IPNS, error) {
	var entry IPNS
	im.DB.Where("ipns_hash = ?", ipnsHash).First(&entry)
	/*if entry.CreatedAt == nilTime {
		return nil, errors.New("ipns hash does not exist")
	}*/
	entry.Sequence++
	entry.IPNSHash = ipnsHash
	entry.IPFSHashes = append(entry.IPFSHashes, ipfsHash)
	entry.CurrentIPFSHash = ipfsHash
	entry.LifeTime = lifetime
	entry.TTL = ttl
	entry.Key = key
	// only update  changed fields
	check := im.DB.Table("ip_ns").Model(&nilEntry).Updates(map[string]interface{}{
		"sequence":          &entry.Sequence,
		"ipns_hash":         &entry.IPNSHash,
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

func (im *IpnsManager) CreateEntry(expiryDateUnix int64, ipnsHash, ipfsHash, lifetime, ttl, key string, forceUpdate bool) (*IPNS, error) {
	var entry *IPNS
	entry, err := im.FindByIPNSHash(ipnsHash)
	if err == nil {
		return nil, errors.New("ipns hash already exists")
	}
	return entry, nil
}
