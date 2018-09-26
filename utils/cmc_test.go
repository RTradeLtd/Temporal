package utils_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

func TestRetrieveEthUsdPrice(t *testing.T) {
	price, err := utils.RetrieveEthUsdPrice()
	if err != nil {
		t.Fatal(err)
	}
	if price == 0 {
		t.Fatal("price is 0, unexpected error occured")
	}
}
