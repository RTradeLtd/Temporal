package utils_test

import (
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/jinzhu/gorm"
	"github.com/RTradeLtd/rtfs/v2"
)

const (
	testHash       = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	nodeOneAPIAddr = "192.168.1.101:5001"
	testSize       = int64(132520817)
)

func TestUtils_CalculatePinCost(t *testing.T) {
	manager, err := rtfs.NewManager(nodeOneAPIAddr, "", time.Minute*60)
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
		{"Fail", args{"thisusertotallydoesnotexist", testHash, int64(10)}, true},
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

func TestUtils_FloatToBigInt(t *testing.T) {
	want := int64(500000000000000000)
	bigInt := utils.FloatToBigInt(0.5)
	if bigInt.Int64() != want {
		t.Fatal("failed to properly calculate big int")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	dbm, err := database.New(cfg, database.Options{SSLModeDisable: true})
	if err != nil {
		return nil, err
	}
	return dbm.DB, nil
}
