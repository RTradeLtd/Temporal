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
	EmailAddress      string `gorm:"type:varchar(255);unique"`
	EnterpriseEnabled bool   `gorm:"type:boolean"`
	AccountEnabled    bool   `gorm:"type:boolean"`
	APIAccess         bool   `gorm:"type:boolean"`
	EmailEnabled      bool   `gorm:"type:boolean"`
	HashedPassword    string `gorm:"type:varchar(255)"`
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

// ChangePassword is used to change a users password
func (um *UserManager) ChangePassword(ethAddress, currentPassword, newPassword string) (bool, error) {
	var user User
	um.DB.Where("eth_address = ?", ethAddress).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(currentPassword)); err != nil {
		return false, errors.New("invalid current password")
	}
	newHashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	check := um.DB.Model(&user).Update("hashed_password", string(newHashedPass))
	if check.Error != nil {
		return false, err
	}
	return true, nil
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

// SignIn is used to authenticate a user, and check if their account is enabled.
// Returns bool on succesful login, or false with an error on failure
func (um *UserManager) SignIn(ethAddress, password string) (bool, error) {
	var user User
	um.DB.Where("eth_address = ?", ethAddress).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	if !user.AccountEnabled {
		return false, errors.New("account is marked is disabled")
	}
	validPassword, err := um.ComparePlaintextPasswordToHash(ethAddress, password)
	if err != nil {
		return false, err
	}
	if !validPassword {
		return false, errors.New("invalid password supplied")
	}
	return true, nil
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
