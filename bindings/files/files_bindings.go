package files

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

// FilesABI is the input ABI used to generate the binding from.
const FilesABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_paymentProcessorAddress\",\"type\":\"address\"}],\"name\":\"setPaymentProcessor\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cids\",\"outputs\":[{\"name\":\"hashedCID\",\"type\":\"bytes32\"},{\"name\":\"numberOfTimesUploaded\",\"type\":\"uint256\"},{\"name\":\"removalDate\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_uploader\",\"type\":\"address\"},{\"name\":\"_hashedCID\",\"type\":\"bytes32\"},{\"name\":\"_retentionPeriodInMonths\",\"type\":\"uint256\"}],\"name\":\"addUploaderForCid\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"paymentProcessor\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_paymentProcessorAddress\",\"type\":\"address\"}],\"name\":\"PaymentProcessorSet\",\"type\":\"event\"}]"

// FilesBin is the compiled bytecode used for deploying new contracts.
const FilesBin = `[{"constant":false,"inputs":[{"name":"_paymentProcessorAddress","type":"address"}],"name":"setPaymentProcessor","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"bytes32"}],"name":"cids","outputs":[{"name":"hashedCID","type":"bytes32"},{"name":"numberOfTimesUploaded","type":"uint256"},{"name":"removalDate","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_uploader","type":"address"},{"name":"_hashedCID","type":"bytes32"},{"name":"_retentionPeriodInMonths","type":"uint256"}],"name":"addUploaderForCid","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"paymentProcessor","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_paymentProcessorAddress","type":"address"}],"name":"PaymentProcessorSet","type":"event"}]`

// DeployFiles deploys a new Ethereum contract, binding an instance of Files to it.
func DeployFiles(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Files, error) {
	parsed, err := abi.JSON(strings.NewReader(FilesABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(FilesBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Files{FilesCaller: FilesCaller{contract: contract}, FilesTransactor: FilesTransactor{contract: contract}, FilesFilterer: FilesFilterer{contract: contract}}, nil
}

// Files is an auto generated Go binding around an Ethereum contract.
type Files struct {
	FilesCaller     // Read-only binding to the contract
	FilesTransactor // Write-only binding to the contract
	FilesFilterer   // Log filterer for contract events
}

// FilesCaller is an auto generated read-only Go binding around an Ethereum contract.
type FilesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FilesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FilesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FilesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FilesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FilesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FilesSession struct {
	Contract     *Files            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FilesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FilesCallerSession struct {
	Contract *FilesCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// FilesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FilesTransactorSession struct {
	Contract     *FilesTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FilesRaw is an auto generated low-level Go binding around an Ethereum contract.
type FilesRaw struct {
	Contract *Files // Generic contract binding to access the raw methods on
}

// FilesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FilesCallerRaw struct {
	Contract *FilesCaller // Generic read-only contract binding to access the raw methods on
}

// FilesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FilesTransactorRaw struct {
	Contract *FilesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFiles creates a new instance of Files, bound to a specific deployed contract.
func NewFiles(address common.Address, backend bind.ContractBackend) (*Files, error) {
	contract, err := bindFiles(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Files{FilesCaller: FilesCaller{contract: contract}, FilesTransactor: FilesTransactor{contract: contract}, FilesFilterer: FilesFilterer{contract: contract}}, nil
}

// NewFilesCaller creates a new read-only instance of Files, bound to a specific deployed contract.
func NewFilesCaller(address common.Address, caller bind.ContractCaller) (*FilesCaller, error) {
	contract, err := bindFiles(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FilesCaller{contract: contract}, nil
}

// NewFilesTransactor creates a new write-only instance of Files, bound to a specific deployed contract.
func NewFilesTransactor(address common.Address, transactor bind.ContractTransactor) (*FilesTransactor, error) {
	contract, err := bindFiles(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FilesTransactor{contract: contract}, nil
}

// NewFilesFilterer creates a new log filterer instance of Files, bound to a specific deployed contract.
func NewFilesFilterer(address common.Address, filterer bind.ContractFilterer) (*FilesFilterer, error) {
	contract, err := bindFiles(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FilesFilterer{contract: contract}, nil
}

// bindFiles binds a generic wrapper to an already deployed contract.
func bindFiles(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FilesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Files *FilesRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Files.Contract.FilesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Files *FilesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Files.Contract.FilesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Files *FilesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Files.Contract.FilesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Files *FilesCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Files.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Files *FilesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Files.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Files *FilesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Files.Contract.contract.Transact(opts, method, params...)
}

// Cids is a free data retrieval call binding the contract method 0x44e0a526.
//
// Solidity: function cids( bytes32) constant returns(hashedCID bytes32, numberOfTimesUploaded uint256, removalDate uint256)
func (_Files *FilesCaller) Cids(opts *bind.CallOpts, arg0 [32]byte) (struct {
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
	err := _Files.contract.Call(opts, out, "cids", arg0)
	return *ret, err
}

// Cids is a free data retrieval call binding the contract method 0x44e0a526.
//
// Solidity: function cids( bytes32) constant returns(hashedCID bytes32, numberOfTimesUploaded uint256, removalDate uint256)
func (_Files *FilesSession) Cids(arg0 [32]byte) (struct {
	HashedCID             [32]byte
	NumberOfTimesUploaded *big.Int
	RemovalDate           *big.Int
}, error) {
	return _Files.Contract.Cids(&_Files.CallOpts, arg0)
}

// Cids is a free data retrieval call binding the contract method 0x44e0a526.
//
// Solidity: function cids( bytes32) constant returns(hashedCID bytes32, numberOfTimesUploaded uint256, removalDate uint256)
func (_Files *FilesCallerSession) Cids(arg0 [32]byte) (struct {
	HashedCID             [32]byte
	NumberOfTimesUploaded *big.Int
	RemovalDate           *big.Int
}, error) {
	return _Files.Contract.Cids(&_Files.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Files *FilesCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Files.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Files *FilesSession) Owner() (common.Address, error) {
	return _Files.Contract.Owner(&_Files.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Files *FilesCallerSession) Owner() (common.Address, error) {
	return _Files.Contract.Owner(&_Files.CallOpts)
}

// PaymentProcessor is a free data retrieval call binding the contract method 0xf1c6bdf8.
//
// Solidity: function paymentProcessor() constant returns(address)
func (_Files *FilesCaller) PaymentProcessor(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Files.contract.Call(opts, out, "paymentProcessor")
	return *ret0, err
}

// PaymentProcessor is a free data retrieval call binding the contract method 0xf1c6bdf8.
//
// Solidity: function paymentProcessor() constant returns(address)
func (_Files *FilesSession) PaymentProcessor() (common.Address, error) {
	return _Files.Contract.PaymentProcessor(&_Files.CallOpts)
}

// PaymentProcessor is a free data retrieval call binding the contract method 0xf1c6bdf8.
//
// Solidity: function paymentProcessor() constant returns(address)
func (_Files *FilesCallerSession) PaymentProcessor() (common.Address, error) {
	return _Files.Contract.PaymentProcessor(&_Files.CallOpts)
}

// AddUploaderForCid is a paid mutator transaction binding the contract method 0x6eb033f4.
//
// Solidity: function addUploaderForCid(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256) returns(bool)
func (_Files *FilesTransactor) AddUploaderForCid(opts *bind.TransactOpts, _uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int) (*types.Transaction, error) {
	return _Files.contract.Transact(opts, "addUploaderForCid", _uploader, _hashedCID, _retentionPeriodInMonths)
}

// AddUploaderForCid is a paid mutator transaction binding the contract method 0x6eb033f4.
//
// Solidity: function addUploaderForCid(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256) returns(bool)
func (_Files *FilesSession) AddUploaderForCid(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int) (*types.Transaction, error) {
	return _Files.Contract.AddUploaderForCid(&_Files.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths)
}

// AddUploaderForCid is a paid mutator transaction binding the contract method 0x6eb033f4.
//
// Solidity: function addUploaderForCid(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256) returns(bool)
func (_Files *FilesTransactorSession) AddUploaderForCid(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int) (*types.Transaction, error) {
	return _Files.Contract.AddUploaderForCid(&_Files.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths)
}

// SetPaymentProcessor is a paid mutator transaction binding the contract method 0x088b0d75.
//
// Solidity: function setPaymentProcessor(_paymentProcessorAddress address) returns(bool)
func (_Files *FilesTransactor) SetPaymentProcessor(opts *bind.TransactOpts, _paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Files.contract.Transact(opts, "setPaymentProcessor", _paymentProcessorAddress)
}

// SetPaymentProcessor is a paid mutator transaction binding the contract method 0x088b0d75.
//
// Solidity: function setPaymentProcessor(_paymentProcessorAddress address) returns(bool)
func (_Files *FilesSession) SetPaymentProcessor(_paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Files.Contract.SetPaymentProcessor(&_Files.TransactOpts, _paymentProcessorAddress)
}

// SetPaymentProcessor is a paid mutator transaction binding the contract method 0x088b0d75.
//
// Solidity: function setPaymentProcessor(_paymentProcessorAddress address) returns(bool)
func (_Files *FilesTransactorSession) SetPaymentProcessor(_paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Files.Contract.SetPaymentProcessor(&_Files.TransactOpts, _paymentProcessorAddress)
}

// FilesPaymentProcessorSetIterator is returned from FilterPaymentProcessorSet and is used to iterate over the raw logs and unpacked data for PaymentProcessorSet events raised by the Files contract.
type FilesPaymentProcessorSetIterator struct {
	Event *FilesPaymentProcessorSet // Event containing the contract specifics and raw log

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
func (it *FilesPaymentProcessorSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FilesPaymentProcessorSet)
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
		it.Event = new(FilesPaymentProcessorSet)
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
func (it *FilesPaymentProcessorSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FilesPaymentProcessorSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FilesPaymentProcessorSet represents a PaymentProcessorSet event raised by the Files contract.
type FilesPaymentProcessorSet struct {
	PaymentProcessorAddress common.Address
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterPaymentProcessorSet is a free log retrieval operation binding the contract event 0x8985ba152810f26e00d857552aaff0a9cae15a4fe9bbeb9ec0be19e3e1f064db.
//
// Solidity: e PaymentProcessorSet(_paymentProcessorAddress address)
func (_Files *FilesFilterer) FilterPaymentProcessorSet(opts *bind.FilterOpts) (*FilesPaymentProcessorSetIterator, error) {

	logs, sub, err := _Files.contract.FilterLogs(opts, "PaymentProcessorSet")
	if err != nil {
		return nil, err
	}
	return &FilesPaymentProcessorSetIterator{contract: _Files.contract, event: "PaymentProcessorSet", logs: logs, sub: sub}, nil
}

// WatchPaymentProcessorSet is a free log subscription operation binding the contract event 0x8985ba152810f26e00d857552aaff0a9cae15a4fe9bbeb9ec0be19e3e1f064db.
//
// Solidity: e PaymentProcessorSet(_paymentProcessorAddress address)
func (_Files *FilesFilterer) WatchPaymentProcessorSet(opts *bind.WatchOpts, sink chan<- *FilesPaymentProcessorSet) (event.Subscription, error) {

	logs, sub, err := _Files.contract.WatchLogs(opts, "PaymentProcessorSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FilesPaymentProcessorSet)
				if err := _Files.contract.UnpackLog(event, "PaymentProcessorSet", log); err != nil {
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
