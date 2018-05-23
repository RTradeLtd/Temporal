// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// BindingsABI is the input ABI used to generate the binding from.
const BindingsABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_paymentProcessorAddress\",\"type\":\"address\"}],\"name\":\"setPaymentProcessor\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cids\",\"outputs\":[{\"name\":\"hashedCID\",\"type\":\"bytes32\"},{\"name\":\"numberOfTimesUploaded\",\"type\":\"uint256\"},{\"name\":\"removalDate\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_uploader\",\"type\":\"address\"},{\"name\":\"_hashedCID\",\"type\":\"bytes32\"},{\"name\":\"_retentionPeriodInMonths\",\"type\":\"uint256\"}],\"name\":\"addUploaderForCid\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"paymentProcessor\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_paymentProcessorAddress\",\"type\":\"address\"}],\"name\":\"PaymentProcessorSet\",\"type\":\"event\"}]"

// BindingsBin is the compiled bytecode used for deploying new contracts.
const BindingsBin = `[{"constant":false,"inputs":[{"name":"_paymentProcessorAddress","type":"address"}],"name":"setPaymentProcessor","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"bytes32"}],"name":"cids","outputs":[{"name":"hashedCID","type":"bytes32"},{"name":"numberOfTimesUploaded","type":"uint256"},{"name":"removalDate","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_uploader","type":"address"},{"name":"_hashedCID","type":"bytes32"},{"name":"_retentionPeriodInMonths","type":"uint256"}],"name":"addUploaderForCid","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"paymentProcessor","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_paymentProcessorAddress","type":"address"}],"name":"PaymentProcessorSet","type":"event"}]`

// DeployBindings deploys a new Ethereum contract, binding an instance of Bindings to it.
func DeployBindings(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Bindings, error) {
	parsed, err := abi.JSON(strings.NewReader(BindingsABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(BindingsBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Bindings{BindingsCaller: BindingsCaller{contract: contract}, BindingsTransactor: BindingsTransactor{contract: contract}, BindingsFilterer: BindingsFilterer{contract: contract}}, nil
}

// Bindings is an auto generated Go binding around an Ethereum contract.
type Bindings struct {
	BindingsCaller     // Read-only binding to the contract
	BindingsTransactor // Write-only binding to the contract
	BindingsFilterer   // Log filterer for contract events
}

// BindingsCaller is an auto generated read-only Go binding around an Ethereum contract.
type BindingsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BindingsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BindingsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BindingsSession struct {
	Contract     *Bindings         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BindingsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BindingsCallerSession struct {
	Contract *BindingsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// BindingsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BindingsTransactorSession struct {
	Contract     *BindingsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// BindingsRaw is an auto generated low-level Go binding around an Ethereum contract.
type BindingsRaw struct {
	Contract *Bindings // Generic contract binding to access the raw methods on
}

// BindingsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BindingsCallerRaw struct {
	Contract *BindingsCaller // Generic read-only contract binding to access the raw methods on
}

// BindingsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BindingsTransactorRaw struct {
	Contract *BindingsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBindings creates a new instance of Bindings, bound to a specific deployed contract.
func NewBindings(address common.Address, backend bind.ContractBackend) (*Bindings, error) {
	contract, err := bindBindings(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bindings{BindingsCaller: BindingsCaller{contract: contract}, BindingsTransactor: BindingsTransactor{contract: contract}, BindingsFilterer: BindingsFilterer{contract: contract}}, nil
}

// NewBindingsCaller creates a new read-only instance of Bindings, bound to a specific deployed contract.
func NewBindingsCaller(address common.Address, caller bind.ContractCaller) (*BindingsCaller, error) {
	contract, err := bindBindings(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BindingsCaller{contract: contract}, nil
}

// NewBindingsTransactor creates a new write-only instance of Bindings, bound to a specific deployed contract.
func NewBindingsTransactor(address common.Address, transactor bind.ContractTransactor) (*BindingsTransactor, error) {
	contract, err := bindBindings(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BindingsTransactor{contract: contract}, nil
}

// NewBindingsFilterer creates a new log filterer instance of Bindings, bound to a specific deployed contract.
func NewBindingsFilterer(address common.Address, filterer bind.ContractFilterer) (*BindingsFilterer, error) {
	contract, err := bindBindings(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BindingsFilterer{contract: contract}, nil
}

// bindBindings binds a generic wrapper to an already deployed contract.
func bindBindings(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BindingsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bindings *BindingsRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Bindings.Contract.BindingsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bindings *BindingsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bindings.Contract.BindingsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bindings *BindingsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bindings.Contract.BindingsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bindings *BindingsCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Bindings.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bindings *BindingsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bindings.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bindings *BindingsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bindings.Contract.contract.Transact(opts, method, params...)
}

// Cids is a free data retrieval call binding the contract method 0x44e0a526.
//
// Solidity: function cids( bytes32) constant returns(hashedCID bytes32, numberOfTimesUploaded uint256, removalDate uint256)
func (_Bindings *BindingsCaller) Cids(opts *bind.CallOpts, arg0 [32]byte) (struct {
	HashedCID             [32]byte
	NumberOfTimesUploaded *big.Int
	RemovalDate           *big.Int
}, error) {
	ret := new(struct {
		HashedCID             [32]byte
		NumberOfTimesUploaded *big.Int
		RemovalDate           *big.Int
	})
	out := ret
	err := _Bindings.contract.Call(opts, out, "cids", arg0)
	return *ret, err
}

// Cids is a free data retrieval call binding the contract method 0x44e0a526.
//
// Solidity: function cids( bytes32) constant returns(hashedCID bytes32, numberOfTimesUploaded uint256, removalDate uint256)
func (_Bindings *BindingsSession) Cids(arg0 [32]byte) (struct {
	HashedCID             [32]byte
	NumberOfTimesUploaded *big.Int
	RemovalDate           *big.Int
}, error) {
	return _Bindings.Contract.Cids(&_Bindings.CallOpts, arg0)
}

// Cids is a free data retrieval call binding the contract method 0x44e0a526.
//
// Solidity: function cids( bytes32) constant returns(hashedCID bytes32, numberOfTimesUploaded uint256, removalDate uint256)
func (_Bindings *BindingsCallerSession) Cids(arg0 [32]byte) (struct {
	HashedCID             [32]byte
	NumberOfTimesUploaded *big.Int
	RemovalDate           *big.Int
}, error) {
	return _Bindings.Contract.Cids(&_Bindings.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Bindings *BindingsCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Bindings.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Bindings *BindingsSession) Owner() (common.Address, error) {
	return _Bindings.Contract.Owner(&_Bindings.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Bindings *BindingsCallerSession) Owner() (common.Address, error) {
	return _Bindings.Contract.Owner(&_Bindings.CallOpts)
}

// PaymentProcessor is a free data retrieval call binding the contract method 0xf1c6bdf8.
//
// Solidity: function paymentProcessor() constant returns(address)
func (_Bindings *BindingsCaller) PaymentProcessor(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Bindings.contract.Call(opts, out, "paymentProcessor")
	return *ret0, err
}

// PaymentProcessor is a free data retrieval call binding the contract method 0xf1c6bdf8.
//
// Solidity: function paymentProcessor() constant returns(address)
func (_Bindings *BindingsSession) PaymentProcessor() (common.Address, error) {
	return _Bindings.Contract.PaymentProcessor(&_Bindings.CallOpts)
}

// PaymentProcessor is a free data retrieval call binding the contract method 0xf1c6bdf8.
//
// Solidity: function paymentProcessor() constant returns(address)
func (_Bindings *BindingsCallerSession) PaymentProcessor() (common.Address, error) {
	return _Bindings.Contract.PaymentProcessor(&_Bindings.CallOpts)
}

// AddUploaderForCid is a paid mutator transaction binding the contract method 0x6eb033f4.
//
// Solidity: function addUploaderForCid(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256) returns(bool)
func (_Bindings *BindingsTransactor) AddUploaderForCid(opts *bind.TransactOpts, _uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "addUploaderForCid", _uploader, _hashedCID, _retentionPeriodInMonths)
}

// AddUploaderForCid is a paid mutator transaction binding the contract method 0x6eb033f4.
//
// Solidity: function addUploaderForCid(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256) returns(bool)
func (_Bindings *BindingsSession) AddUploaderForCid(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.AddUploaderForCid(&_Bindings.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths)
}

// AddUploaderForCid is a paid mutator transaction binding the contract method 0x6eb033f4.
//
// Solidity: function addUploaderForCid(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256) returns(bool)
func (_Bindings *BindingsTransactorSession) AddUploaderForCid(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.AddUploaderForCid(&_Bindings.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths)
}

// SetPaymentProcessor is a paid mutator transaction binding the contract method 0x088b0d75.
//
// Solidity: function setPaymentProcessor(_paymentProcessorAddress address) returns(bool)
func (_Bindings *BindingsTransactor) SetPaymentProcessor(opts *bind.TransactOpts, _paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setPaymentProcessor", _paymentProcessorAddress)
}

// SetPaymentProcessor is a paid mutator transaction binding the contract method 0x088b0d75.
//
// Solidity: function setPaymentProcessor(_paymentProcessorAddress address) returns(bool)
func (_Bindings *BindingsSession) SetPaymentProcessor(_paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetPaymentProcessor(&_Bindings.TransactOpts, _paymentProcessorAddress)
}

// SetPaymentProcessor is a paid mutator transaction binding the contract method 0x088b0d75.
//
// Solidity: function setPaymentProcessor(_paymentProcessorAddress address) returns(bool)
func (_Bindings *BindingsTransactorSession) SetPaymentProcessor(_paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetPaymentProcessor(&_Bindings.TransactOpts, _paymentProcessorAddress)
}

// BindingsPaymentProcessorSetIterator is returned from FilterPaymentProcessorSet and is used to iterate over the raw logs and unpacked data for PaymentProcessorSet events raised by the Bindings contract.
type BindingsPaymentProcessorSetIterator struct {
	Event *BindingsPaymentProcessorSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BindingsPaymentProcessorSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsPaymentProcessorSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BindingsPaymentProcessorSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BindingsPaymentProcessorSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsPaymentProcessorSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsPaymentProcessorSet represents a PaymentProcessorSet event raised by the Bindings contract.
type BindingsPaymentProcessorSet struct {
	PaymentProcessorAddress common.Address
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterPaymentProcessorSet is a free log retrieval operation binding the contract event 0x8985ba152810f26e00d857552aaff0a9cae15a4fe9bbeb9ec0be19e3e1f064db.
//
// Solidity: e PaymentProcessorSet(_paymentProcessorAddress address)
func (_Bindings *BindingsFilterer) FilterPaymentProcessorSet(opts *bind.FilterOpts) (*BindingsPaymentProcessorSetIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "PaymentProcessorSet")
	if err != nil {
		return nil, err
	}
	return &BindingsPaymentProcessorSetIterator{contract: _Bindings.contract, event: "PaymentProcessorSet", logs: logs, sub: sub}, nil
}

// WatchPaymentProcessorSet is a free log subscription operation binding the contract event 0x8985ba152810f26e00d857552aaff0a9cae15a4fe9bbeb9ec0be19e3e1f064db.
//
// Solidity: e PaymentProcessorSet(_paymentProcessorAddress address)
func (_Bindings *BindingsFilterer) WatchPaymentProcessorSet(opts *bind.WatchOpts, sink chan<- *BindingsPaymentProcessorSet) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "PaymentProcessorSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsPaymentProcessorSet)
				if err := _Bindings.contract.UnpackLog(event, "PaymentProcessorSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
