package mail_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/mail"
)

func TestMail(t *testing.T) {
	cfg, err := config.LoadConfig("/home/solidity/config.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = mail.GenerateMailManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
}
