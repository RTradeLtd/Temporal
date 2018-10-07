package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// Zone is a TNS zone
type Zone struct {
	gorm.Model
	UserName             string `gorm:"type:varchar(255)"`
	Name                 string `gorm:"type:varchar(255)"`
	ManagerPublicKeyHash string `gorm:"type:varchar(255)"`
	ZonePublicKeyHash    string `gorm:"type:varchar(255)"`
	LatestIPFSHash       string `gorm:"type:varchar(255)"`
}

// ZoneManager is used to manipulate zone entries in the database
type ZoneManager struct {
	DB *gorm.DB
}

// NewZoneManager is used to generate our zone manager helper to interact with the db
func NewZoneManager(db *gorm.DB) *ZoneManager {
	return &ZoneManager{DB: db}
}

// NewZone is used to create a new zone in the database
func (zm *ZoneManager) NewZone(username, name, managerPK, zonePK, latestIPFSHash string) (*Zone, error) {
	zone, err := zm.FindZoneByNameAndUser(name, username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		return nil, errors.New("zone already exists for user")
	}
	zone = &Zone{
		UserName:             username,
		Name:                 name,
		ManagerPublicKeyHash: managerPK,
		ZonePublicKeyHash:    zonePK,
		LatestIPFSHash:       latestIPFSHash,
	}
	if check := zm.DB.Create(zone); check.Error != nil {
		return nil, check.Error
	}
	return zone, nil
}

// FindZoneByNameAndUser is used to lookup a zone by name and user
func (zm *ZoneManager) FindZoneByNameAndUser(name, username string) (*Zone, error) {
	z := Zone{}
	if check := zm.DB.Where("name = ? AND user_name = ?", name, username).First(&z); check.Error != nil {
		return nil, check.Error
	}
	return &z, nil
}

// UpdateLatestIPFSHashForZone is used to update hte latest IPFS hash for a zone file
func (zm *ZoneManager) UpdateLatestIPFSHashForZone(name, username, hash string) (*Zone, error) {
	z := Zone{}
	if check := zm.DB.Where("name = ? AND user_name = ?", name, username).First(&z); check.Error != nil {
		return nil, check.Error
	}
	z.LatestIPFSHash = hash
	if check := zm.DB.Model(&z).Update("latest_ip_fs_hash", z.LatestIPFSHash); check.Error != nil {
		return nil, check.Error
	}
	return &z, nil
}
