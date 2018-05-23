package server

import (
	"github.com/RTradeLtd/Temporal/bindings/files"
	"github.com/ethereum/go-ethereum/common"
)

// NewFilesContract is used to generate a new files contract handler
func (sm *ServerManager) NewFilesContract(address common.Address) error {
	contract, err := files.NewFiles(address, sm.Client)
	if err != nil {
		return err
	}
	sm.FilesContract = contract
	return nil
}
