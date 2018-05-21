package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	EthAddress string   `gorm:"type:varchar(255);"`
	Uploads    []Upload `gorm:"many2many:user_uploads;"`
}

type UserManager struct {
	DB *gorm.DB
}

var nilTime time.Time

func NewUserManager(db *gorm.DB) *UserManager {
	um := UserManager{}
	um.DB = db
	return &um
}

func (um *UserManager) FindByAddress(address string) *User {
	u := User{}

	um.DB.Find(&u).Where("eth_address = ?", address)
	fmt.Println(u)
	if u.CreatedAt == nilTime {
		um.createIfNotFound(address)
	}
	return &u
}

func (um *UserManager) createIfNotFound(address string) {
	um.DB.Create(&User{EthAddress: address})
}
