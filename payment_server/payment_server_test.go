package payment_server_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	payment "github.com/RTradeLtd/Temporal/payment_server"
	"github.com/ethereum/go-ethereum/common"
)

var testHash = "0x65a0fc07391cb16f60c589836a72bdeaf2ab8bcb16c819126abba4408e069bcb"
var testHash2 = "0x3c2ce419315b3ab3901bcbe18b1faae8f7b4e237550c745359f0453ac890292c"

func TestHash(t *testing.T) {
	hash := payment.GenerateSignedPaymentMessage(common.HexToAddress("0xa1d9e8788414eA9827f9639c4bd81bA8f3A29758"), uint8(0), big.NewInt(0), big.NewInt(0))
	prefixedHash := fmt.Sprintf("0x%s", hex.EncodeToString(hash))

	if prefixedHash != testHash {
		t.Fatal("invalid hash generated")
	}
	hash = payment.GenerateSignedPaymentMessage(common.HexToAddress("0xa1d9e8788414eA9827f9639c4bd81bA8f3A29758"), uint8(10), big.NewInt(10), big.NewInt(10))
	prefixedHash = fmt.Sprintf("0x%s", hex.EncodeToString(hash))
	if prefixedHash != testHash2 {
		t.Fatal("invalid hash generated")
	}
}
