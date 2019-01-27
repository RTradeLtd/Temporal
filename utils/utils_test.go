package utils_test

import (
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/gorm"
	"github.com/RTradeLtd/rtfs"
)

const (
	testHash       = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"
	nodeOneAPIAddr = "192.168.1.101:5001"
	testSize       = int64(132520817)
)

func TestUtils_CalculatePinCost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}
	manager, err := rtfs.NewManager(nodeOneAPIAddr, "", time.Minute*10, false)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig("../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	usage := models.NewUsageManager(db)
	type args struct {
		username string
		hash     string
		months   int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Pass", args{"testuser", testHash, int64(10)}, false},
		{"Fail", args{"thisusertotallydoesnotexist", testHash, int64(10)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := utils.CalculatePinCost(
				tt.args.username,
				tt.args.hash,
				tt.args.months,
				manager,
				usage,
			); (err != nil) != tt.wantErr {
				t.Fatalf("CalculatePinCost err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUtils_CalculateFileCost(t *testing.T) {
	cfg, err := config.LoadConfig("../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	usage := models.NewUsageManager(db)
	type args struct {
		username string
		size     int64
		months   int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Pass", args{"testuser", testSize, int64(10)}, false},
		{"Fail", args{"totallydoesnotexist", testSize, int64(10)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := utils.CalculateFileCost(
				tt.args.username,
				tt.args.months,
				tt.args.size,
				usage,
			); (err != nil) != tt.wantErr {
				t.Fatalf("CalculateFileCost err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUtils_CalculateAPICallCost(t *testing.T) {
	type args struct {
		callType string
		private  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ipns-public", args{"ipns", false}, false},
		{"ipns-private", args{"ipns", true}, false},
		{"pubsub-public", args{"pubsub", false}, false},
		{"pubsub-private", args{"pubsub", true}, false},
		{"ed25519-public", args{"ed25519", false}, false},
		{"ed25519-private", args{"ed25519", true}, false},
		{"rsa-public", args{"rsa", false}, false},
		{"rsa-private", args{"rsa", true}, false},
		{"invalid", args{"invalid", false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := utils.CalculateAPICallCost(tt.args.callType, tt.args.private)
			if (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
			if cost == 0 && tt.name != "invalid" {
				t.Fatal("invalid cost returned")
			}
		})
	}
}

func TestUtils_FloatToBigInt(t *testing.T) {
	want := int64(500000000000000000)
	bigInt := utils.FloatToBigInt(0.5)
	if bigInt.Int64() != want {
		t.Fatal("failed to properly calculate big int")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	return database.OpenDBConnection(database.DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: true,
	})
}
