package models

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

/*
	EMAIL ADDRESS MUST BE PROVIDED
*/
type User struct {
	gorm.Model
	EthAddress        string `gorm:"type:varchar(255);unique"`
	UserName          string `gorm:"type:varchar(255);unique"`
	EmailAddress      string `gorm:"type:varchar(255);unique"`
	EnterpriseEnabled bool   `gorm:"type:boolean"`
	AccountEnabled    bool   `gorm:"type:boolean"`
	APIAccess         bool   `gorm:"type:boolean"`
	EmailEnabled      bool   `gorm:"type:boolean"`
	HashedPassword    string `gorm:"type:varchar(255)"`
	// IPFSKeyNames is an array of IPFS keys this user has created
	IPFSKeyNames     pq.StringArray `gorm:"type:text[];column:ipfs_key_names"`
	IPFSKeyIDs       pq.StringArray `gorm:"type:text[];column:ipfs_key_ids"`
	IPFSNetworkNames pq.StringArray `gorm:"type:text[];column:ipfs_network_names"`
}

type UserManager struct {
	DB *gorm.DB
}

func NewUserManager(db *gorm.DB) *UserManager {
	um := UserManager{}
	um.DB = db
	return &um
}

func (um *UserManager) GetPrivateIPFSNetworksForUser(username string) ([]string, error) {
	u := &User{}
	if check := um.DB.Where("user_name = ?", username).First(u); check.Error != nil {
		return nil, check.Error
	}
	return u.IPFSNetworkNames, nil
}

func (um *UserManager) CheckIfUserHasAccessToNetwork(username, networkName string) (bool, error) {
	u := &User{}
	if check := um.DB.Where("user_name = ?", username).First(u); check.Error != nil {
		return false, check.Error
	}
	for _, v := range u.IPFSNetworkNames {
		if v == networkName {
			return true, nil
		}
	}
	return false, nil
}
func (um *UserManager) AddIPFSNetworkForUser(username, networkName string) error {
	u := &User{}
	if check := um.DB.Where("user_name = ?", username).First(u); check.Error != nil {
		return check.Error
	}
	for _, v := range u.IPFSNetworkNames {
		if v == networkName {
			return errors.New("network already configured for user")
		}
	}
	u.IPFSNetworkNames = append(u.IPFSNetworkNames, networkName)
	if check := um.DB.Model(u).Update("ipfs_network_names", u.IPFSNetworkNames); check.Error != nil {
		return check.Error
	}

	return nil
}

func (um *UserManager) AddIPFSKeyForUser(username, keyName, keyID string) error {
	var user User
	if errCheck := um.DB.Where("user_name = ?", username).First(&user); errCheck.Error != nil {
		return errCheck.Error
	}

	if user.CreatedAt == nilTime {
		return errors.New("user account does not exist")
	}

	for _, v := range user.IPFSKeyNames {
		if v == keyName {
			fmt.Println("key already exists in db, skipping")
			return nil
		}
	}
	user.IPFSKeyNames = append(user.IPFSKeyNames, keyName)
	user.IPFSKeyIDs = append(user.IPFSKeyIDs, keyID)
	// The following only updates the specified column for the given model
	if errCheck := um.DB.Model(&user).Updates(map[string]interface{}{
		"ipfs_key_names": user.IPFSKeyNames,
		"ipfs_key_ids":   user.IPFSKeyIDs,
	}); errCheck.Error != nil {
		return errCheck.Error
	}
	return nil
}

func (um *UserManager) GetKeysForUser(username string) (map[string][]string, error) {
	var user User
	keys := make(map[string][]string)
	if errCheck := um.DB.Where("user_name = ?", username).First(&user); errCheck.Error != nil {
		return nil, errCheck.Error
	}

	if user.CreatedAt == nilTime {
		return nil, errors.New("user account does not exist")
	}

	keys["key_names"] = user.IPFSKeyNames
	keys["key_ids"] = user.IPFSKeyIDs
	return keys, nil
}

func (um *UserManager) GetKeyIDByName(username, keyName string) (string, error) {
	var user User
	if errCheck := um.DB.Where("user_name = ?", username).First(&user); errCheck.Error != nil {
		return "", errCheck.Error
	}

	if user.CreatedAt == nilTime {
		return "", errors.New("user account does not exist")
	}
	for k, v := range user.IPFSKeyNames {
		if v == keyName {
			return user.IPFSKeyIDs[k], nil
		}
	}
	return "", errors.New("key not found")
}

func (um *UserManager) CheckIfKeyOwnedByUser(username, keyName string) (bool, error) {
	var user User
	if errCheck := um.DB.Where("user_name = ?", username).First(&user); errCheck.Error != nil {
		return false, errCheck.Error
	}

	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}

	for _, v := range user.IPFSKeyNames {
		if v == keyName {
			return true, nil
		}
	}
	return false, nil
}

func (um *UserManager) CheckIfUserAccountEnabled(username string, db *gorm.DB) (bool, error) {
	var user User
	db.Where("user_name = ?", username).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	return user.AccountEnabled, nil
}

// ChangePassword is used to change a users password
func (um *UserManager) ChangePassword(username, currentPassword, newPassword string) (bool, error) {
	var user User
	um.DB.Where("user_name = ?", username).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	decodedPassword, err := hex.DecodeString(user.HashedPassword)
	if err != nil {
		return false, err
	}
	if err := bcrypt.CompareHashAndPassword(decodedPassword, []byte(currentPassword)); err != nil {
		return false, errors.New("invalid current password")
	}
	newHashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	encodedNewHashedPass := hex.EncodeToString(newHashedPass)
	if check := um.DB.Model(&user).Update("hashed_password", encodedNewHashedPass); check.Error != nil {
		return false, check.Error
	}
	return true, nil
}

func (um *UserManager) NewUserAccount(ethAddress, username, password, email string, enterpriseEnabled bool) (*User, error) {
	user := User{}
	check := um.DB.Where("user_name = ?", username).First(&user)
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	if user.CreatedAt != nilTime {
		return nil, errors.New("user account already created")
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if ethAddress != "" {
		user.EthAddress = ethAddress
	}
	user.UserName = username
	user.EnterpriseEnabled = enterpriseEnabled
	user.HashedPassword = hex.EncodeToString(hashedPass)
	user.EmailAddress = email
	user.AccountEnabled = true
	if check := um.DB.Create(&user); check.Error != nil {
		return nil, check.Error
	}
	return &user, nil
}

// SignIn is used to authenticate a user, and check if their account is enabled.
// Returns bool on succesful login, or false with an error on failure
func (um *UserManager) SignIn(username, password string) (bool, error) {
	var user User
	if check := um.DB.Where("user_name = ?", username).First(&user); check.Error != nil {
		return false, check.Error
	}
	if !user.AccountEnabled {
		return false, errors.New("account is disabled")
	}
	validPassword, err := um.ComparePlaintextPasswordToHash(username, password)
	if err != nil {
		return false, err
	}
	if !validPassword {
		return false, errors.New("invalid password supplied")
	}
	return true, nil
}

func (um *UserManager) ComparePlaintextPasswordToHash(username, password string) (bool, error) {
	var user User
	um.DB.Where("user_name = ?", username).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	passwordBytes, err := hex.DecodeString(user.HashedPassword)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword(passwordBytes, []byte(password))
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

func (um *UserManager) FindEthAddressByUserName(username string) (string, error) {
	u := User{}
	if check := um.DB.Where("user_name = ?", username).First(&u); check.Error != nil {
		return "", check.Error
	}
	return u.EthAddress, nil
}

// FindEmailByUserName is used to find an email address by searching for the users eth address
// the returned map contains their eth address as a key, and their email address as a value
func (um *UserManager) FindEmailByUserName(username string) (map[string]string, error) {
	u := User{}
	check := um.DB.Where("user_name = ?", username).First(&u)
	if check.Error != nil {
		return nil, check.Error
	}
	emails := make(map[string]string)
	emails[username] = u.EmailAddress
	return emails, nil
}

// ChangeEthereumAddress is used to change a user's ethereum address
func (um *UserManager) ChangeEthereumAddress(username, ethAddress string) (*User, error) {
	u := User{}
	if check := um.DB.Where("user_name = ?", username).First(&u); check.Error != nil {
		return nil, check.Error
	}
	u.EthAddress = ethAddress
	if check := um.DB.Model(u).Update("eth_address", ethAddress); check.Error != nil {
		return nil, check.Error
	}
	return &u, nil
}
