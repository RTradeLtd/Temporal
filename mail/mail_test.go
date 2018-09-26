package mail_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/config"
)

var (
	recipient = "insertemailhere"
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
	mm, err := mail.GenerateMailManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprint("<br>WowSuchEmail<br>WowSuchFormat")
	_, err = mm.SendEmail("testEmail", content, "", "wowmuchemail", recipient)
	if err != nil {
		t.Fatal(err)
	}
}
