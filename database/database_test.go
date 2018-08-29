package database_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/jinzhu/gorm"
)

const travis = true

var dbPass string

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
	db, err := database.OpenDBConnection(dbPass, "127.0.0.1", "postgres")
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(database.UploadObj)
	db.AutoMigrate(database.UserObj)
	db.AutoMigrate(database.PaymentObj)
	db.Close()
}
