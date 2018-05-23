package server

import (
	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/ethereum/go-ethereum/common"
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
