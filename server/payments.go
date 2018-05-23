package server

import (
	"errors"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
func (sm *ServerManager) RegisterPaymentForUploader(uploaderAddress string, contentHash string, retentionPeriodInMonths *big.Int, chargeAmountInWei *big.Int, method uint8) (*types.Transaction, error) {
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
	tx, err := sm.PaymentsContract.RegisterPayment(sm.Auth, common.HexToAddress(uploaderAddress), b, retentionPeriodInMonths, chargeAmountInWei, method)
	if err != nil {
		return nil, err
	}
	numPayments, err := sm.PaymentsContract.NumPayments(&bind.CallOpts{Pending: true}, common.HexToAddress(uploaderAddress))
	if err != nil {
		return tx, err
	}
	paymentID, err := sm.PaymentsContract.PaymentIDs(&bind.CallOpts{Pending: true}, common.HexToAddress(uploaderAddress), numPayments)
	if err != nil {
		return tx, err
	}
	pr := queue.PaymentRegister{
		UploaderAddress: uploaderAddress,
		CID:             contentHash,
		HashedCID:       hashedCID.String(),
		PaymentID:       string(paymentID[:]), // this converts the paymentID byte array to a string
	}
	qm, err := queue.Initialize(queue.PaymentRegisterQueue)
	if err != nil {
		return tx, err
	}
	err = qm.PublishMessage(pr)
	if err != nil {
		return tx, err
	}
	return tx, nil
}
