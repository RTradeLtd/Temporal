package mail_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
)

var (
	testRecipientEmail1 = "postables+test1@rtradetechnologies.com"
	testRecipientEmail2 = "postables+test2@rtradetechnologies.com"
	testRecipientName1  = "postables1"
	testRecipientName2  = "postables2"
	testCfgPath         = "../testenv/config.json"
)

func TestMail(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dbm, err := database.New(cfg, database.Options{SSLModeDisable: true})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Sendgrid.APIKey == "" {
		cfg.Sendgrid.APIKey = os.Getenv("SENDGRID_API_KEY")
		cfg.Sendgrid.EmailAddress = "temporal@rtradetechnologies.com"
		cfg.Sendgrid.EmailName = "Temporal TravisCI Test"
	}
	mm, err := mail.NewManager(cfg, dbm.DB)
	if err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprint("<br>WowSuchEmail<br>WowSuchFormat")
	if _, err = mm.SendEmail(
		"testEmail",
		content,
		"",
		testRecipientName1,
		testRecipientEmail1,
	); err != nil {
		t.Fatal(err)
	}
	if _, err = mm.SendEmail(
		"testEmail",
		content,
		"text/html",
		testRecipientName1,
		testRecipientEmail1,
	); err != nil {
		t.Fatal(err)
	}
	if err = mm.BulkSend(
		"testEmail",
		content,
		"text/html",
		[]string{testRecipientName1, testRecipientName2},
		[]string{testRecipientEmail1, testRecipientEmail2},
	); err != nil {
		t.Fatal(err)
	}
}
