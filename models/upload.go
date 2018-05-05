package models

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
)

type Upload struct {
	gorm.Model
	Hash             string         `gorm:not null;`
	Type             string         `gorm:not null;` //  file, pin
	HoldTimeInMonths int64          `gorm:not null;`
	UploadAddress    common.Address `gorm:not null;`
}
