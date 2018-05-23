package server

import (
	"errors"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// NewPaymentsContract is used to generate a new payment contract handler
func (sm *ServerManager) NewPaymentsContract(address common.Address) error {
	contract, err := payments.NewPayments(address, sm.Client)
	if err != nil {
		return err
	}
	sm.PaymentsContract = contract
	return nil
}

// RegisterPaymentForUploader is used to register a payment for the given uploader
func (sm *ServerManager) RegisterPaymentForUploader(uploaderAddress common.Address, contentHash string, retentionPeriodInMonths *big.Int, chargeAmountInWei *big.Int, method uint8) (*types.Transaction, error) {
	if method > 1 || method < 0 {
		return nil, errors.New("invalid payment method. 0 = RTC, 1 = ETH")
	}
	var b [32]byte
	// convert hash to byte slice
	data := []byte(contentHash)
	// runs the byte slice through keccak256
	hashedCIDByte := crypto.Keccak256(data)
	// convert to hash
	hashedCID := common.BytesToHash(hashedCIDByte)
	// convert byte slice to byte array
	copy(b[:], hashedCID.Bytes()[:32])
	tx, err := sm.PaymentsContract.RegisterPayment(sm.Auth, uploaderAddress, b, retentionPeriodInMonths, chargeAmountInWei, method)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
