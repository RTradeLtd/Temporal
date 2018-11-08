package models_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/config"
)

func TestPaymentManager_NewPayment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	pm := models.NewPaymentManager(db)
	type args struct {
		depositAddress string
		txHash         string
		usdValue       float64
		blockchain     string
		paymentType    string
		username       string
	}
	tests := []struct {
		name string
		args args
	}{
		{"Payment1", args{"depositAddress", "txHash", 0.124, "blockchain", "paymentType", "userName"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := pm.NewPayment(
				0,
				tt.args.depositAddress,
				tt.args.txHash,
				tt.args.usdValue,
				tt.args.usdValue,
				tt.args.blockchain,
				tt.args.paymentType,
				tt.args.username,
			)
			if err != nil {
				t.Fatal(err)
			}
			defer pm.DB.Unscoped().Delete(payment)
			lastNumber, err := pm.GetLatestPaymentNumber(tt.args.username)
			if err != nil {
				t.Fatal(err)
			}
			newPaymentNumber := lastNumber + 1
			payment2, err := pm.NewPayment(
				newPaymentNumber,
				tt.args.depositAddress,
				"new tx hash",
				tt.args.usdValue,
				tt.args.usdValue,
				tt.args.blockchain,
				tt.args.paymentType,
				tt.args.username,
			)
			if err != nil {
				t.Fatal(err)
			}
			defer pm.DB.Unscoped().Delete(payment2)
		})
	}
}
