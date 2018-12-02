package mail_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
)

var (
	recipient = "alext@rtradetechnologies.com"
	cfgPath   = filepath.Join(os.Getenv("HOME"), "config.json")
)

func TestMail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.OpenDBConnection(database.DBOptions{
		User:     cfg.Database.Username,
		Password: cfg.Database.Password,
		Address:  cfg.Database.URL,
		Port:     cfg.Database.Port,
	})
	if err != nil {
		t.Fatal(err)
	}
	mm, err := mail.NewManager(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprint("<br>WowSuchEmail<br>WowSuchFormat")
	_, err = mm.SendEmail("testEmail", content, "", "wowmuchemail", recipient)
	if err != nil {
		t.Fatal(err)
	}
}
