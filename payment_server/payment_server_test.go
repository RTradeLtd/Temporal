package payment_server_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	payment "github.com/RTradeLtd/Temporal/payment_server"
	"github.com/ethereum/go-ethereum/common"
)

func TestGenerateSignedPaymentMessage(t *testing.T) {
	address := common.HexToAddress("0xa1d9e8788414eA9827f9639c4bd81bA8f3A29758")
	method := uint8(0)
	number := big.NewInt(0)
	amount := big.NewInt(0)
	hash := payment.GenerateSignedPaymentMessage(address, method, number, amount)
	fmt.Println(hex.EncodeToString(hash))
}
