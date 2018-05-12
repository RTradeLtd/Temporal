package database_test

import (
	"log"
	"testing"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	rollbar "github.com/rollbar/rollbar-go"
)

func rollbarError(err error) {
	rollbar.SetToken("046baf3d0cd4422c8801891f5f79b65d")
	rollbar.SetServerRoot("github.com/RTradeLtd/Temporal")
	rollbar.Error(err)
	rollbar.Wait()
}

func openDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./ipfs_database.db")
	if err != nil {
		rollbarError(err)
		log.Fatal(err)
	}
	return db
}
func TestOpenDBConnection(t *testing.T) {
	db, err := gorm.Open("sqlite3", "./ipfs_database.db")
	defer db.Close()
	if err != nil {
		rollbarError(err)
		t.Fatal(err)
	}
}

func TestAddHash(t *testing.T) {
	var upload models.Upload
	hash := "Qmcr3dDpmpkLNRchqy5iSizTFyx9w5CUU8VoULoxxKo1As"
	upload.Hash = hash
	upload.Type = "pin"
	db := openDB()
	db.Create(&upload)
	db.Close()
}
