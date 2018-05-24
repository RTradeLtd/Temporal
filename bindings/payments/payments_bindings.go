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
const PaymentsABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"payments\",\"outputs\":[{\"name\":\"uploader\",\"type\":\"address\"},{\"name\":\"paymentID\",\"type\":\"bytes32\"},{\"name\":\"hashedCID\",\"type\":\"bytes32\"},{\"name\":\"retentionPeriodInMonths\",\"type\":\"uint256\"},{\"name\":\"paymentAmount\",\"type\":\"uint256\"},{\"name\":\"state\",\"type\":\"uint8\"},{\"name\":\"method\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"numPayments\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_filesContractAddress\",\"type\":\"address\"}],\"name\":\"setFilesInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"payRtcForPaymentID\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_usersContractAddress\",\"type\":\"address\"}],\"name\":\"setUsersInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_uploader\",\"type\":\"address\"},{\"name\":\"_hashedCID\",\"type\":\"bytes32\"},{\"name\":\"_retentionPeriodInMonths\",\"type\":\"uint256\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_method\",\"type\":\"uint8\"}],\"name\":\"registerPayment\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newAdmin\",\"type\":\"address\"}],\"name\":\"changeAdmin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"payEthForPaymentID\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"paymentIDs\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"fI\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"uI\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_filesContractAddress\",\"type\":\"address\"}],\"name\":\"FilesContractSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_usersContractAddress\",\"type\":\"address\"}],\"name\":\"UsersContractSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_hashedCID\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_retentionPeriodInMonths\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_paymentID\",\"type\":\"bytes32\"}],\"name\":\"PaymentRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_method\",\"type\":\"uint8\"}],\"name\":\"PaymentReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_paymentID\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_method\",\"type\":\"uint8\"}],\"name\":\"PaymentReceivedNoIndex\",\"type\":\"event\"}]"

// PaymentsBin is the compiled bytecode used for deploying new contracts.
const PaymentsBin = `60c0604052601c60808190527f19457468657265756d205369676e6564204d6573736167653a0a33320000000060a090815261003e9160029190610066565b506000805433600160a060020a03199182168117909255600180549091169091179055610101565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100a757805160ff19168380011785556100d4565b828001600101855582156100d4579182015b828111156100d45782518255916020019190600101906100b9565b506100e09291506100e4565b5090565b6100fe91905b808211156100e057600081556001016100ea565b90565b610dc0806101106000396000f3006080604052600436106100c45763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630716326d81146100c95780630858830b1461014a5780631308bb661461017d57806357a5d426146101b25780635b8e47af146101cd5780638da5cb5b146101ee5780638ebed58f1461021f5780638f2839701461024f5780639818acb9146102705780639c8d6ee21461028b578063db8ce76c146102af578063ebbec31e146102c4578063f851a440146102d9575b600080fd5b3480156100d557600080fd5b506100e16004356102ee565b60408051600160a060020a038916815260208101889052908101869052606081018590526080810184905260a0810183600281111561011c57fe5b60ff16815260200182600181111561013057fe5b60ff16815260200197505050505050505060405180910390f35b34801561015657600080fd5b5061016b600160a060020a0360043516610339565b60408051918252519081900360200190f35b34801561018957600080fd5b5061019e600160a060020a036004351661034b565b604080519115158252519081900360200190f35b3480156101be57600080fd5b5061019e6004356024356103fd565b3480156101d957600080fd5b5061019e600160a060020a03600435166106db565b3480156101fa57600080fd5b50610203610776565b60408051600160a060020a039092168252519081900360200190f35b34801561022b57600080fd5b5061019e600160a060020a036004351660243560443560643560ff60843516610785565b34801561025b57600080fd5b5061019e600160a060020a0360043516610a4a565b34801561027c57600080fd5b5061019e600435602435610a99565b34801561029757600080fd5b5061016b600160a060020a0360043516602435610d31565b3480156102bb57600080fd5b50610203610d4e565b3480156102d057600080fd5b50610203610d5d565b3480156102e557600080fd5b50610203610d6c565b6005602081905260009182526040909120805460018201546002830154600384015460048501549490950154600160a060020a03909316949193909260ff8082169161010090041687565b60066020526000908152604090205481565b600080543390600160a060020a03168114806103745750600154600160a060020a038281169116145b151561037f57600080fd5b82600160a060020a038116151561039557600080fd5b60048054600160a060020a03861673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f57df6050063bfc7245fb45847eab30542686438bc930cf2f1d0947158615071c9181900360200190a15060019392505050565b60008260016000828152600560208190526040909120015460ff16600281111561042357fe5b1461042d57600080fd5b60008481526005602052604090205484903390600160a060020a0316811461045457600080fd5b60008681526005602052604090206004015486908690811461047557600080fd5b8760008060008381526005602081905260409091200154610100900460ff16600181111561049f57fe5b146104a957600080fd5b60008a8152600560208181526040808420909201805460ff1916600217905581518d81529081018c9052808201929092525133917ffa3ee50f898224e13a1d5371d4718d72c50302f72b86f3cff70b1173f3cfe158919081900360600190a260408051338152602081018c90528082018b90526000606082015290517fd31f82a1ce72a95fc27f987e544f4ceaf77e260b791acde8f6c2bf38f346b82b9181900360800190a16004805460008c8152600560209081526040808320600281015460039091015482517f6eb033f400000000000000000000000000000000000000000000000000000000815233978101979097526024870191909152604486015251600160a060020a0390931693636eb033f49360648083019491928390030190829087803b1580156105da57600080fd5b505af11580156105ee573d6000803e3d6000fd5b505050506040513d602081101561060457600080fd5b5051151561061157600080fd5b60035460008b81526005602090815260408083206002015481517f66a04e91000000000000000000000000000000000000000000000000000000008152336004820152602481018f905260448101919091529051600160a060020a03909416936366a04e9193606480840194938390030190829087803b15801561069457600080fd5b505af11580156106a8573d6000803e3d6000fd5b505050506040513d60208110156106be57600080fd5b505115156106cb57600080fd5b5060019998505050505050505050565b600080543390600160a060020a03168114806107045750600154600160a060020a038281169116145b151561070f57600080fd5b60038054600160a060020a03851673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f0c7303206058ab0e0d85e1f17933330be8aba69aa937a13044b6c10e40886e209181900360200190a150600192915050565b600054600160a060020a031681565b6000805481903390600160a060020a03168114806107b05750600154600160a060020a038281169116145b15156107bb57600080fd5b84600081116107c957600080fd5b86600081116107d757600080fd5b89600160a060020a03811615156107ed57600080fd5b60018760ff161180610802575060008760ff16105b1561080c57600080fd5b600160a060020a038b166000818152600660205260408082205481516c010000000000000000000000009094028452601484018e9052603484018d90526054840152426074840152519182900360940190912095506000868152600560208190526040909120015460ff16600281111561088257fe5b1461088c57600080fd5b6040805160e081018252600160a060020a038d168152602081018790529081018b9052606081018a90526080810189905260a08101600181526020018860ff1660018111156108d757fe5b60018111156108e257fe5b90526000868152600560208181526040928390208451815473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a039091161781559084015160018083019190915592840151600280830191909155606085015160038301556080850151600483015560a085015192820180549294909260ff191691849081111561096b57fe5b021790555060c082015160058201805461ff00191661010083600181111561098f57fe5b02179055505050600160a060020a038b166000908152600660205260409020546109c090600163ffffffff610d7b16565b600160a060020a038c166000908152600660209081526040808320849055600782528083209383529281529082902087905581518c81529081018b90528082018a905260608101879052905133917fcb6bac54d8308a9609a62a77e0389ef2b7d019f81686bfb4e556643f70110f7a919081900360800190a25060019a9950505050505050505050565b600080548290600160a060020a03808316911614610a6757600080fd5b60018054600160a060020a03851673ffffffffffffffffffffffffffffffffffffffff19909116178155915050919050565b60008260016000828152600560208190526040909120015460ff166002811115610abf57fe5b14610ac957600080fd5b60008481526005602052604090205484903390600160a060020a03168114610af057600080fd5b600086815260056020526040902060040154869086908114610b1157600080fd5b8760018060008381526005602081905260409091200154610100900460ff166001811115610b3b57fe5b14610b4557600080fd5b60008a815260056020818152604092839020909101805460ff1916600217905581518c81529081018b9052600181830152905133917ffa3ee50f898224e13a1d5371d4718d72c50302f72b86f3cff70b1173f3cfe158919081900360600190a260408051338152602081018c90528082018b90526001606082015290517fd31f82a1ce72a95fc27f987e544f4ceaf77e260b791acde8f6c2bf38f346b82b9181900360800190a16004805460008c8152600560209081526040808320600281015460039091015482517f6eb033f400000000000000000000000000000000000000000000000000000000815233978101979097526024870191909152604486015251600160a060020a0390931693636eb033f49360648083019491928390030190829087803b158015610c7757600080fd5b505af1158015610c8b573d6000803e3d6000fd5b505050506040513d6020811015610ca157600080fd5b50511515610cae57600080fd5b60035460008b81526005602090815260408083206002015481517f70b9c01e000000000000000000000000000000000000000000000000000000008152336004820152602481018f905260448101919091529051600160a060020a03909416936370b9c01e93606480840194938390030190829087803b15801561069457600080fd5b600760209081526000928352604080842090915290825290205481565b600454600160a060020a031681565b600354600160a060020a031681565b600154600160a060020a031681565b600082820183811015610d8d57600080fd5b93925050505600a165627a7a723058205b31cd2e1b8678908238379f329b25bb337c5b345060a438c539a967eb1413ae0029`

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

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Payments *PaymentsCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "admin")
	return *ret0, err
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Payments *PaymentsSession) Admin() (common.Address, error) {
	return _Payments.Contract.Admin(&_Payments.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Payments *PaymentsCallerSession) Admin() (common.Address, error) {
	return _Payments.Contract.Admin(&_Payments.CallOpts)
}

// FI is a free data retrieval call binding the contract method 0xdb8ce76c.
//
// Solidity: function fI() constant returns(address)
func (_Payments *PaymentsCaller) FI(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "fI")
	return *ret0, err
}

// FI is a free data retrieval call binding the contract method 0xdb8ce76c.
//
// Solidity: function fI() constant returns(address)
func (_Payments *PaymentsSession) FI() (common.Address, error) {
	return _Payments.Contract.FI(&_Payments.CallOpts)
}

// FI is a free data retrieval call binding the contract method 0xdb8ce76c.
//
// Solidity: function fI() constant returns(address)
func (_Payments *PaymentsCallerSession) FI() (common.Address, error) {
	return _Payments.Contract.FI(&_Payments.CallOpts)
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

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Payments *PaymentsCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Payments *PaymentsSession) Owner() (common.Address, error) {
	return _Payments.Contract.Owner(&_Payments.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Payments *PaymentsCallerSession) Owner() (common.Address, error) {
	return _Payments.Contract.Owner(&_Payments.CallOpts)
}

// PaymentIDs is a free data retrieval call binding the contract method 0x9c8d6ee2.
//
// Solidity: function paymentIDs( address,  uint256) constant returns(bytes32)
func (_Payments *PaymentsCaller) PaymentIDs(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "paymentIDs", arg0, arg1)
	return *ret0, err
}

// PaymentIDs is a free data retrieval call binding the contract method 0x9c8d6ee2.
//
// Solidity: function paymentIDs( address,  uint256) constant returns(bytes32)
func (_Payments *PaymentsSession) PaymentIDs(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _Payments.Contract.PaymentIDs(&_Payments.CallOpts, arg0, arg1)
}

// PaymentIDs is a free data retrieval call binding the contract method 0x9c8d6ee2.
//
// Solidity: function paymentIDs( address,  uint256) constant returns(bytes32)
func (_Payments *PaymentsCallerSession) PaymentIDs(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _Payments.Contract.PaymentIDs(&_Payments.CallOpts, arg0, arg1)
}

// Payments is a free data retrieval call binding the contract method 0x0716326d.
//
// Solidity: function payments( bytes32) constant returns(uploader address, paymentID bytes32, hashedCID bytes32, retentionPeriodInMonths uint256, paymentAmount uint256, state uint8, method uint8)
func (_Payments *PaymentsCaller) Payments(opts *bind.CallOpts, arg0 [32]byte) (struct {
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
	err := _Payments.contract.Call(opts, out, "payments", arg0)
	return *ret, err
}

// Payments is a free data retrieval call binding the contract method 0x0716326d.
//
// Solidity: function payments( bytes32) constant returns(uploader address, paymentID bytes32, hashedCID bytes32, retentionPeriodInMonths uint256, paymentAmount uint256, state uint8, method uint8)
func (_Payments *PaymentsSession) Payments(arg0 [32]byte) (struct {
	Uploader                common.Address
	PaymentID               [32]byte
	HashedCID               [32]byte
	RetentionPeriodInMonths *big.Int
	PaymentAmount           *big.Int
	State                   uint8
	Method                  uint8
}, error) {
	return _Payments.Contract.Payments(&_Payments.CallOpts, arg0)
}

// Payments is a free data retrieval call binding the contract method 0x0716326d.
//
// Solidity: function payments( bytes32) constant returns(uploader address, paymentID bytes32, hashedCID bytes32, retentionPeriodInMonths uint256, paymentAmount uint256, state uint8, method uint8)
func (_Payments *PaymentsCallerSession) Payments(arg0 [32]byte) (struct {
	Uploader                common.Address
	PaymentID               [32]byte
	HashedCID               [32]byte
	RetentionPeriodInMonths *big.Int
	PaymentAmount           *big.Int
	State                   uint8
	Method                  uint8
}, error) {
	return _Payments.Contract.Payments(&_Payments.CallOpts, arg0)
}

// UI is a free data retrieval call binding the contract method 0xebbec31e.
//
// Solidity: function uI() constant returns(address)
func (_Payments *PaymentsCaller) UI(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Payments.contract.Call(opts, out, "uI")
	return *ret0, err
}

// UI is a free data retrieval call binding the contract method 0xebbec31e.
//
// Solidity: function uI() constant returns(address)
func (_Payments *PaymentsSession) UI() (common.Address, error) {
	return _Payments.Contract.UI(&_Payments.CallOpts)
}

// UI is a free data retrieval call binding the contract method 0xebbec31e.
//
// Solidity: function uI() constant returns(address)
func (_Payments *PaymentsCallerSession) UI() (common.Address, error) {
	return _Payments.Contract.UI(&_Payments.CallOpts)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Payments *PaymentsTransactor) ChangeAdmin(opts *bind.TransactOpts, _newAdmin common.Address) (*types.Transaction, error) {
	return _Payments.contract.Transact(opts, "changeAdmin", _newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Payments *PaymentsSession) ChangeAdmin(_newAdmin common.Address) (*types.Transaction, error) {
	return _Payments.Contract.ChangeAdmin(&_Payments.TransactOpts, _newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Payments *PaymentsTransactorSession) ChangeAdmin(_newAdmin common.Address) (*types.Transaction, error) {
	return _Payments.Contract.ChangeAdmin(&_Payments.TransactOpts, _newAdmin)
}

// PayEthForPaymentID is a paid mutator transaction binding the contract method 0x9818acb9.
//
// Solidity: function payEthForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Payments *PaymentsTransactor) PayEthForPaymentID(opts *bind.TransactOpts, _paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Payments.contract.Transact(opts, "payEthForPaymentID", _paymentID, _amount)
}

// PayEthForPaymentID is a paid mutator transaction binding the contract method 0x9818acb9.
//
// Solidity: function payEthForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Payments *PaymentsSession) PayEthForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Payments.Contract.PayEthForPaymentID(&_Payments.TransactOpts, _paymentID, _amount)
}

// PayEthForPaymentID is a paid mutator transaction binding the contract method 0x9818acb9.
//
// Solidity: function payEthForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Payments *PaymentsTransactorSession) PayEthForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Payments.Contract.PayEthForPaymentID(&_Payments.TransactOpts, _paymentID, _amount)
}

// PayRtcForPaymentID is a paid mutator transaction binding the contract method 0x57a5d426.
//
// Solidity: function payRtcForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Payments *PaymentsTransactor) PayRtcForPaymentID(opts *bind.TransactOpts, _paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Payments.contract.Transact(opts, "payRtcForPaymentID", _paymentID, _amount)
}

// PayRtcForPaymentID is a paid mutator transaction binding the contract method 0x57a5d426.
//
// Solidity: function payRtcForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Payments *PaymentsSession) PayRtcForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Payments.Contract.PayRtcForPaymentID(&_Payments.TransactOpts, _paymentID, _amount)
}

// PayRtcForPaymentID is a paid mutator transaction binding the contract method 0x57a5d426.
//
// Solidity: function payRtcForPaymentID(_paymentID bytes32, _amount uint256) returns(bool)
func (_Payments *PaymentsTransactorSession) PayRtcForPaymentID(_paymentID [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Payments.Contract.PayRtcForPaymentID(&_Payments.TransactOpts, _paymentID, _amount)
}

// RegisterPayment is a paid mutator transaction binding the contract method 0x8ebed58f.
//
// Solidity: function registerPayment(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _method uint8) returns(bool)
func (_Payments *PaymentsTransactor) RegisterPayment(opts *bind.TransactOpts, _uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int, _amount *big.Int, _method uint8) (*types.Transaction, error) {
	return _Payments.contract.Transact(opts, "registerPayment", _uploader, _hashedCID, _retentionPeriodInMonths, _amount, _method)
}

// RegisterPayment is a paid mutator transaction binding the contract method 0x8ebed58f.
//
// Solidity: function registerPayment(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _method uint8) returns(bool)
func (_Payments *PaymentsSession) RegisterPayment(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int, _amount *big.Int, _method uint8) (*types.Transaction, error) {
	return _Payments.Contract.RegisterPayment(&_Payments.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths, _amount, _method)
}

// RegisterPayment is a paid mutator transaction binding the contract method 0x8ebed58f.
//
// Solidity: function registerPayment(_uploader address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _method uint8) returns(bool)
func (_Payments *PaymentsTransactorSession) RegisterPayment(_uploader common.Address, _hashedCID [32]byte, _retentionPeriodInMonths *big.Int, _amount *big.Int, _method uint8) (*types.Transaction, error) {
	return _Payments.Contract.RegisterPayment(&_Payments.TransactOpts, _uploader, _hashedCID, _retentionPeriodInMonths, _amount, _method)
}

// SetFilesInterface is a paid mutator transaction binding the contract method 0x1308bb66.
//
// Solidity: function setFilesInterface(_filesContractAddress address) returns(bool)
func (_Payments *PaymentsTransactor) SetFilesInterface(opts *bind.TransactOpts, _filesContractAddress common.Address) (*types.Transaction, error) {
	return _Payments.contract.Transact(opts, "setFilesInterface", _filesContractAddress)
}

// SetFilesInterface is a paid mutator transaction binding the contract method 0x1308bb66.
//
// Solidity: function setFilesInterface(_filesContractAddress address) returns(bool)
func (_Payments *PaymentsSession) SetFilesInterface(_filesContractAddress common.Address) (*types.Transaction, error) {
	return _Payments.Contract.SetFilesInterface(&_Payments.TransactOpts, _filesContractAddress)
}

// SetFilesInterface is a paid mutator transaction binding the contract method 0x1308bb66.
//
// Solidity: function setFilesInterface(_filesContractAddress address) returns(bool)
func (_Payments *PaymentsTransactorSession) SetFilesInterface(_filesContractAddress common.Address) (*types.Transaction, error) {
	return _Payments.Contract.SetFilesInterface(&_Payments.TransactOpts, _filesContractAddress)
}

// SetUsersInterface is a paid mutator transaction binding the contract method 0x5b8e47af.
//
// Solidity: function setUsersInterface(_usersContractAddress address) returns(bool)
func (_Payments *PaymentsTransactor) SetUsersInterface(opts *bind.TransactOpts, _usersContractAddress common.Address) (*types.Transaction, error) {
	return _Payments.contract.Transact(opts, "setUsersInterface", _usersContractAddress)
}

// SetUsersInterface is a paid mutator transaction binding the contract method 0x5b8e47af.
//
// Solidity: function setUsersInterface(_usersContractAddress address) returns(bool)
func (_Payments *PaymentsSession) SetUsersInterface(_usersContractAddress common.Address) (*types.Transaction, error) {
	return _Payments.Contract.SetUsersInterface(&_Payments.TransactOpts, _usersContractAddress)
}

// SetUsersInterface is a paid mutator transaction binding the contract method 0x5b8e47af.
//
// Solidity: function setUsersInterface(_usersContractAddress address) returns(bool)
func (_Payments *PaymentsTransactorSession) SetUsersInterface(_usersContractAddress common.Address) (*types.Transaction, error) {
	return _Payments.Contract.SetUsersInterface(&_Payments.TransactOpts, _usersContractAddress)
}

// PaymentsFilesContractSetIterator is returned from FilterFilesContractSet and is used to iterate over the raw logs and unpacked data for FilesContractSet events raised by the Payments contract.
type PaymentsFilesContractSetIterator struct {
	Event *PaymentsFilesContractSet // Event containing the contract specifics and raw log

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
func (it *PaymentsFilesContractSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentsFilesContractSet)
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
		it.Event = new(PaymentsFilesContractSet)
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
func (it *PaymentsFilesContractSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentsFilesContractSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentsFilesContractSet represents a FilesContractSet event raised by the Payments contract.
type PaymentsFilesContractSet struct {
	FilesContractAddress common.Address
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterFilesContractSet is a free log retrieval operation binding the contract event 0x57df6050063bfc7245fb45847eab30542686438bc930cf2f1d0947158615071c.
//
// Solidity: e FilesContractSet(_filesContractAddress address)
func (_Payments *PaymentsFilterer) FilterFilesContractSet(opts *bind.FilterOpts) (*PaymentsFilesContractSetIterator, error) {

	logs, sub, err := _Payments.contract.FilterLogs(opts, "FilesContractSet")
	if err != nil {
		return nil, err
	}
	return &PaymentsFilesContractSetIterator{contract: _Payments.contract, event: "FilesContractSet", logs: logs, sub: sub}, nil
}

// WatchFilesContractSet is a free log subscription operation binding the contract event 0x57df6050063bfc7245fb45847eab30542686438bc930cf2f1d0947158615071c.
//
// Solidity: e FilesContractSet(_filesContractAddress address)
func (_Payments *PaymentsFilterer) WatchFilesContractSet(opts *bind.WatchOpts, sink chan<- *PaymentsFilesContractSet) (event.Subscription, error) {

	logs, sub, err := _Payments.contract.WatchLogs(opts, "FilesContractSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentsFilesContractSet)
				if err := _Payments.contract.UnpackLog(event, "FilesContractSet", log); err != nil {
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

// PaymentsPaymentReceivedIterator is returned from FilterPaymentReceived and is used to iterate over the raw logs and unpacked data for PaymentReceived events raised by the Payments contract.
type PaymentsPaymentReceivedIterator struct {
	Event *PaymentsPaymentReceived // Event containing the contract specifics and raw log

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
func (it *PaymentsPaymentReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentsPaymentReceived)
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
		it.Event = new(PaymentsPaymentReceived)
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
func (it *PaymentsPaymentReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentsPaymentReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentsPaymentReceived represents a PaymentReceived event raised by the Payments contract.
type PaymentsPaymentReceived struct {
	Uploader  common.Address
	PaymentID [32]byte
	Amount    *big.Int
	Method    uint8
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPaymentReceived is a free log retrieval operation binding the contract event 0xfa3ee50f898224e13a1d5371d4718d72c50302f72b86f3cff70b1173f3cfe158.
//
// Solidity: e PaymentReceived(_uploader indexed address, _paymentID bytes32, _amount uint256, _method uint8)
func (_Payments *PaymentsFilterer) FilterPaymentReceived(opts *bind.FilterOpts, _uploader []common.Address) (*PaymentsPaymentReceivedIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Payments.contract.FilterLogs(opts, "PaymentReceived", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &PaymentsPaymentReceivedIterator{contract: _Payments.contract, event: "PaymentReceived", logs: logs, sub: sub}, nil
}

// WatchPaymentReceived is a free log subscription operation binding the contract event 0xfa3ee50f898224e13a1d5371d4718d72c50302f72b86f3cff70b1173f3cfe158.
//
// Solidity: e PaymentReceived(_uploader indexed address, _paymentID bytes32, _amount uint256, _method uint8)
func (_Payments *PaymentsFilterer) WatchPaymentReceived(opts *bind.WatchOpts, sink chan<- *PaymentsPaymentReceived, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Payments.contract.WatchLogs(opts, "PaymentReceived", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentsPaymentReceived)
				if err := _Payments.contract.UnpackLog(event, "PaymentReceived", log); err != nil {
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

// PaymentsPaymentReceivedNoIndexIterator is returned from FilterPaymentReceivedNoIndex and is used to iterate over the raw logs and unpacked data for PaymentReceivedNoIndex events raised by the Payments contract.
type PaymentsPaymentReceivedNoIndexIterator struct {
	Event *PaymentsPaymentReceivedNoIndex // Event containing the contract specifics and raw log

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
func (it *PaymentsPaymentReceivedNoIndexIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentsPaymentReceivedNoIndex)
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
		it.Event = new(PaymentsPaymentReceivedNoIndex)
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
func (it *PaymentsPaymentReceivedNoIndexIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentsPaymentReceivedNoIndexIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentsPaymentReceivedNoIndex represents a PaymentReceivedNoIndex event raised by the Payments contract.
type PaymentsPaymentReceivedNoIndex struct {
	Uploader  common.Address
	PaymentID [32]byte
	Amount    *big.Int
	Method    uint8
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPaymentReceivedNoIndex is a free log retrieval operation binding the contract event 0xd31f82a1ce72a95fc27f987e544f4ceaf77e260b791acde8f6c2bf38f346b82b.
//
// Solidity: e PaymentReceivedNoIndex(_uploader address, _paymentID bytes32, _amount uint256, _method uint8)
func (_Payments *PaymentsFilterer) FilterPaymentReceivedNoIndex(opts *bind.FilterOpts) (*PaymentsPaymentReceivedNoIndexIterator, error) {

	logs, sub, err := _Payments.contract.FilterLogs(opts, "PaymentReceivedNoIndex")
	if err != nil {
		return nil, err
	}
	return &PaymentsPaymentReceivedNoIndexIterator{contract: _Payments.contract, event: "PaymentReceivedNoIndex", logs: logs, sub: sub}, nil
}

// WatchPaymentReceivedNoIndex is a free log subscription operation binding the contract event 0xd31f82a1ce72a95fc27f987e544f4ceaf77e260b791acde8f6c2bf38f346b82b.
//
// Solidity: e PaymentReceivedNoIndex(_uploader address, _paymentID bytes32, _amount uint256, _method uint8)
func (_Payments *PaymentsFilterer) WatchPaymentReceivedNoIndex(opts *bind.WatchOpts, sink chan<- *PaymentsPaymentReceivedNoIndex) (event.Subscription, error) {

	logs, sub, err := _Payments.contract.WatchLogs(opts, "PaymentReceivedNoIndex")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentsPaymentReceivedNoIndex)
				if err := _Payments.contract.UnpackLog(event, "PaymentReceivedNoIndex", log); err != nil {
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

// PaymentsPaymentRegisteredIterator is returned from FilterPaymentRegistered and is used to iterate over the raw logs and unpacked data for PaymentRegistered events raised by the Payments contract.
type PaymentsPaymentRegisteredIterator struct {
	Event *PaymentsPaymentRegistered // Event containing the contract specifics and raw log

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
func (it *PaymentsPaymentRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentsPaymentRegistered)
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
		it.Event = new(PaymentsPaymentRegistered)
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
func (it *PaymentsPaymentRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentsPaymentRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentsPaymentRegistered represents a PaymentRegistered event raised by the Payments contract.
type PaymentsPaymentRegistered struct {
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
func (_Payments *PaymentsFilterer) FilterPaymentRegistered(opts *bind.FilterOpts, _uploader []common.Address) (*PaymentsPaymentRegisteredIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Payments.contract.FilterLogs(opts, "PaymentRegistered", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &PaymentsPaymentRegisteredIterator{contract: _Payments.contract, event: "PaymentRegistered", logs: logs, sub: sub}, nil
}

// WatchPaymentRegistered is a free log subscription operation binding the contract event 0xcb6bac54d8308a9609a62a77e0389ef2b7d019f81686bfb4e556643f70110f7a.
//
// Solidity: e PaymentRegistered(_uploader indexed address, _hashedCID bytes32, _retentionPeriodInMonths uint256, _amount uint256, _paymentID bytes32)
func (_Payments *PaymentsFilterer) WatchPaymentRegistered(opts *bind.WatchOpts, sink chan<- *PaymentsPaymentRegistered, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Payments.contract.WatchLogs(opts, "PaymentRegistered", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentsPaymentRegistered)
				if err := _Payments.contract.UnpackLog(event, "PaymentRegistered", log); err != nil {
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

// PaymentsUsersContractSetIterator is returned from FilterUsersContractSet and is used to iterate over the raw logs and unpacked data for UsersContractSet events raised by the Payments contract.
type PaymentsUsersContractSetIterator struct {
	Event *PaymentsUsersContractSet // Event containing the contract specifics and raw log

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
func (it *PaymentsUsersContractSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentsUsersContractSet)
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
		it.Event = new(PaymentsUsersContractSet)
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
func (it *PaymentsUsersContractSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentsUsersContractSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentsUsersContractSet represents a UsersContractSet event raised by the Payments contract.
type PaymentsUsersContractSet struct {
	UsersContractAddress common.Address
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterUsersContractSet is a free log retrieval operation binding the contract event 0x0c7303206058ab0e0d85e1f17933330be8aba69aa937a13044b6c10e40886e20.
//
// Solidity: e UsersContractSet(_usersContractAddress address)
func (_Payments *PaymentsFilterer) FilterUsersContractSet(opts *bind.FilterOpts) (*PaymentsUsersContractSetIterator, error) {

	logs, sub, err := _Payments.contract.FilterLogs(opts, "UsersContractSet")
	if err != nil {
		return nil, err
	}
	return &PaymentsUsersContractSetIterator{contract: _Payments.contract, event: "UsersContractSet", logs: logs, sub: sub}, nil
}

// WatchUsersContractSet is a free log subscription operation binding the contract event 0x0c7303206058ab0e0d85e1f17933330be8aba69aa937a13044b6c10e40886e20.
//
// Solidity: e UsersContractSet(_usersContractAddress address)
func (_Payments *PaymentsFilterer) WatchUsersContractSet(opts *bind.WatchOpts, sink chan<- *PaymentsUsersContractSet) (event.Subscription, error) {

	logs, sub, err := _Payments.contract.WatchLogs(opts, "UsersContractSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentsUsersContractSet)
				if err := _Payments.contract.UnpackLog(event, "UsersContractSet", log); err != nil {
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
