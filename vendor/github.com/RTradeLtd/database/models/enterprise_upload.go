package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type EnterpriseUpload struct {
	gorm.Model
	CompanyName        string `gorm:"type:varchar(255)" json:"company_name"`
	Hash               string `gorm:"type:varchar(255)" json:"hash"`
	GarbageCollectDate time.Time
}
