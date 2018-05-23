package models

import "github.com/jinzhu/gorm"

type Payment struct {
	gorm.Model
	Uploader  string `gorm:"type:varchar(255);"`
	CID       string `gorm:"type:varchar(255);"`
	HashedCID string `gorm:"type:varchar(255);"`
	PaymentID string `gorm:"type:varchar(255);"`
	Paid      string `gorm:"type:boolean;"`
}
