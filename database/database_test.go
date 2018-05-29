package database_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/database"
)

const travis = true

var dbPass string

func TestDatabase(t *testing.T) {
	if !travis {
		dbPass = "password123"
	} else {
		dbPass = ""
	}
	db := database.OpenDBConnection(dbPass)
	db.Close()
}
