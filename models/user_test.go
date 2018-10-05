package models_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"
)

var (
	testCfgPath = "../test/config.json"
)

type args struct {
	userName          string
	email             string
	password          string
	enterpriseEnabled bool
}

func TestUserManager_ChangePassword(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	um := models.NewUserManager(db)

	var (
		randUtils = utils.GenerateRandomUtils()
		username  = randUtils.GenerateString(10, utils.LetterBytes)
		email     = randUtils.GenerateString(10, utils.LetterBytes)
	)

	tests := []struct {
		name string
		args args
	}{
		{"ChangePassword", args{username, email, "password123", false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := um.NewUserAccount(tt.args.userName, tt.args.password, tt.args.email, tt.args.enterpriseEnabled)
			if err != nil {
				t.Fatal(err)
			}
			defer um.DB.Delete(user)
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
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	um := models.NewUserManager(db)

	var (
		randUtils = utils.GenerateRandomUtils()
		username  = randUtils.GenerateString(10, utils.LetterBytes)
		email     = randUtils.GenerateString(10, utils.LetterBytes)
	)

	tests := []struct {
		name string
		args args
	}{
		{"AccountCreation", args{username, email, "password123", false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := um.NewUserAccount(tt.args.userName, tt.args.password, tt.args.email, tt.args.enterpriseEnabled)
			if err != nil {
				t.Fatal(err)
			}
			um.DB.Delete(user)
		})
	}
}

func TestUserManager_SignIn(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	um := models.NewUserManager(db)

	var (
		randUtils = utils.GenerateRandomUtils()
		username  = randUtils.GenerateString(10, utils.LetterBytes)
		email     = randUtils.GenerateString(10, utils.LetterBytes)
	)

	tests := []struct {
		name string
		args args
	}{
		{"AccountCreation", args{username, email, "password123", false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := um.NewUserAccount(tt.args.userName, tt.args.password, tt.args.email, tt.args.enterpriseEnabled)
			if err != nil {
				t.Fatal(err)
			}
			defer um.DB.Delete(user)
			success, err := um.SignIn(tt.args.userName, tt.args.password)
			if err != nil {
				t.Fatal(err)
			}
			if !success {
				t.Error("sign in failed")
			}
		})
	}
}

func openDatabaseConnection(t *testing.T, cfg *config.TemporalConfig) (*gorm.DB, error) {
	dbConnURL := fmt.Sprintf("host=127.0.0.1 port=5433 user=postgres dbname=temporal password=%s sslmode=disable",
		cfg.Database.Password)

	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		t.Fatal(err)
	}
	return db, nil
}
