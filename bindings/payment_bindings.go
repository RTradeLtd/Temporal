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
const BindingsABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"payments\",\"outputs\":[{\"name\":\"uploader\",\"type\":\"address\"},{\"name\":\"paymentID\",\"type\":\"bytes32\"},{\"name\":\"hashedCID\",\"type\":\"bytes32\"},{\"name\":\"retentionPeriodInMonths\",\"type\":\"uint256\"},{\"name\":\"paymentAmount\",\"type\":\"uint256\"},{\"name\":\"state\",\"type\":\"uint8\"},{\"name\":\"method\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"numPayments\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_filesContractAddress\",\"type\":\"address\"}],\"name\":\"setFilesInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"payRtcForPaymentID\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_usersContractAddress\",\"type\":\"address\"}],\"name\":\"setUsersInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_uploader\",\"type\":\"address\"},{\"name\":\"_hashedCID\",\"type\":\"bytes32\"},{\"name\":\"_retentionPeriodInMonths\",\"type\":\"uint256\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_method\",\"type\":\"uint8\"}],\"name\":\"registerPayment\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newAdmin\",\"type\":\"address\"}],\"name\":\"changeAdmin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"payEthForPaymentID\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"fI\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"uI\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_filesContractAddress\",\"type\":\"address\"}],\"name\":\"FilesContractSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_usersContractAddress\",\"type\":\"address\"}],\"name\":\"UsersContractSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_hashedCID\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_retentionPeriodInMonths\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_paymentID\",\"type\":\"bytes32\"}],\"name\":\"PaymentRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"EthPaymentReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"RtcPaymentReceived\",\"type\":\"event\"}]"

// BindingsBin is the compiled bytecode used for deploying new contracts.
const BindingsBin = `60c0604052601c60808190527f19457468657265756d205369676e6564204d6573736167653a0a33320000000060a090815261003e9160029190610066565b506000805433600160a060020a03199182168117909255600180549091169091179055610101565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100a757805160ff19168380011785556100d4565b828001600101855582156100d4579182015b828111156100d45782518255916020019190600101906100b9565b506100e09291506100e4565b5090565b6100fe91905b808211156100e057600081556001016100ea565b90565b610c5c806101106000396000f3006080604052600436106100b95763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630716326d81146100be5780630858830b1461013f5780631308bb661461017257806357a5d426146101a75780635b8e47af146101c25780638da5cb5b146101e35780638ebed58f146102145780638f283970146102445780639818acb914610265578063db8ce76c14610280578063ebbec31e14610295578063f851a440146102aa575b600080fd5b3480156100ca57600080fd5b506100d66004356102bf565b60408051600160a060020a038916815260208101889052908101869052606081018590526080810184905260a0810183600281111561011157fe5b60ff16815260200182600181111561012557fe5b60ff16815260200197505050505050505060405180910390f35b34801561014b57600080fd5b50610160600160a060020a036004351661030a565b60408051918252519081900360200190f35b34801561017e57600080fd5b50610193600160a060020a036004351661031c565b604080519115158252519081900360200190f35b3480156101b357600080fd5b506101936004356024356103ce565b3480156101ce57600080fd5b50610193600160a060020a036004351661065d565b3480156101ef57600080fd5b506101f86106f8565b60408051600160a060020a039092168252519081900360200190f35b34801561022057600080fd5b50610193600160a060020a036004351660243560443560643560ff60843516610707565b34801561025057600080fd5b50610193600160a060020a036004351661096c565b34801561027157600080fd5b506101936004356024356109bb565b34801561028c57600080fd5b506101f8610c03565b3480156102a157600080fd5b506101f8610c12565b3480156102b657600080fd5b506101f8610c21565b6005602081905260009182526040909120805460018201546002830154600384015460048501549490950154600160a060020a03909316949193909260ff8082169161010090041687565b60066020526000908152604090205481565b600080543390600160a060020a03168114806103455750600154600160a060020a038281169116145b151561035057600080fd5b82600160a060020a038116151561036657600080fd5b60048054600160a060020a03861673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f57df6050063bfc7245fb45847eab30542686438bc930cf2f1d0947158615071c9181900360200190a15060019392505050565b60008260016000828152600560208190526040909120015460ff1660028111156103f457fe5b146103fe57600080fd5b60008481526005602052604090205484903390600160a060020a0316811461042557600080fd5b60008681526005602052604090206004015486908690811461044657600080fd5b8760008060008381526005602081905260409091200154610100900460ff16600181111561047057fe5b1461047a57600080fd5b60008a815260056020818152604092839020909101805460ff1916600217905581518c81529081018b9052815133927f71536f9dd7c4e8db4b8cb8226889aaea1c562bca1da233f3e0e7846f4e65d57b928290030190a26004805460008c8152600560209081526040808320600281015460039091015482517f6eb033f400000000000000000000000000000000000000000000000000000000815233978101979097526024870191909152604486015251600160a060020a0390931693636eb033f49360648083019491928390030190829087803b15801561055c57600080fd5b505af1158015610570573d6000803e3d6000fd5b505050506040513d602081101561058657600080fd5b5051151561059357600080fd5b60035460008b81526005602090815260408083206002015481517f66a04e91000000000000000000000000000000000000000000000000000000008152336004820152602481018f905260448101919091529051600160a060020a03909416936366a04e9193606480840194938390030190829087803b15801561061657600080fd5b505af115801561062a573d6000803e3d6000fd5b505050506040513d602081101561064057600080fd5b5051151561064d57600080fd5b5060019998505050505050505050565b600080543390600160a060020a03168114806106865750600154600160a060020a038281169116145b151561069157600080fd5b60038054600160a060020a03851673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f0c7303206058ab0e0d85e1f17933330be8aba69aa937a13044b6c10e40886e209181900360200190a150600192915050565b600054600160a060020a031681565b6000805481903390600160a060020a03168114806107325750600154600160a060020a038281169116145b151561073d57600080fd5b846000811161074b57600080fd5b866000811161075957600080fd5b89600160a060020a038116151561076f57600080fd5b60018760ff161180610784575060008760ff16105b1561078e57600080fd5b600160a060020a038b166000818152600660205260408082205481516c010000000000000000000000009094028452601484018e9052603484018d90526054840152519182900360740190912095506000868152600560208190526040909120015460ff1660028111156107fe57fe5b1461080857600080fd5b6040805160e081018252600160a060020a038d168152602081018790529081018b9052606081018a90526080810189905260a08101600181526020018860ff16600181111561085357fe5b600181111561085e57fe5b90526000868152600560208181526040928390208451815473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a039091161781559084015160018083019190915592840151600280830191909155606085015160038301556080850151600483015560a085015192820180549294909260ff19169184908111156108e757fe5b021790555060c082015160058201805461ff00191661010083600181111561090b57fe5b021790555050604080518c8152602081018c90528082018b90526060810188905290513392507fcb6bac54d8308a9609a62a77e0389ef2b7d019f81686bfb4e556643f70110f7a9181900360800190a25060019a9950505050505050505050565b600080548290600160a060020a0380831691161461098957600080fd5b60018054600160a060020a03851673ffffffffffffffffffffffffffffffffffffffff19909116178155915050919050565b60008260016000828152600560208190526040909120015460ff1660028111156109e157fe5b146109eb57600080fd5b60008481526005602052604090205484903390600160a060020a03168114610a1257600080fd5b600086815260056020526040902060040154869086908114610a3357600080fd5b8760018060008381526005602081905260409091200154610100900460ff166001811115610a5d57fe5b14610a6757600080fd5b60008a815260056020818152604092839020909101805460ff1916600217905581518c81529081018b9052815133927f9d04fb9b0795e110b6f4b206c3f8c1e6767ac5cf7fd560647ccf3ffd6c9dd5ee928290030190a26004805460008c8152600560209081526040808320600281015460039091015482517f6eb033f400000000000000000000000000000000000000000000000000000000815233978101979097526024870191909152604486015251600160a060020a0390931693636eb033f49360648083019491928390030190829087803b158015610b4957600080fd5b505af1158015610b5d573d6000803e3d6000fd5b505050506040513d6020811015610b7357600080fd5b50511515610b8057600080fd5b60035460008b81526005602090815260408083206002015481517f70b9c01e000000000000000000000000000000000000000000000000000000008152336004820152602481018f905260448101919091529051600160a060020a03909416936370b9c01e93606480840194938390030190829087803b15801561061657600080fd5b600454600160a060020a031681565b600354600160a060020a031681565b600154600160a060020a0316815600a165627a7a72305820e60cdd7b632873f79e8cb26e74393fd1230b42173eb7d250fd225a45456ea3e50029`

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

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Bindings *BindingsCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Bindings.contract.Call(opts, out, "admin")
	return *ret0, err
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Bindings *BindingsSession) Admin() (common.Address, error) {
	return _Bindings.Contract.Admin(&_Bindings.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Bindings *BindingsCallerSession) Admin() (common.Address, error) {
	return _Bindings.Contract.Admin(&_Bindings.CallOpts)
}

// FI is a free data retrieval call binding the contract method 0xdb8ce76c.
//
// Solidity: function fI() constant returns(address)
func (_Bindings *BindingsCaller) FI(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Bindings.contract.Call(opts, out, "fI")
	return *ret0, err
}

// FI is a free data retrieval call binding the contract method 0xdb8ce76c.
//
// Solidity: function fI() constant returns(address)
func (_Bindings *BindingsSession) FI() (common.Address, error) {
	return _Bindings.Contract.FI(&_Bindings.CallOpts)
}

// FI is a free data retrieval call binding the contract method 0xdb8ce76c.
//
// Solidity: function fI() constant returns(address)
func (_Bindings *BindingsCallerSession) FI() (common.Address, error) {
	return _Bindings.Contract.FI(&_Bindings.CallOpts)
}

// NumPayments is a free data retrieval call binding the contract method 0x0858830b.
//
// Solidity: function numPayments( address) constant returns(uint256)
func (_Bindings *BindingsCaller) NumPayments(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Bindings.contract.Call(opts, out, "numPayments", arg0)
	return *ret0, err
}

// NumPayments is a free data retrieval call binding the contract method 0x0858830b.
//
// Solidity: function numPayments( address) constant returns(uint256)
func (_Bindings *BindingsSession) NumPayments(arg0 common.Address) (*big.Int, error) {
	return _Bindings.Contract.NumPayments(&_Bindings.CallOpts, arg0)
}

// NumPayments is a free data retrieval call binding the contract method 0x0858830b.
//
// Solidity: function numPayments( address) constant returns(uint256)
func (_Bindings *BindingsCallerSession) NumPayments(arg0 common.Address) (*big.Int, error) {
	return _Bindings.Contract.NumPayments(&_Bindings.CallOpts, arg0)
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

// Payments is a free data retrieval call binding the contract method 0x0716326d.
//
// Solidity: function payments( bytes32) constant returns(uploader address, paymentID bytes32, hashedCID bytes32, retentionPeriodInMonths uint256, paymentAmount uint256, state uint8, method uint8)
func (_Bindings *BindingsCaller) Payments(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Uploader                common.Address
	PaymentID               [32]byte
	HashedCID               [32]byte
	RetentionPeriodInMonths *big.Int
	PaymentAmount           *big.Int
	State                   uint8
	Method                  uint8
}, error) {
	ret := new(struct {
		Uploader                common.Address
		PaymentID               [32]byte
		HashedCID               [32]byte
		RetentionPeriodInMonths *big.Int
		PaymentAmount           *big.Int
		State                   uint8
		Method                  uint8
	})
	out := ret
	err := _Bindings.contract.Call(opts, out, "payments", arg0)
	return *ret, err
}

// Payments is a free data retrieval call binding the contract method 0x0716326d.
//
// Solidity: function payments( bytes32) constant returns(uploader address, paymentID bytes32, hashedCID bytes32, retentionPeriodInMonths uint256, paymentAmount uint256, state uint8, method uint8)
func (_Bindings *BindingsSession) Payments(arg0 [32]byte) (struct {
	Uploader                common.Address
	PaymentID               [32]byte
	HashedCID               [32]byte
	RetentionPeriodInMonths *big.Int
	PaymentAmount           *big.Int
	State                   uint8
	Method                  uint8
}, error) {
	return _Bindings.Contract.Payments(&_Bindings.CallOpts, arg0)
}

// Payments is a free data retrieval call binding the contract method 0x0716326d.
//
// Solidity: function payments( bytes32) constant returns(uploader address, paymentID bytes32, hashedCID bytes32, retentionPeriodInMonths uint256, paymentAmount uint256, state uint8, method uint8)
func (_Bindings *BindingsCallerSession) Payments(arg0 [32]byte) (struct {
	Uploader                common.Address
	PaymentID               [32]byte
	HashedCID               [32]byte
	RetentionPeriodInMonths *big.Int
	PaymentAmount           *big.Int
	State                   uint8
	Method                  uint8
}, error) {
	return _Bindings.Contract.Payments(&_Bindings.CallOpts, arg0)
}

// UI is a free data retrieval call binding the contract method 0xebbec31e.
//
// Solidity: function uI() constant returns(address)
func (_Bindings *BindingsCaller) UI(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Bindings.contract.Call(opts, out, "uI")
	return *ret0, err
}

// UI is a free data retrieval call binding the contract method 0xebbec31e.
//
// Solidity: function uI() constant returns(address)
func (_Bindings *BindingsSession) UI() (common.Address, error) {
	return _Bindings.Contract.UI(&_Bindings.CallOpts)
}

// UI is a free data retrieval call binding the contract method 0xebbec31e.
//
// Solidity: function uI() constant returns(address)
func (_Bindings *BindingsCallerSession) UI() (common.Address, error) {
	return _Bindings.Contract.UI(&_Bindings.CallOpts)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Bindings *BindingsTransactor) ChangeAdmin(opts *bind.TransactOpts, _newAdmin common.Address) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "changeAdmin", _newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Bindings *BindingsSession) ChangeAdmin(_newAdmin common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.ChangeAdmin(&_Bindings.TransactOpts, _newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Bindings *BindingsTransactorSession) ChangeAdmin(_newAdmin common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.ChangeAdmin(&_Bindings.TransactOpts, _newAdmin)
}

// PayEthForPaymentID is a paid mutator transaction binding the contract method 0x9818acb9.
//
// Solidity: function payEthForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Bindings *BindingsTransactor) PayEthForPaymentID(opts *bind.TransactOpts, _paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "payEthForPaymentID", _paymentID, _amount)
}

// PayEthForPaymentID is a paid mutator transaction binding the contract method 0x9818acb9.
//
// Solidity: function payEthForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Bindings *BindingsSession) PayEthForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.PayEthForPaymentID(&_Bindings.TransactOpts, _paymentID, _amount)
}

// PayEthForPaymentID is a paid mutator transaction binding the contract method 0x9818acb9.
//
// Solidity: function payEthForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Bindings *BindingsTransactorSession) PayEthForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.PayEthForPaymentID(&_Bindings.TransactOpts, _paymentID, _amount)
}

// PayRtcForPaymentID is a paid mutator transaction binding the contract method 0x57a5d426.
//
// Solidity: function payRtcForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Bindings *BindingsTransactor) PayRtcForPaymentID(opts *bind.TransactOpts, _paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "payRtcForPaymentID", _paymentID, _amount)
}

// PayRtcForPaymentID is a paid mutator transaction binding the contract method 0x57a5d426.
//
// Solidity: function payRtcForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Bindings *BindingsSession) PayRtcForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.PayRtcForPaymentID(&_Bindings.TransactOpts, _paymentID, _amount)
}

// PayRtcForPaymentID is a paid mutator transaction binding the contract method 0x57a5d426.
//
// Solidity: function payRtcForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Bindings *BindingsTransactorSession) PayRtcForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.PayRtcForPaymentID(&_Bindings.TransactOpts, _paymentID, _amount)
}

// RegisterPayment is a paid mutator transaction binding the contract method 0x8ebed58f.
//
// Solidity: function registerPayment(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _method uint8) returns(bool)
func (_Bindings *BindingsTransactor) RegisterPayment(opts *bind.TransactOpts, _uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int, _amount *big.Int, _method uint8) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "registerPayment", _uploader, _hashedCID, _retentionPeriodInMonths, _amount, _method)
}

// RegisterPayment is a paid mutator transaction binding the contract method 0x8ebed58f.
//
// Solidity: function registerPayment(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _method uint8) returns(bool)
func (_Bindings *BindingsSession) RegisterPayment(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int, _amount *big.Int, _method uint8) (*types.Transaction, error) {
	return _Bindings.Contract.RegisterPayment(&_Bindings.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths, _amount, _method)
}

// RegisterPayment is a paid mutator transaction binding the contract method 0x8ebed58f.
//
// Solidity: function registerPayment(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _method uint8) returns(bool)
func (_Bindings *BindingsTransactorSession) RegisterPayment(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int, _amount *big.Int, _method uint8) (*types.Transaction, error) {
	return _Bindings.Contract.RegisterPayment(&_Bindings.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths, _amount, _method)
}

// SetFilesInterface is a paid mutator transaction binding the contract method 0x1308bb66.
//
// Solidity: function setFilesInterface(_filesContractAddress address) returns(bool)
func (_Bindings *BindingsTransactor) SetFilesInterface(opts *bind.TransactOpts, _filesContractAddress common.Address) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setFilesInterface", _filesContractAddress)
}

// SetFilesInterface is a paid mutator transaction binding the contract method 0x1308bb66.
//
// Solidity: function setFilesInterface(_filesContractAddress address) returns(bool)
func (_Bindings *BindingsSession) SetFilesInterface(_filesContractAddress common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetFilesInterface(&_Bindings.TransactOpts, _filesContractAddress)
}

// SetFilesInterface is a paid mutator transaction binding the contract method 0x1308bb66.
//
// Solidity: function setFilesInterface(_filesContractAddress address) returns(bool)
func (_Bindings *BindingsTransactorSession) SetFilesInterface(_filesContractAddress common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetFilesInterface(&_Bindings.TransactOpts, _filesContractAddress)
}

// SetUsersInterface is a paid mutator transaction binding the contract method 0x5b8e47af.
//
// Solidity: function setUsersInterface(_usersContractAddress address) returns(bool)
func (_Bindings *BindingsTransactor) SetUsersInterface(opts *bind.TransactOpts, _usersContractAddress common.Address) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setUsersInterface", _usersContractAddress)
}

// SetUsersInterface is a paid mutator transaction binding the contract method 0x5b8e47af.
//
// Solidity: function setUsersInterface(_usersContractAddress address) returns(bool)
func (_Bindings *BindingsSession) SetUsersInterface(_usersContractAddress common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetUsersInterface(&_Bindings.TransactOpts, _usersContractAddress)
}

// SetUsersInterface is a paid mutator transaction binding the contract method 0x5b8e47af.
//
// Solidity: function setUsersInterface(_usersContractAddress address) returns(bool)
func (_Bindings *BindingsTransactorSession) SetUsersInterface(_usersContractAddress common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetUsersInterface(&_Bindings.TransactOpts, _usersContractAddress)
}

// BindingsEthPaymentReceivedIterator is returned from FilterEthPaymentReceived and is used to iterate over the raw logs and unpacked data for EthPaymentReceived events raised by the Bindings contract.
type BindingsEthPaymentReceivedIterator struct {
	Event *BindingsEthPaymentReceived // Event containing the contract specifics and raw log

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
func (it *BindingsEthPaymentReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsEthPaymentReceived)
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
		it.Event = new(BindingsEthPaymentReceived)
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
func (it *BindingsEthPaymentReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsEthPaymentReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsEthPaymentReceived represents a EthPaymentReceived event raised by the Bindings contract.
type BindingsEthPaymentReceived struct {
	Uploader  common.Address
	PaymentID [32]byte
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEthPaymentReceived is a free log retrieval operation binding the contract event 0x9d04fb9b0795e110b6f4b206c3f8c1e6767ac5cf7fd560647ccf3ffd6c9dd5ee.
//
// Solidity: e EthPaymentReceived(_uploader indexed address, _paymentID bytes32, _amount uint256)
func (_Bindings *BindingsFilterer) FilterEthPaymentReceived(opts *bind.FilterOpts, _uploader []common.Address) (*BindingsEthPaymentReceivedIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "EthPaymentReceived", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &BindingsEthPaymentReceivedIterator{contract: _Bindings.contract, event: "EthPaymentReceived", logs: logs, sub: sub}, nil
}

// WatchEthPaymentReceived is a free log subscription operation binding the contract event 0x9d04fb9b0795e110b6f4b206c3f8c1e6767ac5cf7fd560647ccf3ffd6c9dd5ee.
//
// Solidity: e EthPaymentReceived(_uploader indexed address, _paymentID bytes32, _amount uint256)
func (_Bindings *BindingsFilterer) WatchEthPaymentReceived(opts *bind.WatchOpts, sink chan<- *BindingsEthPaymentReceived, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "EthPaymentReceived", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsEthPaymentReceived)
				if err := _Bindings.contract.UnpackLog(event, "EthPaymentReceived", log); err != nil {
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

// BindingsFilesContractSetIterator is returned from FilterFilesContractSet and is used to iterate over the raw logs and unpacked data for FilesContractSet events raised by the Bindings contract.
type BindingsFilesContractSetIterator struct {
	Event *BindingsFilesContractSet // Event containing the contract specifics and raw log

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
func (it *BindingsFilesContractSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsFilesContractSet)
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
		it.Event = new(BindingsFilesContractSet)
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
func (it *BindingsFilesContractSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsFilesContractSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsFilesContractSet represents a FilesContractSet event raised by the Bindings contract.
type BindingsFilesContractSet struct {
	FilesContractAddress common.Address
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterFilesContractSet is a free log retrieval operation binding the contract event 0x57df6050063bfc7245fb45847eab30542686438bc930cf2f1d0947158615071c.
//
// Solidity: e FilesContractSet(_filesContractAddress address)
func (_Bindings *BindingsFilterer) FilterFilesContractSet(opts *bind.FilterOpts) (*BindingsFilesContractSetIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "FilesContractSet")
	if err != nil {
		return nil, err
	}
	return &BindingsFilesContractSetIterator{contract: _Bindings.contract, event: "FilesContractSet", logs: logs, sub: sub}, nil
}

// WatchFilesContractSet is a free log subscription operation binding the contract event 0x57df6050063bfc7245fb45847eab30542686438bc930cf2f1d0947158615071c.
//
// Solidity: e FilesContractSet(_filesContractAddress address)
func (_Bindings *BindingsFilterer) WatchFilesContractSet(opts *bind.WatchOpts, sink chan<- *BindingsFilesContractSet) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "FilesContractSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsFilesContractSet)
				if err := _Bindings.contract.UnpackLog(event, "FilesContractSet", log); err != nil {
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

// BindingsPaymentRegisteredIterator is returned from FilterPaymentRegistered and is used to iterate over the raw logs and unpacked data for PaymentRegistered events raised by the Bindings contract.
type BindingsPaymentRegisteredIterator struct {
	Event *BindingsPaymentRegistered // Event containing the contract specifics and raw log

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
func (it *BindingsPaymentRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsPaymentRegistered)
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
		it.Event = new(BindingsPaymentRegistered)
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
func (it *BindingsPaymentRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsPaymentRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsPaymentRegistered represents a PaymentRegistered event raised by the Bindings contract.
type BindingsPaymentRegistered struct {
	Uploader                common.Address
	HashedCID               [32]byte
	RetentionPeriodInMonths *big.Int
	Amount                  *big.Int
	PaymentID               [32]byte
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterPaymentRegistered is a free log retrieval operation binding the contract event 0xcb6bac54d8308a9609a62a77e0389ef2b7d019f81686bfb4e556643f70110f7a.
//
// Solidity: e PaymentRegistered(_uploader indexed address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _paymentID bytes32)
func (_Bindings *BindingsFilterer) FilterPaymentRegistered(opts *bind.FilterOpts, _uploader []common.Address) (*BindingsPaymentRegisteredIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "PaymentRegistered", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &BindingsPaymentRegisteredIterator{contract: _Bindings.contract, event: "PaymentRegistered", logs: logs, sub: sub}, nil
}

// WatchPaymentRegistered is a free log subscription operation binding the contract event 0xcb6bac54d8308a9609a62a77e0389ef2b7d019f81686bfb4e556643f70110f7a.
//
// Solidity: e PaymentRegistered(_uploader indexed address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _paymentID bytes32)
func (_Bindings *BindingsFilterer) WatchPaymentRegistered(opts *bind.WatchOpts, sink chan<- *BindingsPaymentRegistered, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "PaymentRegistered", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsPaymentRegistered)
				if err := _Bindings.contract.UnpackLog(event, "PaymentRegistered", log); err != nil {
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

// BindingsRtcPaymentReceivedIterator is returned from FilterRtcPaymentReceived and is used to iterate over the raw logs and unpacked data for RtcPaymentReceived events raised by the Bindings contract.
type BindingsRtcPaymentReceivedIterator struct {
	Event *BindingsRtcPaymentReceived // Event containing the contract specifics and raw log

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
func (it *BindingsRtcPaymentReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsRtcPaymentReceived)
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
		it.Event = new(BindingsRtcPaymentReceived)
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
func (it *BindingsRtcPaymentReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsRtcPaymentReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsRtcPaymentReceived represents a RtcPaymentReceived event raised by the Bindings contract.
type BindingsRtcPaymentReceived struct {
	Uploader  common.Address
	PaymentID [32]byte
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRtcPaymentReceived is a free log retrieval operation binding the contract event 0x71536f9dd7c4e8db4b8cb8226889aaea1c562bca1da233f3e0e7846f4e65d57b.
//
// Solidity: e RtcPaymentReceived(_uploader indexed address, _paymentID bytes32, _amount uint256)
func (_Bindings *BindingsFilterer) FilterRtcPaymentReceived(opts *bind.FilterOpts, _uploader []common.Address) (*BindingsRtcPaymentReceivedIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "RtcPaymentReceived", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &BindingsRtcPaymentReceivedIterator{contract: _Bindings.contract, event: "RtcPaymentReceived", logs: logs, sub: sub}, nil
}

// WatchRtcPaymentReceived is a free log subscription operation binding the contract event 0x71536f9dd7c4e8db4b8cb8226889aaea1c562bca1da233f3e0e7846f4e65d57b.
//
// Solidity: e RtcPaymentReceived(_uploader indexed address, _paymentID bytes32, _amount uint256)
func (_Bindings *BindingsFilterer) WatchRtcPaymentReceived(opts *bind.WatchOpts, sink chan<- *BindingsRtcPaymentReceived, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "RtcPaymentReceived", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsRtcPaymentReceived)
				if err := _Bindings.contract.UnpackLog(event, "RtcPaymentReceived", log); err != nil {
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

// BindingsUsersContractSetIterator is returned from FilterUsersContractSet and is used to iterate over the raw logs and unpacked data for UsersContractSet events raised by the Bindings contract.
type BindingsUsersContractSetIterator struct {
	Event *BindingsUsersContractSet // Event containing the contract specifics and raw log

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
func (it *BindingsUsersContractSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsUsersContractSet)
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
		it.Event = new(BindingsUsersContractSet)
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
func (it *BindingsUsersContractSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsUsersContractSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsUsersContractSet represents a UsersContractSet event raised by the Bindings contract.
type BindingsUsersContractSet struct {
	UsersContractAddress common.Address
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterUsersContractSet is a free log retrieval operation binding the contract event 0x0c7303206058ab0e0d85e1f17933330be8aba69aa937a13044b6c10e40886e20.
//
// Solidity: e UsersContractSet(_usersContractAddress address)
func (_Bindings *BindingsFilterer) FilterUsersContractSet(opts *bind.FilterOpts) (*BindingsUsersContractSetIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "UsersContractSet")
	if err != nil {
		return nil, err
	}
	return &BindingsUsersContractSetIterator{contract: _Bindings.contract, event: "UsersContractSet", logs: logs, sub: sub}, nil
}

// WatchUsersContractSet is a free log subscription operation binding the contract event 0x0c7303206058ab0e0d85e1f17933330be8aba69aa937a13044b6c10e40886e20.
//
// Solidity: e UsersContractSet(_usersContractAddress address)
func (_Bindings *BindingsFilterer) WatchUsersContractSet(opts *bind.WatchOpts, sink chan<- *BindingsUsersContractSet) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "UsersContractSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsUsersContractSet)
				if err := _Bindings.contract.UnpackLog(event, "UsersContractSet", log); err != nil {
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
