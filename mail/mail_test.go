package mail_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
)

var (
	testRecipient = "test@example.com"
	testCfgPath   = "../testenv/config.json"
)

func TestMail(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.OpenDBConnection(database.DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	mm, err := mail.NewManager(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	content := fmt.Sprint("<br>WowSuchEmail<br>WowSuchFormat")
	_, err = mm.SendEmail("testEmail", content, "", "wowmuchemail", testRecipient)
	if err != nil {
		t.Fatal(err)
	}
}
