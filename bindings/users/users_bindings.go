package users

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

// UsersABI is the input ABI used to generate the binding from.
const UsersABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"hotWallet\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_tokenContractAddress\",\"type\":\"address\"}],\"name\":\"setRTCInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"registerUser\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_uploaderAddress\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_hashedCID\",\"type\":\"bytes32\"}],\"name\":\"paymentProcessorWithdrawRtcForUploader\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_uploaderAddress\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_hashedCID\",\"type\":\"bytes32\"}],\"name\":\"paymentProcessorWithdrawEthForUploader\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"uploaders\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newAdmin\",\"type\":\"address\"}],\"name\":\"changeAdmin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"depositEther\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hotWalletAddress\",\"type\":\"address\"}],\"name\":\"setHotWallet\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"rtcI\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"depositRtc\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"users\",\"outputs\":[{\"name\":\"uploaderAddress\",\"type\":\"address\"},{\"name\":\"availableEthBalance\",\"type\":\"uint256\"},{\"name\":\"availableRtcBalance\",\"type\":\"uint256\"},{\"name\":\"state\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"paymentProcessorAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_paymentProcessorAddress\",\"type\":\"address\"}],\"name\":\"setPaymentProcessorAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_paymentProcessor\",\"type\":\"address\"}],\"name\":\"PaymentProcessorSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_wallet\",\"type\":\"address\"}],\"name\":\"HotWalletSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_tokenContractAddress\",\"type\":\"address\"}],\"name\":\"RTCInterfaceSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_uploader\",\"type\":\"address\"}],\"name\":\"UserRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"RtcDeposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_hashedCID\",\"type\":\"bytes32\"}],\"name\":\"RtcPaymentWithdrawnForUploader\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"EthDeposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_uploader\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_hashedCID\",\"type\":\"bytes32\"}],\"name\":\"EthPaymentWithdrawnForUpload\",\"type\":\"event\"}]"

// UsersBin is the compiled bytecode used for deploying new contracts.
const UsersBin = `608060405260028054600160a060020a03199081169091556000805482163390811790915560018054909216179055610cbe8061003d6000396000f3006080604052600436106100e55763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166329113bc881146100ea578063303d65e81461011b5780634d3820eb1461015057806366a04e911461016557806370b9c01e1461018c5780638b25321a146101b35780638da5cb5b146101cb5780638f283970146101e057806398ea5fca146102015780639fb755d714610209578063a46574141461022a578063a678b8be1461023f578063a87430ba14610257578063b694a2f2146102c3578063f657c167146102d8578063f851a440146102f9575b600080fd5b3480156100f657600080fd5b506100ff61030e565b60408051600160a060020a039092168252519081900360200190f35b34801561012757600080fd5b5061013c600160a060020a036004351661031d565b604080519115158252519081900360200190f35b34801561015c57600080fd5b5061013c6103cf565b34801561017157600080fd5b5061013c600160a060020a03600435166024356044356104be565b34801561019857600080fd5b5061013c600160a060020a036004351660243560443561068e565b3480156101bf57600080fd5b506100ff6004356107f0565b3480156101d757600080fd5b506100ff610818565b3480156101ec57600080fd5b5061013c600160a060020a0360043516610827565b61013c610876565b34801561021557600080fd5b5061013c600160a060020a0360043516610935565b34801561023657600080fd5b506100ff6109e7565b34801561024b57600080fd5b5061013c6004356109f6565b34801561026357600080fd5b50610278600160a060020a0360043516610b60565b6040518085600160a060020a0316600160a060020a031681526020018481526020018381526020018260028111156102ac57fe5b60ff16815260200194505050505060405180910390f35b3480156102cf57600080fd5b506100ff610b94565b3480156102e457600080fd5b5061013c600160a060020a0360043516610ba3565b34801561030557600080fd5b506100ff610c55565b600354600160a060020a031681565b600080543390600160a060020a03168114806103465750600154600160a060020a038281169116145b151561035157600080fd5b82600160a060020a038116151561036757600080fd5b60028054600160a060020a03861673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f08481da9fb8764513d40cbc8a3c7eabb80f438636910a1462b703d49f0d4c6d89181900360200190a15060019392505050565b60003381600160a060020a03821660009081526006602052604090206003015460ff1660028111156103fd57fe5b1461040757600080fd5b33600081815260066020908152604080832060038101805460ff19166001908117909155815473ffffffffffffffffffffffffffffffffffffffff1990811687179092556005805491820181559094527f036b6384b5eca791c62761152d0c79bb0604c104a5fb6f4eb0703f3154bb3db090930180549093168417909255815192835290517f54db7a5cb4735e1aac1f53db512d3390390bb6637bd30ad4bf9fc98667d9b9b99281900390910190a1600191505090565b600080846001600160a060020a03821660009081526006602052604090206003015460ff1660028111156104ee57fe5b146104f857600080fd5b846000811161050657600080fd5b6004543390600160a060020a0316811461051f57600080fd5b600160a060020a0388166000908152600660205260409020600201548890889081111561054b57600080fd5b600160a060020a038a16600090815260066020526040902060020154610577908a63ffffffff610c6416565b600160a060020a038b1660008181526006602090815260409182902060020184905581518d81529081018c9052815193995091927f23b8eb1a184ac06193ae1da4a63994ffcb68da99c46802f3ba69876de66eaa3e9281900390910190a2600254600354604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a039283166004820152602481018d90529051919092169163a9059cbb9160448083019260209291908290030181600087803b15801561064757600080fd5b505af115801561065b573d6000803e3d6000fd5b505050506040513d602081101561067157600080fd5b5051151561067e57600080fd5b5060019998505050505050505050565b600080846001600160a060020a03821660009081526006602052604090206003015460ff1660028111156106be57fe5b146106c857600080fd5b84600081116106d657600080fd5b6004543390600160a060020a031681146106ef57600080fd5b600160a060020a0388166000908152600660205260409020600101548890889081111561071b57600080fd5b600160a060020a038a16600090815260066020526040902060010154610747908a63ffffffff610c6416565b600160a060020a038b1660008181526006602090815260409182902060010184905581518d81529081018c9052815193995091927f31324c88656ae4aa00bcf1d66deb3f51c278ac2c0cda0244de13db75672d06d19281900390910190a2600354604051600160a060020a03909116908a156108fc02908b906000818181858888f193505050501580156107df573d6000803e3d6000fd5b5060019a9950505050505050505050565b60058054829081106107fe57fe5b600091825260209091200154600160a060020a0316905081565b600054600160a060020a031681565b600080548290600160a060020a0380831691161461084457600080fd5b60018054600160a060020a03851673ffffffffffffffffffffffffffffffffffffffff19909116178155915050919050565b6000336001600160a060020a03821660009081526006602052604090206003015460ff1660028111156108a557fe5b146108af57600080fd5b34600081116108bd57600080fd5b336000908152600660205260409020600101546108e0903463ffffffff610c7916565b33600081815260066020908152604091829020600101939093558051348152905191927f66ff7c8f71ccc7c36152a41920d0d3b46ef3034359f76aa1498ed4478c204b5c92918290030190a260019250505090565b600080543390600160a060020a031681148061095e5750600154600160a060020a038281169116145b151561096957600080fd5b82600160a060020a038116151561097f57600080fd5b60038054600160a060020a03861673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f6f67ba524a8b4aeff0f137d0089ce5a866653b727943b409c166ec963873775f9181900360200190a15060019392505050565b600254600160a060020a031681565b6000336001600160a060020a03821660009081526006602052604090206003015460ff166002811115610a2557fe5b14610a2f57600080fd5b8260008111610a3d57600080fd5b33600090815260066020526040902060020154610a60908563ffffffff610c7916565b33600081815260066020908152604091829020600201939093558051878152905191927f141066711d4e165efac93efabd7eaa910cd70210721bb432e2d8c194dc2139ca92918290030190a2600254604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018790529051600160a060020a03909216916323b872dd916064808201926020929091908290030181600087803b158015610b1f57600080fd5b505af1158015610b33573d6000803e3d6000fd5b505050506040513d6020811015610b4957600080fd5b50511515610b5657600080fd5b5060019392505050565b6006602052600090815260409020805460018201546002830154600390930154600160a060020a0390921692909160ff1684565b600454600160a060020a031681565b600080543390600160a060020a0316811480610bcc5750600154600160a060020a038281169116145b1515610bd757600080fd5b82600160a060020a0381161515610bed57600080fd5b60048054600160a060020a03861673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f8985ba152810f26e00d857552aaff0a9cae15a4fe9bbeb9ec0be19e3e1f064db9181900360200190a15060019392505050565b600154600160a060020a031681565b600082821115610c7357600080fd5b50900390565b600082820183811015610c8b57600080fd5b93925050505600a165627a7a72305820ec4b59033eb198582470f6be3e66f312113690563424abeb759e8d1eab190f470029`

// DeployUsers deploys a new Ethereum contract, binding an instance of Users to it.
func DeployUsers(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Users, error) {
	parsed, err := abi.JSON(strings.NewReader(UsersABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(UsersBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Users{UsersCaller: UsersCaller{contract: contract}, UsersTransactor: UsersTransactor{contract: contract}, UsersFilterer: UsersFilterer{contract: contract}}, nil
}

// Users is an auto generated Go binding around an Ethereum contract.
type Users struct {
	UsersCaller     // Read-only binding to the contract
	UsersTransactor // Write-only binding to the contract
	UsersFilterer   // Log filterer for contract events
}

// UsersCaller is an auto generated read-only Go binding around an Ethereum contract.
type UsersCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UsersTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UsersTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UsersFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UsersFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UsersSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UsersSession struct {
	Contract     *Users            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UsersCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UsersCallerSession struct {
	Contract *UsersCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// UsersTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UsersTransactorSession struct {
	Contract     *UsersTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UsersRaw is an auto generated low-level Go binding around an Ethereum contract.
type UsersRaw struct {
	Contract *Users // Generic contract binding to access the raw methods on
}

// UsersCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UsersCallerRaw struct {
	Contract *UsersCaller // Generic read-only contract binding to access the raw methods on
}

// UsersTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UsersTransactorRaw struct {
	Contract *UsersTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUsers creates a new instance of Users, bound to a specific deployed contract.
func NewUsers(address common.Address, backend bind.ContractBackend) (*Users, error) {
	contract, err := bindUsers(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Users{UsersCaller: UsersCaller{contract: contract}, UsersTransactor: UsersTransactor{contract: contract}, UsersFilterer: UsersFilterer{contract: contract}}, nil
}

// NewUsersCaller creates a new read-only instance of Users, bound to a specific deployed contract.
func NewUsersCaller(address common.Address, caller bind.ContractCaller) (*UsersCaller, error) {
	contract, err := bindUsers(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UsersCaller{contract: contract}, nil
}

// NewUsersTransactor creates a new write-only instance of Users, bound to a specific deployed contract.
func NewUsersTransactor(address common.Address, transactor bind.ContractTransactor) (*UsersTransactor, error) {
	contract, err := bindUsers(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UsersTransactor{contract: contract}, nil
}

// NewUsersFilterer creates a new log filterer instance of Users, bound to a specific deployed contract.
func NewUsersFilterer(address common.Address, filterer bind.ContractFilterer) (*UsersFilterer, error) {
	contract, err := bindUsers(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UsersFilterer{contract: contract}, nil
}

// bindUsers binds a generic wrapper to an already deployed contract.
func bindUsers(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(UsersABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Users *UsersRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Users.Contract.UsersCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Users *UsersRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Users.Contract.UsersTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Users *UsersRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Users.Contract.UsersTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Users *UsersCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Users.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Users *UsersTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Users.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Users *UsersTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Users.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Users *UsersCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Users.contract.Call(opts, out, "admin")
	return *ret0, err
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Users *UsersSession) Admin() (common.Address, error) {
	return _Users.Contract.Admin(&_Users.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Users *UsersCallerSession) Admin() (common.Address, error) {
	return _Users.Contract.Admin(&_Users.CallOpts)
}

// HotWallet is a free data retrieval call binding the contract method 0x29113bc8.
//
// Solidity: function hotWallet() constant returns(address)
func (_Users *UsersCaller) HotWallet(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Users.contract.Call(opts, out, "hotWallet")
	return *ret0, err
}

// HotWallet is a free data retrieval call binding the contract method 0x29113bc8.
//
// Solidity: function hotWallet() constant returns(address)
func (_Users *UsersSession) HotWallet() (common.Address, error) {
	return _Users.Contract.HotWallet(&_Users.CallOpts)
}

// HotWallet is a free data retrieval call binding the contract method 0x29113bc8.
//
// Solidity: function hotWallet() constant returns(address)
func (_Users *UsersCallerSession) HotWallet() (common.Address, error) {
	return _Users.Contract.HotWallet(&_Users.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Users *UsersCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Users.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Users *UsersSession) Owner() (common.Address, error) {
	return _Users.Contract.Owner(&_Users.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Users *UsersCallerSession) Owner() (common.Address, error) {
	return _Users.Contract.Owner(&_Users.CallOpts)
}

// PaymentProcessorAddress is a free data retrieval call binding the contract method 0xb694a2f2.
//
// Solidity: function paymentProcessorAddress() constant returns(address)
func (_Users *UsersCaller) PaymentProcessorAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Users.contract.Call(opts, out, "paymentProcessorAddress")
	return *ret0, err
}

// PaymentProcessorAddress is a free data retrieval call binding the contract method 0xb694a2f2.
//
// Solidity: function paymentProcessorAddress() constant returns(address)
func (_Users *UsersSession) PaymentProcessorAddress() (common.Address, error) {
	return _Users.Contract.PaymentProcessorAddress(&_Users.CallOpts)
}

// PaymentProcessorAddress is a free data retrieval call binding the contract method 0xb694a2f2.
//
// Solidity: function paymentProcessorAddress() constant returns(address)
func (_Users *UsersCallerSession) PaymentProcessorAddress() (common.Address, error) {
	return _Users.Contract.PaymentProcessorAddress(&_Users.CallOpts)
}

// RtcI is a free data retrieval call binding the contract method 0xa4657414.
//
// Solidity: function rtcI() constant returns(address)
func (_Users *UsersCaller) RtcI(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Users.contract.Call(opts, out, "rtcI")
	return *ret0, err
}

// RtcI is a free data retrieval call binding the contract method 0xa4657414.
//
// Solidity: function rtcI() constant returns(address)
func (_Users *UsersSession) RtcI() (common.Address, error) {
	return _Users.Contract.RtcI(&_Users.CallOpts)
}

// RtcI is a free data retrieval call binding the contract method 0xa4657414.
//
// Solidity: function rtcI() constant returns(address)
func (_Users *UsersCallerSession) RtcI() (common.Address, error) {
	return _Users.Contract.RtcI(&_Users.CallOpts)
}

// Uploaders is a free data retrieval call binding the contract method 0x8b25321a.
//
// Solidity: function uploaders( uint256) constant returns(address)
func (_Users *UsersCaller) Uploaders(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Users.contract.Call(opts, out, "uploaders", arg0)
	return *ret0, err
}

// Uploaders is a free data retrieval call binding the contract method 0x8b25321a.
//
// Solidity: function uploaders( uint256) constant returns(address)
func (_Users *UsersSession) Uploaders(arg0 *big.Int) (common.Address, error) {
	return _Users.Contract.Uploaders(&_Users.CallOpts, arg0)
}

// Uploaders is a free data retrieval call binding the contract method 0x8b25321a.
//
// Solidity: function uploaders( uint256) constant returns(address)
func (_Users *UsersCallerSession) Uploaders(arg0 *big.Int) (common.Address, error) {
	return _Users.Contract.Uploaders(&_Users.CallOpts, arg0)
}

// Users is a free data retrieval call binding the contract method 0xa87430ba.
//
// Solidity: function users( address) constant returns(uploaderAddress address, availableEthBalance uint256, availableRtcBalance uint256, state uint8)
func (_Users *UsersCaller) Users(opts *bind.CallOpts, arg0 common.Address) (struct {
	UploaderAddress     common.Address
	AvailableEthBalance *big.Int
	AvailableRtcBalance *big.Int
	State               uint8
}, error) {
	ret := new(struct {
		UploaderAddress     common.Address
		AvailableEthBalance *big.Int
		AvailableRtcBalance *big.Int
		State               uint8
	})
	out := ret
	err := _Users.contract.Call(opts, out, "users", arg0)
	return *ret, err
}

// Users is a free data retrieval call binding the contract method 0xa87430ba.
//
// Solidity: function users( address) constant returns(uploaderAddress address, availableEthBalance uint256, availableRtcBalance uint256, state uint8)
func (_Users *UsersSession) Users(arg0 common.Address) (struct {
	UploaderAddress     common.Address
	AvailableEthBalance *big.Int
	AvailableRtcBalance *big.Int
	State               uint8
}, error) {
	return _Users.Contract.Users(&_Users.CallOpts, arg0)
}

// Users is a free data retrieval call binding the contract method 0xa87430ba.
//
// Solidity: function users( address) constant returns(uploaderAddress address, availableEthBalance uint256, availableRtcBalance uint256, state uint8)
func (_Users *UsersCallerSession) Users(arg0 common.Address) (struct {
	UploaderAddress     common.Address
	AvailableEthBalance *big.Int
	AvailableRtcBalance *big.Int
	State               uint8
}, error) {
	return _Users.Contract.Users(&_Users.CallOpts, arg0)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Users *UsersTransactor) ChangeAdmin(opts *bind.TransactOpts, _newAdmin common.Address) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "changeAdmin", _newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Users *UsersSession) ChangeAdmin(_newAdmin common.Address) (*types.Transaction, error) {
	return _Users.Contract.ChangeAdmin(&_Users.TransactOpts, _newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(_newAdmin address) returns(bool)
func (_Users *UsersTransactorSession) ChangeAdmin(_newAdmin common.Address) (*types.Transaction, error) {
	return _Users.Contract.ChangeAdmin(&_Users.TransactOpts, _newAdmin)
}

// DepositEther is a paid mutator transaction binding the contract method 0x98ea5fca.
//
// Solidity: function depositEther() returns(bool)
func (_Users *UsersTransactor) DepositEther(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "depositEther")
}

// DepositEther is a paid mutator transaction binding the contract method 0x98ea5fca.
//
// Solidity: function depositEther() returns(bool)
func (_Users *UsersSession) DepositEther() (*types.Transaction, error) {
	return _Users.Contract.DepositEther(&_Users.TransactOpts)
}

// DepositEther is a paid mutator transaction binding the contract method 0x98ea5fca.
//
// Solidity: function depositEther() returns(bool)
func (_Users *UsersTransactorSession) DepositEther() (*types.Transaction, error) {
	return _Users.Contract.DepositEther(&_Users.TransactOpts)
}

// DepositRtc is a paid mutator transaction binding the contract method 0xa678b8be.
//
// Solidity: function depositRtc(_amount uint256) returns(bool)
func (_Users *UsersTransactor) DepositRtc(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "depositRtc", _amount)
}

// DepositRtc is a paid mutator transaction binding the contract method 0xa678b8be.
//
// Solidity: function depositRtc(_amount uint256) returns(bool)
func (_Users *UsersSession) DepositRtc(_amount *big.Int) (*types.Transaction, error) {
	return _Users.Contract.DepositRtc(&_Users.TransactOpts, _amount)
}

// DepositRtc is a paid mutator transaction binding the contract method 0xa678b8be.
//
// Solidity: function depositRtc(_amount uint256) returns(bool)
func (_Users *UsersTransactorSession) DepositRtc(_amount *big.Int) (*types.Transaction, error) {
	return _Users.Contract.DepositRtc(&_Users.TransactOpts, _amount)
}

// PaymentProcessorWithdrawEthForUploader is a paid mutator transaction binding the contract method 0x70b9c01e.
//
// Solidity: function paymentProcessorWithdrawEthForUploader(_uploaderAddress address, _amount uint256, _hashedCID bytes32) returns(bool)
func (_Users *UsersTransactor) PaymentProcessorWithdrawEthForUploader(opts *bind.TransactOpts, _uploaderAddress common.Address, _amount *big.Int, _hashedCID [32]byte) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "paymentProcessorWithdrawEthForUploader", _uploaderAddress, _amount, _hashedCID)
}

// PaymentProcessorWithdrawEthForUploader is a paid mutator transaction binding the contract method 0x70b9c01e.
//
// Solidity: function paymentProcessorWithdrawEthForUploader(_uploaderAddress address, _amount uint256, _hashedCID bytes32) returns(bool)
func (_Users *UsersSession) PaymentProcessorWithdrawEthForUploader(_uploaderAddress common.Address, _amount *big.Int, _hashedCID [32]byte) (*types.Transaction, error) {
	return _Users.Contract.PaymentProcessorWithdrawEthForUploader(&_Users.TransactOpts, _uploaderAddress, _amount, _hashedCID)
}

// PaymentProcessorWithdrawEthForUploader is a paid mutator transaction binding the contract method 0x70b9c01e.
//
// Solidity: function paymentProcessorWithdrawEthForUploader(_uploaderAddress address, _amount uint256, _hashedCID bytes32) returns(bool)
func (_Users *UsersTransactorSession) PaymentProcessorWithdrawEthForUploader(_uploaderAddress common.Address, _amount *big.Int, _hashedCID [32]byte) (*types.Transaction, error) {
	return _Users.Contract.PaymentProcessorWithdrawEthForUploader(&_Users.TransactOpts, _uploaderAddress, _amount, _hashedCID)
}

// PaymentProcessorWithdrawRtcForUploader is a paid mutator transaction binding the contract method 0x66a04e91.
//
// Solidity: function paymentProcessorWithdrawRtcForUploader(_uploaderAddress address, _amount uint256, _hashedCID bytes32) returns(bool)
func (_Users *UsersTransactor) PaymentProcessorWithdrawRtcForUploader(opts *bind.TransactOpts, _uploaderAddress common.Address, _amount *big.Int, _hashedCID [32]byte) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "paymentProcessorWithdrawRtcForUploader", _uploaderAddress, _amount, _hashedCID)
}

// PaymentProcessorWithdrawRtcForUploader is a paid mutator transaction binding the contract method 0x66a04e91.
//
// Solidity: function paymentProcessorWithdrawRtcForUploader(_uploaderAddress address, _amount uint256, _hashedCID bytes32) returns(bool)
func (_Users *UsersSession) PaymentProcessorWithdrawRtcForUploader(_uploaderAddress common.Address, _amount *big.Int, _hashedCID [32]byte) (*types.Transaction, error) {
	return _Users.Contract.PaymentProcessorWithdrawRtcForUploader(&_Users.TransactOpts, _uploaderAddress, _amount, _hashedCID)
}

// PaymentProcessorWithdrawRtcForUploader is a paid mutator transaction binding the contract method 0x66a04e91.
//
// Solidity: function paymentProcessorWithdrawRtcForUploader(_uploaderAddress address, _amount uint256, _hashedCID bytes32) returns(bool)
func (_Users *UsersTransactorSession) PaymentProcessorWithdrawRtcForUploader(_uploaderAddress common.Address, _amount *big.Int, _hashedCID [32]byte) (*types.Transaction, error) {
	return _Users.Contract.PaymentProcessorWithdrawRtcForUploader(&_Users.TransactOpts, _uploaderAddress, _amount, _hashedCID)
}

// RegisterUser is a paid mutator transaction binding the contract method 0x4d3820eb.
//
// Solidity: function registerUser() returns(bool)
func (_Users *UsersTransactor) RegisterUser(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "registerUser")
}

// RegisterUser is a paid mutator transaction binding the contract method 0x4d3820eb.
//
// Solidity: function registerUser() returns(bool)
func (_Users *UsersSession) RegisterUser() (*types.Transaction, error) {
	return _Users.Contract.RegisterUser(&_Users.TransactOpts)
}

// RegisterUser is a paid mutator transaction binding the contract method 0x4d3820eb.
//
// Solidity: function registerUser() returns(bool)
func (_Users *UsersTransactorSession) RegisterUser() (*types.Transaction, error) {
	return _Users.Contract.RegisterUser(&_Users.TransactOpts)
}

// SetHotWallet is a paid mutator transaction binding the contract method 0x9fb755d7.
//
// Solidity: function setHotWallet(_hotWalletAddress address) returns(bool)
func (_Users *UsersTransactor) SetHotWallet(opts *bind.TransactOpts, _hotWalletAddress common.Address) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "setHotWallet", _hotWalletAddress)
}

// SetHotWallet is a paid mutator transaction binding the contract method 0x9fb755d7.
//
// Solidity: function setHotWallet(_hotWalletAddress address) returns(bool)
func (_Users *UsersSession) SetHotWallet(_hotWalletAddress common.Address) (*types.Transaction, error) {
	return _Users.Contract.SetHotWallet(&_Users.TransactOpts, _hotWalletAddress)
}

// SetHotWallet is a paid mutator transaction binding the contract method 0x9fb755d7.
//
// Solidity: function setHotWallet(_hotWalletAddress address) returns(bool)
func (_Users *UsersTransactorSession) SetHotWallet(_hotWalletAddress common.Address) (*types.Transaction, error) {
	return _Users.Contract.SetHotWallet(&_Users.TransactOpts, _hotWalletAddress)
}

// SetPaymentProcessorAddress is a paid mutator transaction binding the contract method 0xf657c167.
//
// Solidity: function setPaymentProcessorAddress(_paymentProcessorAddress address) returns(bool)
func (_Users *UsersTransactor) SetPaymentProcessorAddress(opts *bind.TransactOpts, _paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "setPaymentProcessorAddress", _paymentProcessorAddress)
}

// SetPaymentProcessorAddress is a paid mutator transaction binding the contract method 0xf657c167.
//
// Solidity: function setPaymentProcessorAddress(_paymentProcessorAddress address) returns(bool)
func (_Users *UsersSession) SetPaymentProcessorAddress(_paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Users.Contract.SetPaymentProcessorAddress(&_Users.TransactOpts, _paymentProcessorAddress)
}

// SetPaymentProcessorAddress is a paid mutator transaction binding the contract method 0xf657c167.
//
// Solidity: function setPaymentProcessorAddress(_paymentProcessorAddress address) returns(bool)
func (_Users *UsersTransactorSession) SetPaymentProcessorAddress(_paymentProcessorAddress common.Address) (*types.Transaction, error) {
	return _Users.Contract.SetPaymentProcessorAddress(&_Users.TransactOpts, _paymentProcessorAddress)
}

// SetRTCInterface is a paid mutator transaction binding the contract method 0x303d65e8.
//
// Solidity: function setRTCInterface(_tokenContractAddress address) returns(bool)
func (_Users *UsersTransactor) SetRTCInterface(opts *bind.TransactOpts, _tokenContractAddress common.Address) (*types.Transaction, error) {
	return _Users.contract.Transact(opts, "setRTCInterface", _tokenContractAddress)
}

// SetRTCInterface is a paid mutator transaction binding the contract method 0x303d65e8.
//
// Solidity: function setRTCInterface(_tokenContractAddress address) returns(bool)
func (_Users *UsersSession) SetRTCInterface(_tokenContractAddress common.Address) (*types.Transaction, error) {
	return _Users.Contract.SetRTCInterface(&_Users.TransactOpts, _tokenContractAddress)
}

// SetRTCInterface is a paid mutator transaction binding the contract method 0x303d65e8.
//
// Solidity: function setRTCInterface(_tokenContractAddress address) returns(bool)
func (_Users *UsersTransactorSession) SetRTCInterface(_tokenContractAddress common.Address) (*types.Transaction, error) {
	return _Users.Contract.SetRTCInterface(&_Users.TransactOpts, _tokenContractAddress)
}

// UsersEthDepositedIterator is returned from FilterEthDeposited and is used to iterate over the raw logs and unpacked data for EthDeposited events raised by the Users contract.
type UsersEthDepositedIterator struct {
	Event *UsersEthDeposited // Event containing the contract specifics and raw log

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
func (it *UsersEthDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersEthDeposited)
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
		it.Event = new(UsersEthDeposited)
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
func (it *UsersEthDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersEthDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersEthDeposited represents a EthDeposited event raised by the Users contract.
type UsersEthDeposited struct {
	Uploader common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterEthDeposited is a free log retrieval operation binding the contract event 0x66ff7c8f71ccc7c36152a41920d0d3b46ef3034359f76aa1498ed4478c204b5c.
//
// Solidity: e EthDeposited(_uploader indexed address, _amount uint256)
func (_Users *UsersFilterer) FilterEthDeposited(opts *bind.FilterOpts, _uploader []common.Address) (*UsersEthDepositedIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.FilterLogs(opts, "EthDeposited", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &UsersEthDepositedIterator{contract: _Users.contract, event: "EthDeposited", logs: logs, sub: sub}, nil
}

// WatchEthDeposited is a free log subscription operation binding the contract event 0x66ff7c8f71ccc7c36152a41920d0d3b46ef3034359f76aa1498ed4478c204b5c.
//
// Solidity: e EthDeposited(_uploader indexed address, _amount uint256)
func (_Users *UsersFilterer) WatchEthDeposited(opts *bind.WatchOpts, sink chan<- *UsersEthDeposited, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.WatchLogs(opts, "EthDeposited", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersEthDeposited)
				if err := _Users.contract.UnpackLog(event, "EthDeposited", log); err != nil {
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

// UsersEthPaymentWithdrawnForUploadIterator is returned from FilterEthPaymentWithdrawnForUpload and is used to iterate over the raw logs and unpacked data for EthPaymentWithdrawnForUpload events raised by the Users contract.
type UsersEthPaymentWithdrawnForUploadIterator struct {
	Event *UsersEthPaymentWithdrawnForUpload // Event containing the contract specifics and raw log

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
func (it *UsersEthPaymentWithdrawnForUploadIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersEthPaymentWithdrawnForUpload)
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
		it.Event = new(UsersEthPaymentWithdrawnForUpload)
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
func (it *UsersEthPaymentWithdrawnForUploadIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersEthPaymentWithdrawnForUploadIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersEthPaymentWithdrawnForUpload represents a EthPaymentWithdrawnForUpload event raised by the Users contract.
type UsersEthPaymentWithdrawnForUpload struct {
	Uploader  common.Address
	Amount    *big.Int
	HashedCID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEthPaymentWithdrawnForUpload is a free log retrieval operation binding the contract event 0x31324c88656ae4aa00bcf1d66deb3f51c278ac2c0cda0244de13db75672d06d1.
//
// Solidity: e EthPaymentWithdrawnForUpload(_uploader indexed address, _amount uint256, _hashedCID bytes32)
func (_Users *UsersFilterer) FilterEthPaymentWithdrawnForUpload(opts *bind.FilterOpts, _uploader []common.Address) (*UsersEthPaymentWithdrawnForUploadIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.FilterLogs(opts, "EthPaymentWithdrawnForUpload", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &UsersEthPaymentWithdrawnForUploadIterator{contract: _Users.contract, event: "EthPaymentWithdrawnForUpload", logs: logs, sub: sub}, nil
}

// WatchEthPaymentWithdrawnForUpload is a free log subscription operation binding the contract event 0x31324c88656ae4aa00bcf1d66deb3f51c278ac2c0cda0244de13db75672d06d1.
//
// Solidity: e EthPaymentWithdrawnForUpload(_uploader indexed address, _amount uint256, _hashedCID bytes32)
func (_Users *UsersFilterer) WatchEthPaymentWithdrawnForUpload(opts *bind.WatchOpts, sink chan<- *UsersEthPaymentWithdrawnForUpload, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.WatchLogs(opts, "EthPaymentWithdrawnForUpload", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersEthPaymentWithdrawnForUpload)
				if err := _Users.contract.UnpackLog(event, "EthPaymentWithdrawnForUpload", log); err != nil {
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

// UsersHotWalletSetIterator is returned from FilterHotWalletSet and is used to iterate over the raw logs and unpacked data for HotWalletSet events raised by the Users contract.
type UsersHotWalletSetIterator struct {
	Event *UsersHotWalletSet // Event containing the contract specifics and raw log

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
func (it *UsersHotWalletSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersHotWalletSet)
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
		it.Event = new(UsersHotWalletSet)
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
func (it *UsersHotWalletSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersHotWalletSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersHotWalletSet represents a HotWalletSet event raised by the Users contract.
type UsersHotWalletSet struct {
	Wallet common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterHotWalletSet is a free log retrieval operation binding the contract event 0x6f67ba524a8b4aeff0f137d0089ce5a866653b727943b409c166ec963873775f.
//
// Solidity: e HotWalletSet(_wallet address)
func (_Users *UsersFilterer) FilterHotWalletSet(opts *bind.FilterOpts) (*UsersHotWalletSetIterator, error) {

	logs, sub, err := _Users.contract.FilterLogs(opts, "HotWalletSet")
	if err != nil {
		return nil, err
	}
	return &UsersHotWalletSetIterator{contract: _Users.contract, event: "HotWalletSet", logs: logs, sub: sub}, nil
}

// WatchHotWalletSet is a free log subscription operation binding the contract event 0x6f67ba524a8b4aeff0f137d0089ce5a866653b727943b409c166ec963873775f.
//
// Solidity: e HotWalletSet(_wallet address)
func (_Users *UsersFilterer) WatchHotWalletSet(opts *bind.WatchOpts, sink chan<- *UsersHotWalletSet) (event.Subscription, error) {

	logs, sub, err := _Users.contract.WatchLogs(opts, "HotWalletSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersHotWalletSet)
				if err := _Users.contract.UnpackLog(event, "HotWalletSet", log); err != nil {
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

// UsersPaymentProcessorSetIterator is returned from FilterPaymentProcessorSet and is used to iterate over the raw logs and unpacked data for PaymentProcessorSet events raised by the Users contract.
type UsersPaymentProcessorSetIterator struct {
	Event *UsersPaymentProcessorSet // Event containing the contract specifics and raw log

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
func (it *UsersPaymentProcessorSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersPaymentProcessorSet)
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
		it.Event = new(UsersPaymentProcessorSet)
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
func (it *UsersPaymentProcessorSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersPaymentProcessorSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersPaymentProcessorSet represents a PaymentProcessorSet event raised by the Users contract.
type UsersPaymentProcessorSet struct {
	PaymentProcessor common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterPaymentProcessorSet is a free log retrieval operation binding the contract event 0x8985ba152810f26e00d857552aaff0a9cae15a4fe9bbeb9ec0be19e3e1f064db.
//
// Solidity: e PaymentProcessorSet(_paymentProcessor address)
func (_Users *UsersFilterer) FilterPaymentProcessorSet(opts *bind.FilterOpts) (*UsersPaymentProcessorSetIterator, error) {

	logs, sub, err := _Users.contract.FilterLogs(opts, "PaymentProcessorSet")
	if err != nil {
		return nil, err
	}
	return &UsersPaymentProcessorSetIterator{contract: _Users.contract, event: "PaymentProcessorSet", logs: logs, sub: sub}, nil
}

// WatchPaymentProcessorSet is a free log subscription operation binding the contract event 0x8985ba152810f26e00d857552aaff0a9cae15a4fe9bbeb9ec0be19e3e1f064db.
//
// Solidity: e PaymentProcessorSet(_paymentProcessor address)
func (_Users *UsersFilterer) WatchPaymentProcessorSet(opts *bind.WatchOpts, sink chan<- *UsersPaymentProcessorSet) (event.Subscription, error) {

	logs, sub, err := _Users.contract.WatchLogs(opts, "PaymentProcessorSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersPaymentProcessorSet)
				if err := _Users.contract.UnpackLog(event, "PaymentProcessorSet", log); err != nil {
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

// UsersRTCInterfaceSetIterator is returned from FilterRTCInterfaceSet and is used to iterate over the raw logs and unpacked data for RTCInterfaceSet events raised by the Users contract.
type UsersRTCInterfaceSetIterator struct {
	Event *UsersRTCInterfaceSet // Event containing the contract specifics and raw log

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
func (it *UsersRTCInterfaceSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersRTCInterfaceSet)
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
		it.Event = new(UsersRTCInterfaceSet)
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
func (it *UsersRTCInterfaceSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersRTCInterfaceSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersRTCInterfaceSet represents a RTCInterfaceSet event raised by the Users contract.
type UsersRTCInterfaceSet struct {
	TokenContractAddress common.Address
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterRTCInterfaceSet is a free log retrieval operation binding the contract event 0x08481da9fb8764513d40cbc8a3c7eabb80f438636910a1462b703d49f0d4c6d8.
//
// Solidity: e RTCInterfaceSet(_tokenContractAddress address)
func (_Users *UsersFilterer) FilterRTCInterfaceSet(opts *bind.FilterOpts) (*UsersRTCInterfaceSetIterator, error) {

	logs, sub, err := _Users.contract.FilterLogs(opts, "RTCInterfaceSet")
	if err != nil {
		return nil, err
	}
	return &UsersRTCInterfaceSetIterator{contract: _Users.contract, event: "RTCInterfaceSet", logs: logs, sub: sub}, nil
}

// WatchRTCInterfaceSet is a free log subscription operation binding the contract event 0x08481da9fb8764513d40cbc8a3c7eabb80f438636910a1462b703d49f0d4c6d8.
//
// Solidity: e RTCInterfaceSet(_tokenContractAddress address)
func (_Users *UsersFilterer) WatchRTCInterfaceSet(opts *bind.WatchOpts, sink chan<- *UsersRTCInterfaceSet) (event.Subscription, error) {

	logs, sub, err := _Users.contract.WatchLogs(opts, "RTCInterfaceSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersRTCInterfaceSet)
				if err := _Users.contract.UnpackLog(event, "RTCInterfaceSet", log); err != nil {
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

// UsersRtcDepositedIterator is returned from FilterRtcDeposited and is used to iterate over the raw logs and unpacked data for RtcDeposited events raised by the Users contract.
type UsersRtcDepositedIterator struct {
	Event *UsersRtcDeposited // Event containing the contract specifics and raw log

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
func (it *UsersRtcDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersRtcDeposited)
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
		it.Event = new(UsersRtcDeposited)
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
func (it *UsersRtcDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersRtcDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersRtcDeposited represents a RtcDeposited event raised by the Users contract.
type UsersRtcDeposited struct {
	Uploader common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRtcDeposited is a free log retrieval operation binding the contract event 0x141066711d4e165efac93efabd7eaa910cd70210721bb432e2d8c194dc2139ca.
//
// Solidity: e RtcDeposited(_uploader indexed address, _amount uint256)
func (_Users *UsersFilterer) FilterRtcDeposited(opts *bind.FilterOpts, _uploader []common.Address) (*UsersRtcDepositedIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.FilterLogs(opts, "RtcDeposited", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &UsersRtcDepositedIterator{contract: _Users.contract, event: "RtcDeposited", logs: logs, sub: sub}, nil
}

// WatchRtcDeposited is a free log subscription operation binding the contract event 0x141066711d4e165efac93efabd7eaa910cd70210721bb432e2d8c194dc2139ca.
//
// Solidity: e RtcDeposited(_uploader indexed address, _amount uint256)
func (_Users *UsersFilterer) WatchRtcDeposited(opts *bind.WatchOpts, sink chan<- *UsersRtcDeposited, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.WatchLogs(opts, "RtcDeposited", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersRtcDeposited)
				if err := _Users.contract.UnpackLog(event, "RtcDeposited", log); err != nil {
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

// UsersRtcPaymentWithdrawnForUploaderIterator is returned from FilterRtcPaymentWithdrawnForUploader and is used to iterate over the raw logs and unpacked data for RtcPaymentWithdrawnForUploader events raised by the Users contract.
type UsersRtcPaymentWithdrawnForUploaderIterator struct {
	Event *UsersRtcPaymentWithdrawnForUploader // Event containing the contract specifics and raw log

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
func (it *UsersRtcPaymentWithdrawnForUploaderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersRtcPaymentWithdrawnForUploader)
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
		it.Event = new(UsersRtcPaymentWithdrawnForUploader)
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
func (it *UsersRtcPaymentWithdrawnForUploaderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersRtcPaymentWithdrawnForUploaderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersRtcPaymentWithdrawnForUploader represents a RtcPaymentWithdrawnForUploader event raised by the Users contract.
type UsersRtcPaymentWithdrawnForUploader struct {
	Uploader  common.Address
	Amount    *big.Int
	HashedCID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRtcPaymentWithdrawnForUploader is a free log retrieval operation binding the contract event 0x23b8eb1a184ac06193ae1da4a63994ffcb68da99c46802f3ba69876de66eaa3e.
//
// Solidity: e RtcPaymentWithdrawnForUploader(_uploader indexed address, _amount uint256, _hashedCID bytes32)
func (_Users *UsersFilterer) FilterRtcPaymentWithdrawnForUploader(opts *bind.FilterOpts, _uploader []common.Address) (*UsersRtcPaymentWithdrawnForUploaderIterator, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.FilterLogs(opts, "RtcPaymentWithdrawnForUploader", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return &UsersRtcPaymentWithdrawnForUploaderIterator{contract: _Users.contract, event: "RtcPaymentWithdrawnForUploader", logs: logs, sub: sub}, nil
}

// WatchRtcPaymentWithdrawnForUploader is a free log subscription operation binding the contract event 0x23b8eb1a184ac06193ae1da4a63994ffcb68da99c46802f3ba69876de66eaa3e.
//
// Solidity: e RtcPaymentWithdrawnForUploader(_uploader indexed address, _amount uint256, _hashedCID bytes32)
func (_Users *UsersFilterer) WatchRtcPaymentWithdrawnForUploader(opts *bind.WatchOpts, sink chan<- *UsersRtcPaymentWithdrawnForUploader, _uploader []common.Address) (event.Subscription, error) {

	var _uploaderRule []interface{}
	for _, _uploaderItem := range _uploader {
		_uploaderRule = append(_uploaderRule, _uploaderItem)
	}

	logs, sub, err := _Users.contract.WatchLogs(opts, "RtcPaymentWithdrawnForUploader", _uploaderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersRtcPaymentWithdrawnForUploader)
				if err := _Users.contract.UnpackLog(event, "RtcPaymentWithdrawnForUploader", log); err != nil {
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

// UsersUserRegisteredIterator is returned from FilterUserRegistered and is used to iterate over the raw logs and unpacked data for UserRegistered events raised by the Users contract.
type UsersUserRegisteredIterator struct {
	Event *UsersUserRegistered // Event containing the contract specifics and raw log

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
func (it *UsersUserRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsersUserRegistered)
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
		it.Event = new(UsersUserRegistered)
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
func (it *UsersUserRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsersUserRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsersUserRegistered represents a UserRegistered event raised by the Users contract.
type UsersUserRegistered struct {
	Uploader common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUserRegistered is a free log retrieval operation binding the contract event 0x54db7a5cb4735e1aac1f53db512d3390390bb6637bd30ad4bf9fc98667d9b9b9.
//
// Solidity: e UserRegistered(_uploader address)
func (_Users *UsersFilterer) FilterUserRegistered(opts *bind.FilterOpts) (*UsersUserRegisteredIterator, error) {

	logs, sub, err := _Users.contract.FilterLogs(opts, "UserRegistered")
	if err != nil {
		return nil, err
	}
	return &UsersUserRegisteredIterator{contract: _Users.contract, event: "UserRegistered", logs: logs, sub: sub}, nil
}

// WatchUserRegistered is a free log subscription operation binding the contract event 0x54db7a5cb4735e1aac1f53db512d3390390bb6637bd30ad4bf9fc98667d9b9b9.
//
// Solidity: e UserRegistered(_uploader address)
func (_Users *UsersFilterer) WatchUserRegistered(opts *bind.WatchOpts, sink chan<- *UsersUserRegistered) (event.Subscription, error) {

	logs, sub, err := _Users.contract.WatchLogs(opts, "UserRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsersUserRegistered)
				if err := _Users.contract.UnpackLog(event, "UserRegistered", log); err != nil {
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
