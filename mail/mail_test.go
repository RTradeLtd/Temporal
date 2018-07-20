package mail_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/mail"
)

var recipient = "insertemailhere"

func TestMail(t *testing.T) {
	cfg, err := config.LoadConfig("/home/solidity/config.json")
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
