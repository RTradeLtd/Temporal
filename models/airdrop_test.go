package models_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/config"
)

func TestDrop(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	dropManager := models.NewDropManager(db)
	type args struct {
		username   string
		dropID     string
		ethAddress string
	}
	tests := []struct {
		name string
		args args
	}{
		{"User1", args{"user1", "drop1", "0xD6e33C11CFF866162787b7198030aaC101A61F29"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drop1, err := dropManager.RegisterAirDrop(
				tt.args.dropID,
				tt.args.ethAddress,
				tt.args.username,
			)
			if err != nil {
				t.Fatal(err)
			}
			drop2, err := dropManager.FindByDropID(drop1.DropID)
			if err != nil {
				t.Fatal(err)
			}
			if drop2.UserName != drop1.UserName {
				t.Fatal("bad username recovered")
			}
			drop3, err := dropManager.FindByEthAddress(drop1.EthAddress)
			if err != nil {
				t.Fatal(err)
			}
			if drop3.EthAddress != drop1.EthAddress {
				t.Fatal("bad eth address recovered")
			}
			if _, err := dropManager.RegisterAirDrop(
				tt.args.dropID,
				tt.args.ethAddress,
				tt.args.username,
			); err == nil {
				t.Fatal("no error received when one should've been")
			}
		})
	}
}
