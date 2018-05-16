package models

import (
	"github.com/jinzhu/gorm"
)

type Upload struct {
	gorm.Model
	Hash             string `gorm:not null;`
	Type             string `gorm:not null;` //  file, pin
	HoldTimeInMonths int64  `gorm:not null;`
	UploadAddress    string `gorm:not null;`
}

type UploadDatabase struct {
	DB *gorm.DB
}

func NewUploadDatabase(db *gorm.DB) *UploadDatabase {
	return &UploadDatabase{DB: db}
}

// FindUploadsByHash is used to return all instances of uploads matching the
// given hash
func (ud *UploadDatabase) FindUploadsByHash(hash string) []*Upload {

	uploads := []*Upload{}

	ud.DB.Find(&uploads).Where("hash = ?", hash)

	return uploads
}
