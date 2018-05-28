package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	EthAddress        string `gorm:"type:varchar(255);unique"`
	EnterpriseEnabled bool   `gorm:"type:boolean"`
	AccountEnabled    bool   `gorm:"type:boolean"`
	HashedPassword    string `gorm:"type:varchar(255)`
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

func (um *UserManager) CheckIfUserAccountEnabled(ethAddress string, db *gorm.DB) (bool, error) {
	var user User
	db.Where("eth_address = ?", ethAddress).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	return user.AccountEnabled, nil
}

func (um *UserManager) NewUserAccount(ethAddress, password string, enterpriseEnabled bool) (*User, error) {
	var user User
	um.DB.Where("eth_address = ?", ethAddress).First(&user)
	if user.CreatedAt != nilTime {
		return nil, errors.New("user account already created")
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.EthAddress = ethAddress
	user.EnterpriseEnabled = enterpriseEnabled
	user.HashedPassword = string(hashedPass)
	um.DB.Create(&user)
	return &user, nil
}

func (um *UserManager) ComparePlaintextPasswordToHash(ethAddress, password string) (bool, error) {
	var user User
	um.DB.Where("eth_address = ?", ethAddress).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil

}

func (um *UserManager) FindByAddress(address string) *User {
	u := User{}
	um.DB.Where("eth_address = ?", address).Find(&u)
	if u.CreatedAt == nilTime {
		return nil
	}
	return &u
}
