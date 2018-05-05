package models

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	EthAddress common.Address `gorm:"not null;unique"`
	Uploads    []string       `json:"uploads"`
}
