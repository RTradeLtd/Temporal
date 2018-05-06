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
