package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	EthAddress string   `gorm:"not null;unique"`
	Uploads    []Upload `json:"uploads"`
}

type UserManager struct {
	DB *gorm.DB
}

func NewUserManager(db *gorm.DB) *UserManager {
	um := UserManager{}
	um.DB = db
	return &um
}

func (um *UserManager) FindByAddress(address string) *User {
	u := User{
		EthAddress: address,
	}

	um.DB.First(&u)

	return &u
}
