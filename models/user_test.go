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

func TestUserManager_ChangeEthereumAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	um := models.NewUserManager(db)

	var (
		randUtils  = utils.GenerateRandomUtils()
		username   = randUtils.GenerateString(10, utils.LetterBytes)
		ethAddress = randUtils.GenerateString(10, utils.LetterBytes)
		email      = randUtils.GenerateString(10, utils.LetterBytes)
	)

	tests := []struct {
		name string
		args args
	}{
		{"ChangeEthereumAddress", args{ethAddress, username, email, "password123", false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := um.NewUserAccount(tt.args.ethAddress, tt.args.userName, tt.args.password, tt.args.email, tt.args.enterpriseEnabled); err != nil {
				t.Fatal(err)
			}
			if _, err := um.ChangeEthereumAddress(tt.args.userName, tt.args.ethAddress); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestUserManager_ChangePassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	um := models.NewUserManager(db)

	var (
		randUtils  = utils.GenerateRandomUtils()
		username   = randUtils.GenerateString(10, utils.LetterBytes)
		ethAddress = randUtils.GenerateString(10, utils.LetterBytes)
		email      = randUtils.GenerateString(10, utils.LetterBytes)
	)

	tests := []struct {
		name string
		args args
	}{
		{"ChangePassword", args{ethAddress, username, email, "password123", false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := um.NewUserAccount(tt.args.ethAddress, tt.args.userName, tt.args.password, tt.args.email, tt.args.enterpriseEnabled); err != nil {
				t.Fatal(err)
			}
			changed, err := um.ChangePassword(tt.args.userName, tt.args.password, "newpassword")
			if err != nil {
				t.Fatal(err)
			}
			if !changed {
				t.Error("password changed failed, but no error occured")
			}
		})
	}
}

func TestUserManager_NewAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	um := models.NewUserManager(db)

	var (
		randUtils  = utils.GenerateRandomUtils()
		username   = randUtils.GenerateString(10, utils.LetterBytes)
		ethAddress = randUtils.GenerateString(10, utils.LetterBytes)
		email      = randUtils.GenerateString(10, utils.LetterBytes)
	)

	tests := []struct {
		name string
		args args
	}{
		{"AccountCreation", args{ethAddress, username, email, "password123", false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := um.NewUserAccount(tt.args.ethAddress, tt.args.userName, tt.args.password, tt.args.email, tt.args.enterpriseEnabled); err != nil {
				t.Fatal(err)
			}
		})
	}
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
