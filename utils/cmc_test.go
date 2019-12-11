package utils_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

func TestRetrieveUsdPrice(t *testing.T) {
	type args struct {
		coin string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Ethereum", args{"ethereum"}, false},
		{"Monero", args{"monero"}, false},
		{"Bitcoin", args{"bitcoin"}, false},
		{"Litecoin", args{"litecoin"}, false},
		{"NotARealCoin", args{"NotARealCoin"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := utils.RetrieveUsdPrice(tt.args.coin)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}
			if price == 0 && tt.args.coin != "NotARealCoin" {
				t.Error("price is 0, unexpected result")
			}
			price, err = utils.RetrieveUsdPrice(tt.args.coin)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}
			if price == 0 && tt.args.coin != "NotARealCoin" {
				t.Error("price is 0, unexpected result")
			}
		})
	}
}
