package models_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/jinzhu/gorm"
)

const (
	defaultConfigFile = "/home/solidity/config.json"
)

var (
	travis = os.Getenv("TRAVIS") != ""
	dbPass string
)

type args struct {
	ethAddress        string
	userName          string
	email             string
	password          string
	enterpriseEnabled bool
}

func TestUser(t *testing.T) {
	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	randUtils := utils.GenerateRandomUtils()
	username := randUtils.GenerateString(10, utils.LetterBytes)
	ethAddress := randUtils.GenerateString(10, utils.LetterBytes)
	email := randUtils.GenerateString(10, utils.LetterBytes)
	tests := []struct {
		name string
		args args
	}{
		{"NewAccount", args{ethAddress, username, email, "password123", false}},
		{"ChangePassword", args{ethAddress, username, email, "password123", false}},
		{"ChangeEthereumAddress", args{ethAddress, username, email, "password123", false}},
	}
	for _, tt := range tests {
		switch tt.name {
		case "NewAccount":
			_, err = newAccount(t, db, tt.args)
			if err != nil {
				t.Fatal(err)
			}
		case "ChangePassword":
			changed, err := changePassword(t, db, tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if !changed {
				t.Error("password changed failed but no error was raised")
			}
		case "ChangeEthereumAddress":
			_, err = changeEthereumAddress(t, db, tt.args)
			if err != nil {
				t.Error(err)
			}
		default:
			fmt.Println("skipping... invalid test name")
		}
	}
}

func newAccount(t *testing.T, db *gorm.DB, args args) (*models.User, error) {
	um := models.NewUserManager(db)
	user, err := um.NewUserAccount(args.ethAddress, args.userName, args.password, args.email, args.enterpriseEnabled)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func changePassword(t *testing.T, db *gorm.DB, args args) (bool, error) {
	um := models.NewUserManager(db)
	return um.ChangePassword(args.userName, args.password, "newpassword")
}

func changeEthereumAddress(t *testing.T, db *gorm.DB, args args) (*models.User, error) {
	um := models.NewUserManager(db)
	user, err := um.ChangeEthereumAddress(args.userName, args.ethAddress)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func openDatabaseConnection(t *testing.T, cfg *config.TemporalConfig) (*gorm.DB, error) {
	if !travis {
		dbPass = cfg.Database.Password
	} else {
		dbPass = ""
	}
	dbConnURL := fmt.Sprintf("host=127.0.0.1 port=5432 user=postgres dbname=temporal password=%s sslmode=disable", dbPass)

	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		t.Fatal(err)
	}
	return db, nil
}
