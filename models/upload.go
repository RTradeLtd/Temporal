package models

import (
	"github.com/jinzhu/gorm"
)

type Upload struct {
	gorm.Model
	Hash string `gorm:not null`
}
