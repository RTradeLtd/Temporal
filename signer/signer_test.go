package signer_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var (
	path    = "/home/solidity/.ethereum/keystore/UTC--2018-07-10T00-50-21.775537101Z--41943d1023f899e3637c2ca5bba274f50b7306e6"
	pass    = "password123"
	address = common.HexToAddress("0x20BBcc07CFfa73Df6Bf168265980EDF9eB2376fE")
	method  = uint8(0)
	number  = big.NewInt(0)
	amount  = big.NewInt(0)
)

/*
Used to test the signer package
func TestSigner(t *testing.T) {
	s, err := signer.GeneratePaymentSigner(path, pass)
	if err != nil {
		t.Fatal(err)
	}
	sm, err := s.GenerateSignedPaymentMessage(address, method, number, amount)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", sm)
}
*/
