package utils_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

func TestUtils_RetrieveEthUsdPrice(t *testing.T) {
	priceFloat, err := utils.RetrieveEthUsdPrice()
	if err != nil {
		t.Fatal(err)
	}
	if priceFloat == 0 {
		t.Fatal("priceFloat is 0, unexpected error occured")
	}
	priceInt, err := utils.RetrieveEthUsdPriceNoDecimals()
	if err != nil {
		t.Fatal(err)
	}
	if priceInt == 0 {
		t.Fatal("priceInt is 0, unexpected error occured")
	}
}

func TestRetrieveUsdPrice(t *testing.T) {
	type args struct {
		coin string
	}
	tests := []struct {
		name string
		args args
	}{
		{"Ethereum", args{"ethereum"}},
		{"Monero", args{"monero"}},
		{"Bitcoin", args{"bitcoin"}},
		{"Litecoin", args{"litecoin"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := utils.RetrieveUsdPrice(tt.args.coin)
			if err != nil {
				t.Error(err)
			}
			if price == 0 {
				t.Error("price is 0, unexpected result")
			}
			fmt.Println(price)
		})
	}
}
