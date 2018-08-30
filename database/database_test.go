package database_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/jinzhu/gorm"
)

var (
	travis = os.Getenv("TRAVIS") != ""
	dbPass string
)

func TestDatabase(t *testing.T) {
	if !travis {
		dbPass = "password123"
	} else {
		dbPass = ""
	}
	dbConnURL := fmt.Sprintf("host=127.0.0.1 port=5432 user=postgres dbname=temporal password=%s sslmode=disable", dbPass)

	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestDatabaseMigrations(t *testing.T) {
	if !travis {
		dbPass = "password123"
	} else {
		dbPass = ""
	}
	db, err := database.OpenDBConnection(database.DBOptions{
		User:           "postgres",
		Password:       dbPass,
		Address:        "127.0.0.1",
		SSLModeDisable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(database.UploadObj)
	db.AutoMigrate(database.UserObj)
	db.AutoMigrate(database.PaymentObj)
	db.Close()
}
