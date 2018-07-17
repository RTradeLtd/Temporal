// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package payments

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

// PaymentsABI is the input ABI used to generate the binding from.
const PaymentsABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"numPayments\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"HOTWALLET\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"TOKENADDRESS\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"SIGNER\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"RTI\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"payments\",\"outputs\":[{\"name\":\"paymentNumber\",\"type\":\"uint256\"},{\"name\":\"chargeAmountInWei\",\"type\":\"uint256\"},{\"name\":\"method\",\"type\":\"uint8\"},{\"name\":\"state\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_h\",\"type\":\"bytes32\"},{\"name\":\"_v\",\"type\":\"uint8\"},{\"name\":\"_r\",\"type\":\"bytes32\"},{\"name\":\"_s\",\"type\":\"bytes32\"},{\"name\":\"_paymentNumber\",\"type\":\"uint256\"},{\"name\":\"_paymentMethod\",\"type\":\"uint8\"},{\"name\":\"_chargeAmountInWei\",\"type\":\"uint256\"},{\"name\":\"_prefixed\",\"type\":\"bool\"}],\"name\":\"verifySigner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_h\",\"type\":\"bytes32\"},{\"name\":\"_v\",\"type\":\"uint8\"},{\"name\":\"_r\",\"type\":\"bytes32\"},{\"name\":\"_s\",\"type\":\"bytes32\"},{\"name\":\"_paymentNumber\",\"type\":\"uint256\"},{\"name\":\"_paymentMethod\",\"type\":\"uint8\"},{\"name\":\"_chargeAmountInWei\",\"type\":\"uint256\"},{\"name\":\"_prefixed\",\"type\":\"bool\"}],\"name\":\"makePayment\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_h\",\"type\":\"bytes32\"},{\"name\":\"_paymentNumber\",\"type\":\"uint256\"},{\"name\":\"_paymentMethod\",\"type\":\"uint8\"},{\"name\":\"_chargeAmountInWei\",\"type\":\"uint256\"},{\"name\":\"_prefixed\",\"type\":\"bool\"}],\"name\":\"verifyImages\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_payer\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_paymentNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_paymentMethod\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"_paymentAmount\",\"type\":\"uint256\"}],\"name\":\"PaymentMade\",\"type\":\"event\"}]"

// PaymentsBin is the compiled bytecode used for deploying new contracts.
const PaymentsBin = `608060405234801561001057600080fd5b50610d8a806100206000396000f3006080604052600436106100985763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630858830b811461009d578063342e569f146100dd578063516f89861461011b578063582abd12146100dd57806358e1c1741461011b578063ab63385c14610130578063b8df17f0146101a8578063e4e0c030146101f4578063eb87073a1461021f575b600080fd5b3480156100a957600080fd5b506100cb73ffffffffffffffffffffffffffffffffffffffff60043516610248565b60408051918252519081900360200190f35b3480156100e957600080fd5b506100f261025a565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b34801561012757600080fd5b506100f2610272565b34801561013c57600080fd5b5061016173ffffffffffffffffffffffffffffffffffffffff6004351660243561028a565b6040518085815260200184815260200183600181111561017d57fe5b60ff16815260200182600181111561019157fe5b60ff16815260200194505050505060405180910390f35b3480156101b457600080fd5b506101e060043560ff60243581169060443590606435906084359060a4351660c43560e43515156102be565b604080519115158252519081900360200190f35b6101e060043560ff60243581169060443590606435906084359060a4351660c43560e4351515610453565b34801561022b57600080fd5b506101e060043560243560ff604435166064356084351515610aad565b60006020819052908152604090205481565b737e4a2359c745a982a54653128085eac69e446de181565b73185ae6a87bbb02097923e859d742747bb979ae9a81565b6001602081815260009384526040808520909152918352912080549181015460029091015460ff8082169161010090041684565b6000808060ff861615806102d557508560ff166001145b151561032b576040805160e560020a62461bcd02815260206004820152601660248201527f696e76616c6964207061796d656e74206d6574686f6400000000000000000000604482015290519081900360640190fd5b831561034e5761033c878688610b59565b905061034781610c26565b915061035c565b610359878688610b59565b91505b818b146103b3576040805160e560020a62461bcd02815260206004820152601f60248201527f6661696c656420746f207265636f6e73747275637420707265696d6167657300604482015290519081900360640190fd5b60408051600080825260208083018085528f905260ff8e1683850152606083018d9052608083018c90529251737e4a2359c745a982a54653128085eac69e446de19360019360a0808201949293601f198101939281900390910191865af1158015610422573d6000803e3d6000fd5b5050506020604051035173ffffffffffffffffffffffffffffffffffffffff16149250505098975050505050505050565b600080600080610461610d37565b88600033600090815260016020818152604080842086855290915290912060020154610100900460ff169081111561049557fe5b146104ea576040805160e560020a62461bcd02815260206004820152601460248201527f7061796d656e7420616c7265616479206d616465000000000000000000000000604482015290519081900360640190fd5b60ff891615806104fd57508860ff166001145b1515610553576040805160e560020a62461bcd02815260206004820152601660248201527f696e76616c6964207061796d656e74206d6574686f6400000000000000000000604482015290519081900360640190fd5b8615610576576105648a898b610b59565b935061056f84610c26565b9450610584565b6105818a898b610b59565b94505b848e14610601576040805160e560020a62461bcd02815260206004820152602560248201527f7265636f6e737472756374656420707265696d61676520646f6573206e6f742060448201527f6d61746368000000000000000000000000000000000000000000000000000000606482015290519081900360840190fd5b60018e8e8e8e604051600081526020016040526040518085600019166000191681526020018460ff1660ff1681526020018360001916600019168152602001826000191660001916815260200194505050505060206040516020810390808403906000865af1158015610678573d6000803e3d6000fd5b5050604051601f19015193505073ffffffffffffffffffffffffffffffffffffffff8316737e4a2359c745a982a54653128085eac69e446de114610706576040805160e560020a62461bcd02815260206004820152601f60248201527f7265636f7665726564207369676e657220646f6573206e6f74206d6174636800604482015290519081900360640190fd5b6080604051908101604052808b81526020018981526020018a60ff16600181111561072d57fe5b600181111561073857fe5b8152602001600190523360009081526001602081815260408084208f85528252928390208451815590840151818301559183015160028301805494965086949192909160ff191690838181111561078b57fe5b0217905550606082015160028201805461ff0019166101008360018111156107af57fe5b021790555050336000908152602081905260409020546107d79150600163ffffffff610d1e16565b3360009081526020819052604090205560018960ff1660018111156107f857fe5b600181111561080357fe5b141561091957348814610886576040805160e560020a62461bcd02815260206004820152602660248201527f6d73672e76616c756520646f6573206e6f7420657175616c206368617267652060448201527f616d6f756e740000000000000000000000000000000000000000000000000000606482015290519081900360840190fd5b60408051338152602081018c905260ff8b1681830152606081018a905290517fd18793644b4cb4ec0f937f8153dc09112d762775084c0ae5b4c21a7b91f6909f9181900360800190a1604051737e4a2359c745a982a54653128085eac69e446de1903480156108fc02916000818181858888f1935050505015801561090f573d6000803e3d6000fd5b5060019550610a9c565b60408051338152602081018c905260ff8b1681830152606081018a905290517fd18793644b4cb4ec0f937f8153dc09112d762775084c0ae5b4c21a7b91f6909f9181900360800190a1604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152737e4a2359c745a982a54653128085eac69e446de16024820152604481018a9052905173185ae6a87bbb02097923e859d742747bb979ae9a916323b872dd9160648083019260209291908290030181600087803b1580156109ef57600080fd5b505af1158015610a03573d6000803e3d6000fd5b505050506040513d6020811015610a1957600080fd5b50511515610a97576040805160e560020a62461bcd02815260206004820152602e60248201527f7472617366657246726f6d206661696c65642c206d6f7374206c696b656c792060448201527f6e6565647320617070726f76616c000000000000000000000000000000000000606482015290519081900360840190fd5b600195505b505050505098975050505050505050565b6000808060ff86161580610ac457508560ff166001145b1515610b1a576040805160e560020a62461bcd02815260206004820152601660248201527f696e76616c6964207061796d656e74206d6574686f6400000000000000000000604482015290519081900360640190fd5b8315610b3d57610b2b878688610b59565b9050610b3681610c26565b9150610b4b565b610b48878688610b59565b91505b509590951495945050505050565b604080516c010000000000000000000000003302602080830191909152603482018690527f010000000000000000000000000000000000000000000000000000000000000060ff85160260548301526055808301869052835180840390910181526075909201928390528151600093918291908401908083835b60208310610bf25780518252601f199092019160209182019101610bd3565b5181516020939093036101000a60001901801990911692169190911790526040519201829003909120979650505050505050565b604080518082018252601c8082527f19457468657265756d205369676e6564204d6573736167653a0a33320000000060208084019182529351600094869391019182918083835b60208310610c8c5780518252601f199092019160209182019101610c6d565b51815160209384036101000a600019018019909216911617905292019384525060408051808503815293820190819052835193945092839250908401908083835b60208310610cec5780518252601f199092019160209182019101610ccd565b5181516020939093036101000a6000190180199091169216919091179052604051920182900390912095945050505050565b600082820183811015610d3057600080fd5b9392505050565b604080516080810182526000808252602082018190529091820190815260200160009052905600a165627a7a7230582095bbf938f4d25cb635a14df23b072aca74825aee316a883b743d28b0463bad450029`

// DeployPayments deploys a new Ethereum contract, binding an instance of Payments to it.
func DeployPayments(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Payments, error) {
	parsed, err := abi.JSON(strings.NewReader(PaymentsABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PaymentsBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Payments{PaymentsCaller: PaymentsCaller{contract: contract}, PaymentsTransactor: PaymentsTransactor{contract: contract}, PaymentsFilterer: PaymentsFilterer{contract: contract}}, nil
}

// Payments is an auto generated Go binding around an Ethereum contract.
type Payments struct {
	PaymentsCaller     // Read-only binding to the contract
	PaymentsTransactor // Write-only binding to the contract
	PaymentsFilterer   // Log filterer for contract events
}

// PaymentsCaller is an auto generated read-only Go binding around an Ethereum contract.
type PaymentsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PaymentsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PaymentsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PaymentsSession struct {
	Contract     *Payments         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PaymentsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PaymentsCallerSession struct {
	Contract *PaymentsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// PaymentsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PaymentsTransactorSession struct {
	Contract     *PaymentsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// PaymentsRaw is an auto generated low-level Go binding around an Ethereum contract.
type PaymentsRaw struct {
	Contract *Payments // Generic contract binding to access the raw methods on
}

// PaymentsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PaymentsCallerRaw struct {
	Contract *PaymentsCaller // Generic read-only contract binding to access the raw methods on
}

// PaymentsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PaymentsTransactorRaw struct {
	Contract *PaymentsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPayments creates a new instance of Payments, bound to a specific deployed contract.
func NewPayments(address common.Address, backend bind.ContractBackend) (*Payments, error) {
	contract, err := bindPayments(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Payments{PaymentsCaller: PaymentsCaller{contract: contract}, PaymentsTransactor: PaymentsTransactor{contract: contract}, PaymentsFilterer: PaymentsFilterer{contract: contract}}, nil
}

// NewPaymentsCaller creates a new read-only instance of Payments, bound to a specific deployed contract.
func NewPaymentsCaller(address common.Address, caller bind.ContractCaller) (*PaymentsCaller, error) {
	contract, err := bindPayments(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PaymentsCaller{contract: contract}, nil
}

// NewPaymentsTransactor creates a new write-only instance of Payments, bound to a specific deployed contract.
func NewPaymentsTransactor(address common.Address, transactor bind.ContractTransactor) (*PaymentsTransactor, error) {
	contract, err := bindPayments(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PaymentsTransactor{contract: contract}, nil
}

// NewPaymentsFilterer creates a new log filterer instance of Payments, bound to a specific deployed contract.
func NewPaymentsFilterer(address common.Address, filterer bind.ContractFilterer) (*PaymentsFilterer, error) {
	contract, err := bindPayments(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PaymentsFilterer{contract: contract}, nil
}

// bindPayments binds a generic wrapper to an already deployed contract.
func bindPayments(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PaymentsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Payments *PaymentsRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Payments.Contract.PaymentsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Payments *PaymentsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Payments.Contract.PaymentsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Payments *PaymentsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Payments.Contract.PaymentsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Payments *PaymentsCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Payments.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Payments *PaymentsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Payments.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Payments *PaymentsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Payments.Contract.contract.Transact(opts, method, params...)
}

// HOTWALLET is a free data retrieval call binding the contract method 0x342e569f.
//
// Solidity: function HOTWALLET() constant returns(address)
func (_Payments *PaymentsCaller) HOTWALLET(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "HOTWALLET")
	return *ret0, err
}

// HOTWALLET is a free data retrieval call binding the contract method 0x342e569f.
//
// Solidity: function HOTWALLET() constant returns(address)
func (_Payments *PaymentsSession) HOTWALLET() (common.Address, error) {
	return _Payments.Contract.HOTWALLET(&_Payments.CallOpts)
}

// HOTWALLET is a free data retrieval call binding the contract method 0x342e569f.
//
// Solidity: function HOTWALLET() constant returns(address)
func (_Payments *PaymentsCallerSession) HOTWALLET() (common.Address, error) {
	return _Payments.Contract.HOTWALLET(&_Payments.CallOpts)
}

// RTI is a free data retrieval call binding the contract method 0x58e1c174.
//
// Solidity: function RTI() constant returns(address)
func (_Payments *PaymentsCaller) RTI(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "RTI")
	return *ret0, err
}

// RTI is a free data retrieval call binding the contract method 0x58e1c174.
//
// Solidity: function RTI() constant returns(address)
func (_Payments *PaymentsSession) RTI() (common.Address, error) {
	return _Payments.Contract.RTI(&_Payments.CallOpts)
}

// RTI is a free data retrieval call binding the contract method 0x58e1c174.
//
// Solidity: function RTI() constant returns(address)
func (_Payments *PaymentsCallerSession) RTI() (common.Address, error) {
	return _Payments.Contract.RTI(&_Payments.CallOpts)
}

// SIGNER is a free data retrieval call binding the contract method 0x582abd12.
//
// Solidity: function SIGNER() constant returns(address)
func (_Payments *PaymentsCaller) SIGNER(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "SIGNER")
	return *ret0, err
}

// SIGNER is a free data retrieval call binding the contract method 0x582abd12.
//
// Solidity: function SIGNER() constant returns(address)
func (_Payments *PaymentsSession) SIGNER() (common.Address, error) {
	return _Payments.Contract.SIGNER(&_Payments.CallOpts)
}

// SIGNER is a free data retrieval call binding the contract method 0x582abd12.
//
// Solidity: function SIGNER() constant returns(address)
func (_Payments *PaymentsCallerSession) SIGNER() (common.Address, error) {
	return _Payments.Contract.SIGNER(&_Payments.CallOpts)
}

// TOKENADDRESS is a free data retrieval call binding the contract method 0x516f8986.
//
// Solidity: function TOKENADDRESS() constant returns(address)
func (_Payments *PaymentsCaller) TOKENADDRESS(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "TOKENADDRESS")
	return *ret0, err
}

// TOKENADDRESS is a free data retrieval call binding the contract method 0x516f8986.
//
// Solidity: function TOKENADDRESS() constant returns(address)
func (_Payments *PaymentsSession) TOKENADDRESS() (common.Address, error) {
	return _Payments.Contract.TOKENADDRESS(&_Payments.CallOpts)
}

// TOKENADDRESS is a free data retrieval call binding the contract method 0x516f8986.
//
// Solidity: function TOKENADDRESS() constant returns(address)
func (_Payments *PaymentsCallerSession) TOKENADDRESS() (common.Address, error) {
	return _Payments.Contract.TOKENADDRESS(&_Payments.CallOpts)
}

// NumPayments is a free data retrieval call binding the contract method 0x0858830b.
//
// Solidity: function numPayments( address) constant returns(uint256)
func (_Payments *PaymentsCaller) NumPayments(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "numPayments", arg0)
	return *ret0, err
}

// NumPayments is a free data retrieval call binding the contract method 0x0858830b.
//
// Solidity: function numPayments( address) constant returns(uint256)
func (_Payments *PaymentsSession) NumPayments(arg0 common.Address) (*big.Int, error) {
	return _Payments.Contract.NumPayments(&_Payments.CallOpts, arg0)
}

// NumPayments is a free data retrieval call binding the contract method 0x0858830b.
//
// Solidity: function numPayments( address) constant returns(uint256)
func (_Payments *PaymentsCallerSession) NumPayments(arg0 common.Address) (*big.Int, error) {
	return _Payments.Contract.NumPayments(&_Payments.CallOpts, arg0)
}

// Payments is a free data retrieval call binding the contract method 0xab63385c.
//
// Solidity: function payments( address,  uint256) constant returns(paymentNumber uint256, chargeAmountInWei uint256, method uint8, state uint8)
func (_Payments *PaymentsCaller) Payments(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	PaymentNumber     *big.Int
	ChargeAmountInWei *big.Int
	Method            uint8
	State             uint8
}, error) {
	ret := new(struct {
		PaymentNumber     *big.Int
		ChargeAmountInWei *big.Int
		Method            uint8
		State             uint8
	})
	out := ret
	err := _Payments.contract.Call(opts, out, "payments", arg0, arg1)
	return *ret, err
}

// Payments is a free data retrieval call binding the contract method 0xab63385c.
//
// Solidity: function payments( address,  uint256) constant returns(paymentNumber uint256, chargeAmountInWei uint256, method uint8, state uint8)
func (_Payments *PaymentsSession) Payments(arg0 common.Address, arg1 *big.Int) (struct {
	PaymentNumber     *big.Int
	ChargeAmountInWei *big.Int
	Method            uint8
	State             uint8
}, error) {
	return _Payments.Contract.Payments(&_Payments.CallOpts, arg0, arg1)
}

// Payments is a free data retrieval call binding the contract method 0xab63385c.
//
// Solidity: function payments( address,  uint256) constant returns(paymentNumber uint256, chargeAmountInWei uint256, method uint8, state uint8)
func (_Payments *PaymentsCallerSession) Payments(arg0 common.Address, arg1 *big.Int) (struct {
	PaymentNumber     *big.Int
	ChargeAmountInWei *big.Int
	Method            uint8
	State             uint8
}, error) {
	return _Payments.Contract.Payments(&_Payments.CallOpts, arg0, arg1)
}

// VerifyImages is a free data retrieval call binding the contract method 0xeb87073a.
//
// Solidity: function verifyImages(_h bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) constant returns(bool)
func (_Payments *PaymentsCaller) VerifyImages(opts *bind.CallOpts, _h [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "verifyImages", _h, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
	return *ret0, err
}

// VerifyImages is a free data retrieval call binding the contract method 0xeb87073a.
//
// Solidity: function verifyImages(_h bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) constant returns(bool)
func (_Payments *PaymentsSession) VerifyImages(_h [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (bool, error) {
	return _Payments.Contract.VerifyImages(&_Payments.CallOpts, _h, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
}

// VerifyImages is a free data retrieval call binding the contract method 0xeb87073a.
//
// Solidity: function verifyImages(_h bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) constant returns(bool)
func (_Payments *PaymentsCallerSession) VerifyImages(_h [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (bool, error) {
	return _Payments.Contract.VerifyImages(&_Payments.CallOpts, _h, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
}

// VerifySigner is a free data retrieval call binding the contract method 0xb8df17f0.
//
// Solidity: function verifySigner(_h bytes32, _v uint8, _r bytes32, _s bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) constant returns(bool)
func (_Payments *PaymentsCaller) VerifySigner(opts *bind.CallOpts, _h [32]byte, _v uint8, _r [32]byte, _s [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "verifySigner", _h, _v, _r, _s, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
	return *ret0, err
}

// VerifySigner is a free data retrieval call binding the contract method 0xb8df17f0.
//
// Solidity: function verifySigner(_h bytes32, _v uint8, _r bytes32, _s bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) constant returns(bool)
func (_Payments *PaymentsSession) VerifySigner(_h [32]byte, _v uint8, _r [32]byte, _s [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (bool, error) {
	return _Payments.Contract.VerifySigner(&_Payments.CallOpts, _h, _v, _r, _s, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
}

// VerifySigner is a free data retrieval call binding the contract method 0xb8df17f0.
//
// Solidity: function verifySigner(_h bytes32, _v uint8, _r bytes32, _s bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) constant returns(bool)
func (_Payments *PaymentsCallerSession) VerifySigner(_h [32]byte, _v uint8, _r [32]byte, _s [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (bool, error) {
	return _Payments.Contract.VerifySigner(&_Payments.CallOpts, _h, _v, _r, _s, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
}

// MakePayment is a paid mutator transaction binding the contract method 0xe4e0c030.
//
// Solidity: function makePayment(_h bytes32, _v uint8, _r bytes32, _s bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) returns(bool)
func (_Payments *PaymentsTransactor) MakePayment(opts *bind.TransactOpts, _h [32]byte, _v uint8, _r [32]byte, _s [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (*types.Transaction, error) {
	return _Payments.contract.Transact(opts, "makePayment", _h, _v, _r, _s, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
}

// MakePayment is a paid mutator transaction binding the contract method 0xe4e0c030.
//
// Solidity: function makePayment(_h bytes32, _v uint8, _r bytes32, _s bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) returns(bool)
func (_Payments *PaymentsSession) MakePayment(_h [32]byte, _v uint8, _r [32]byte, _s [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (*types.Transaction, error) {
	return _Payments.Contract.MakePayment(&_Payments.TransactOpts, _h, _v, _r, _s, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
}

// MakePayment is a paid mutator transaction binding the contract method 0xe4e0c030.
//
// Solidity: function makePayment(_h bytes32, _v uint8, _r bytes32, _s bytes32, _paymentNumber uint256, _paymentMethod uint8, _chargeAmountInWei uint256, _prefixed bool) returns(bool)
func (_Payments *PaymentsTransactorSession) MakePayment(_h [32]byte, _v uint8, _r [32]byte, _s [32]byte, _paymentNumber *big.Int, _paymentMethod uint8, _chargeAmountInWei *big.Int, _prefixed bool) (*types.Transaction, error) {
	return _Payments.Contract.MakePayment(&_Payments.TransactOpts, _h, _v, _r, _s, _paymentNumber, _paymentMethod, _chargeAmountInWei, _prefixed)
}

// PaymentsPaymentMadeIterator is returned from FilterPaymentMade and is used to iterate over the raw logs and unpacked data for PaymentMade events raised by the Payments contract.
type PaymentsPaymentMadeIterator struct {
	Event *PaymentsPaymentMade // Event containing the contract specifics and raw log

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
func (it *PaymentsPaymentMadeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentsPaymentMade)
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
		it.Event = new(PaymentsPaymentMade)
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
func (it *PaymentsPaymentMadeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentsPaymentMadeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentsPaymentMade represents a PaymentMade event raised by the Payments contract.
type PaymentsPaymentMade struct {
	Payer         common.Address
	PaymentNumber *big.Int
	PaymentMethod uint8
	PaymentAmount *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterPaymentMade is a free log retrieval operation binding the contract event 0xd18793644b4cb4ec0f937f8153dc09112d762775084c0ae5b4c21a7b91f6909f.
//
// Solidity: e PaymentMade(_payer address, _paymentNumber uint256, _paymentMethod uint8, _paymentAmount uint256)
func (_Payments *PaymentsFilterer) FilterPaymentMade(opts *bind.FilterOpts) (*PaymentsPaymentMadeIterator, error) {

	logs, sub, err := _Payments.contract.FilterLogs(opts, "PaymentMade")
	if err != nil {
		return nil, err
	}
	return &PaymentsPaymentMadeIterator{contract: _Payments.contract, event: "PaymentMade", logs: logs, sub: sub}, nil
}

// WatchPaymentMade is a free log subscription operation binding the contract event 0xd18793644b4cb4ec0f937f8153dc09112d762775084c0ae5b4c21a7b91f6909f.
//
// Solidity: e PaymentMade(_payer address, _paymentNumber uint256, _paymentMethod uint8, _paymentAmount uint256)
func (_Payments *PaymentsFilterer) WatchPaymentMade(opts *bind.WatchOpts, sink chan<- *PaymentsPaymentMade) (event.Subscription, error) {

	logs, sub, err := _Payments.contract.WatchLogs(opts, "PaymentMade")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentsPaymentMade)
				if err := _Payments.contract.UnpackLog(event, "PaymentMade", log); err != nil {
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
