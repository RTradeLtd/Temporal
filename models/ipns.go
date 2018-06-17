package models

import (
	"github.com/jinzhu/gorm"
)

// IPNS will hold all of the IPNS entries in our system
type IPNS struct {
	gorm.Model
	ExpiryDateUnix int64    `gorm:"type:integer" json:"expiry_date_unix"`
	IPNSHash       string   `gorm:"type:varchar(255);unique" json:"ipns_hash"`
	IPFSHash       []string `gorm:"type:text[]" json:"ipfs_hash"`
	Lifetime       string   `gorm:"type:varchar(255)" json:"life_time"`
	TTL            string   `gorm:"type:varchar(255)" json:"ttl"`
	Key            string   `gorm:"type:varchar(255)" json:"key"`
}
