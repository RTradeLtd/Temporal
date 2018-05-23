package server

import (
	"github.com/RTradeLtd/Temporal/bindings/users"
	"github.com/ethereum/go-ethereum/common"
)

// NewUsersContract is used to generate a new users contract handler
func (sm *ServerManager) NewUsersContract(address common.Address) error {
	contract, err := users.NewUsers(address, sm.Client)
	if err != nil {
		return err
	}
	sm.UsersContract = contract
	return nil
}
