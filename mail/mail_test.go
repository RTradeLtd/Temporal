package mail_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
)

var (
	testRecipientEmail = "postables+test@rtradetechnologies.com"
	testRecipientName  = "postables"
	testCfgPath        = "../testenv/config.json"
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
	if cfg.Sendgrid.APIKey == "" {
		cfg.Sendgrid.APIKey = os.Getenv("SENDGRID_API_KEY")
		cfg.Sendgrid.EmailAddress = "temporal@rtradetechnologies.com"
		cfg.Sendgrid.EmailName = "Temporal TravisCI Test"
	}
	mm, err := mail.NewManager(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	/*	if testing.Short() {
			t.Skip("skipping integration test")
		}
	*/
	content := fmt.Sprint("<br>WowSuchEmail<br>WowSuchFormat")
	_, err = mm.SendEmail("testEmail", content, "", testRecipientName, testRecipientEmail)
	if err != nil {
		t.Fatal(err)
	}
}
