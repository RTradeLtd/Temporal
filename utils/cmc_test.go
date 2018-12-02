package utils_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

func TestUtils_RetrieveEthUsdPrice(t *testing.T) {
	priceFloat, err := utils.RetrieveEthUsdPrice()
	if err != nil {
		t.Fatal(err)
	}
	if priceFloat == 0 {
		t.Fatal("priceFloat is 0, unexpected error occurred")
	}
	priceInt, err := utils.RetrieveEthUsdPriceNoDecimals()
	if err != nil {
		t.Fatal(err)
	}
	if priceInt == 0 {
		t.Fatal("priceInt is 0, unexpected error occurred")
	}
}

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
		})
	}
}
