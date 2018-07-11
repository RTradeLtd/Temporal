package payment_server_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	payment "github.com/RTradeLtd/Temporal/payment_server"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var testHash = "0x65a0fc07391cb16f60c589836a72bdeaf2ab8bcb16c819126abba4408e069bcb"
var testHash2 = "0x3c2ce419315b3ab3901bcbe18b1faae8f7b4e237550c745359f0453ac890292c"
var key = `{"address":"b2006881d0eb1c6d45a79c065e77fa1b27642277","crypto":{"cipher":"aes-128-ctr","ciphertext":"d77d340f3b7a1f9231901591a9d0eecc6800f8adf619f49831e2188329ae2242","cipherparams":{"iv":"732cc84edfda5ac26a1ba0234d12bf05"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"37df92ae13ccef90682ecc0346e28f58c527476df9497b79f41eaddad9fa1dd9"},"mac":"6b6c0ad5c8cfbe61d646688f602c41f34b6671a359f9275c75ed1083d651c684"},"id":"511a57d3-fe9d-456f-b89c-43467866edb0","version":3}`

func TestSign(t *testing.T) {
	k, err := keystore.DecryptKey([]byte(key), "password123")
	if err != nil {
		t.Fatal(err)
	}
	hash := crypto.Keccak256([]byte("hello"))
	// sig [r,s,v]
	sig, err := crypto.Sign(hash, k.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(sig))
	r := sig[0:32]
	fmt.Println(len(r))
	s := sig[32:64]
	fmt.Println(len(s))
	v := sig[64]
	fmt.Printf("0x%s\n", hex.EncodeToString(hash))
	fmt.Println(v)
	fmt.Printf("0x%s\n", hex.EncodeToString(r))
	fmt.Printf("0x%s\n", hex.EncodeToString(s))
}
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
