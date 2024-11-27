// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package gov

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// NCPListMetaData contains all meta data concerning the NCPList contract.
var NCPListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_ncpList\",\"type\":\"address[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"ncp\",\"type\":\"address\"}],\"name\":\"NCPAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"ncp\",\"type\":\"address\"}],\"name\":\"NCPRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalType\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"ncp\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"}],\"name\":\"NewProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_proposalID\",\"type\":\"uint256\"}],\"name\":\"ProposalCanceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"proposalID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"accepted\",\"type\":\"bool\"}],\"name\":\"ProposalFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"proposalID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"accept\",\"type\":\"bool\"}],\"name\":\"Vote\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"VOTING_PERIOD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_proposalID\",\"type\":\"uint256\"}],\"name\":\"cancelVote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentProposalID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ncp\",\"type\":\"address\"}],\"name\":\"isNCP\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ncpList\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_newNCP\",\"type\":\"address\"}],\"name\":\"newProposalToAddNCP\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ncp\",\"type\":\"address\"}],\"name\":\"newProposalToRemoveNCP\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_proposalID\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"_accept\",\"type\":\"bool\"}],\"name\":\"vote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"b1610d7e": "VOTING_PERIOD()",
		"bacbe2da": "cancelVote(uint256)",
		"63966190": "currentProposalID()",
		"701693e3": "isNCP(address)",
		"34afb7bd": "ncpList()",
		"a998cb3c": "newProposalToAddNCP(address)",
		"fba44fe0": "newProposalToRemoveNCP(address)",
		"c9d27afe": "vote(uint256,bool)",
	},
	Bin: "0x60806040523480156200001157600080fd5b5060405162000edd38038062000edd83398101604081905262000034916200013b565b60005b81518110156200008e57620000788282815181106200005a576200005a6200020d565b602002602001015160006200009660201b620004da1790919060201c565b5080620000858162000223565b91505062000037565b50506200024b565b6000620000ad836001600160a01b038416620000b6565b90505b92915050565b6000818152600183016020526040812054620000ff57508154600181810184556000848152602080822090930184905584548482528286019093526040902091909155620000b0565b506000620000b0565b634e487b7160e01b600052604160045260246000fd5b80516001600160a01b03811681146200013657600080fd5b919050565b600060208083850312156200014f57600080fd5b82516001600160401b03808211156200016757600080fd5b818501915085601f8301126200017c57600080fd5b81518181111562000191576200019162000108565b8060051b604051601f19603f83011681018181108582111715620001b957620001b962000108565b604052918252848201925083810185019188831115620001d857600080fd5b938501935b828510156200020157620001f1856200011e565b84529385019392850192620001dd565b98975050505050505050565b634e487b7160e01b600052603260045260246000fd5b6000600182016200024457634e487b7160e01b600052601160045260246000fd5b5060010190565b610c82806200025b6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063b1610d7e1161005b578063b1610d7e146100fa578063bacbe2da14610104578063c9d27afe14610117578063fba44fe01461012a57600080fd5b806334afb7bd1461008d57806363966190146100ab578063701693e3146100c2578063a998cb3c146100e5575b600080fd5b61009561013d565b6040516100a29190610a9a565b60405180910390f35b6100b460025481565b6040519081526020016100a2565b6100d56100d0366004610ae7565b61014e565b60405190151581526020016100a2565b6100f86100f3366004610ae7565b610160565b005b6100b462093a8081565b6100f8610112366004610b10565b6101e3565b6100f8610125366004610b29565b61027e565b6100f8610138366004610ae7565b610463565b606061014960006104f6565b905090565b600061015a8183610503565b92915050565b61016b600033610503565b6101905760405162461bcd60e51b815260040161018790610b5e565b60405180910390fd5b61019b600082610503565b156101d55760405162461bcd60e51b815260206004820152600a6024820152696e63702065786973747360b01b6044820152606401610187565b6101e0816001610525565b50565b6101ee600033610503565b61020a5760405162461bcd60e51b815260040161018790610b5e565b600061021582610697565b90508060020154421180610235575060038101546001600160a01b031633145b6102715760405162461bcd60e51b815260206004820152600d60248201526c18d85b9b9bdd0818d85b98d95b609a1b6044820152606401610187565b61027a81610749565b5050565b610289600033610503565b6102a55760405162461bcd60e51b815260040161018790610b5e565b60006102b083610697565b9050600033600090815260088301602052604090205460ff1660028111156102da576102da610b8d565b146103175760405162461bcd60e51b815260206004820152600d60248201526c185b1c9958591e481d9bdd1959609a1b6044820152606401610187565b806002015442111561032c5761032c81610749565b600082156103535750600681018054600191600061034983610bb9565b919050555061036e565b50600781018054600291600061036883610bb9565b91905055505b600582018054600181810183556000928352602080842090920180546001600160a01b0319163390811790915583526008850190915260409091208054839260ff19909116908360028111156103c6576103c6610b8d565b021790555060408051338152841515602082015285917fcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26910160405180910390a260006104136000610798565b905080836006015460026104279190610bd2565b1180610443575080836007015460026104409190610bd2565b10155b1561045c5761045c8384600701548560060154116107a2565b5050505050565b61046e600033610503565b61048a5760405162461bcd60e51b815260040161018790610b5e565b610495600082610503565b6104cf5760405162461bcd60e51b815260206004820152600b60248201526a0696e76616c6964206e63760ac1b6044820152606401610187565b6101e0816002610525565b60006104ef836001600160a01b0384166108ec565b9392505050565b606060006104ef8361093b565b6001600160a01b038116600090815260018301602052604081205415156104ef565b60025460009081526003602052604081205460ff16600481111561054b5761054b610b8d565b146105985760405162461bcd60e51b815260206004820152601c60248201527f70726576696f757320766f746520697320696e2070726f6772657373000000006044820152606401610187565b600254600090815260036020819052604090912090810180546001600160a01b0319163317905542600182018190556105d59062093a8090610bf1565b6002808301919091556004820180546001600160a01b0386166001600160a01b031982168117835585936001600160a81b03199092161790600160a01b90849081111561062457610624610b8d565b0217905550805460ff1916600117815560028054907fe160e5422f314166cb563967f23c23e17326d41e9a52bc5b565759ebdef4f22090849081111561066c5761066c610b8d565b604080519182526001600160a01b0387166020830152339082015260600160405180910390a2505050565b600060025482146106e05760405162461bcd60e51b81526020600482015260136024820152721a5b9d985b1a59081c1c9bdc1bdcd85b081a59606a1b6044820152606401610187565b5060008181526003602052604090206001815460ff16600481111561070757610707610b8d565b146107445760405162461bcd60e51b815260206004820152600d60248201526c6e6f7420696e20766f74696e6760981b6044820152606401610187565b919050565b805460ff191660041781556002546040517f789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c90600090a26002805490600061079083610bb9565b919050555050565b600061015a825490565b801561088f5760016004830154600160a01b900460ff1660028111156107ca576107ca610b8d565b036108295760048201546107e9906000906001600160a01b03166104da565b5060048201546040516001600160a01b03909116907ffdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a932990600090a261087f565b6004820154610843906000906001600160a01b0316610997565b5060048201546040516001600160a01b03909116907f9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c90600090a25b815460ff1916600217825561089b565b815460ff191660031782555b60025460405182151581527fb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a79060200160405180910390a2600280549060006108e383610bb9565b91905055505050565b60008181526001830160205260408120546109335750815460018181018455600084815260208082209093018490558454848252828601909352604090209190915561015a565b50600061015a565b60608160000180548060200260200160405190810160405280929190818152602001828054801561098b57602002820191906000526020600020905b815481526020019060010190808311610977575b50505050509050919050565b60006104ef836001600160a01b03841660008181526001830160205260408120548015610a905760006109cb600183610c09565b85549091506000906109df90600190610c09565b9050818114610a445760008660000182815481106109ff576109ff610c20565b9060005260206000200154905080876000018481548110610a2257610a22610c20565b6000918252602080832090910192909255918252600188019052604090208390555b8554869080610a5557610a55610c36565b60019003818190600052602060002001600090559055856001016000868152602001908152602001600020600090556001935050505061015a565b600091505061015a565b6020808252825182820181905260009190848201906040850190845b81811015610adb5783516001600160a01b031683529284019291840191600101610ab6565b50909695505050505050565b600060208284031215610af957600080fd5b81356001600160a01b03811681146104ef57600080fd5b600060208284031215610b2257600080fd5b5035919050565b60008060408385031215610b3c57600080fd5b8235915060208301358015158114610b5357600080fd5b809150509250929050565b60208082526015908201527406d73672e73656e646572206973206e6f74206e637605c1b604082015260600190565b634e487b7160e01b600052602160045260246000fd5b634e487b7160e01b600052601160045260246000fd5b600060018201610bcb57610bcb610ba3565b5060010190565b6000816000190483118215151615610bec57610bec610ba3565b500290565b60008219821115610c0457610c04610ba3565b500190565b600082821015610c1b57610c1b610ba3565b500390565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fdfea2646970667358221220a3a0116fb16b045aa3d3e2924b0b921b624b78af8d4f8ef156b07a12b55224d664736f6c634300080e0033",
}

// NCPListABI is the input ABI used to generate the binding from.
// Deprecated: Use NCPListMetaData.ABI instead.
var NCPListABI = NCPListMetaData.ABI

// Deprecated: Use NCPListMetaData.Sigs instead.
// NCPListFuncSigs maps the 4-byte function signature to its string representation.
var NCPListFuncSigs = NCPListMetaData.Sigs

// NCPListBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use NCPListMetaData.Bin instead.
var NCPListBin = NCPListMetaData.Bin

// DeployNCPList deploys a new Ethereum contract, binding an instance of NCPList to it.
func DeployNCPList(auth *bind.TransactOpts, backend bind.ContractBackend, _ncpList []common.Address) (common.Address, *types.Transaction, *NCPList, error) {
	parsed, err := NCPListMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(NCPListBin), backend, _ncpList)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &NCPList{NCPListCaller: NCPListCaller{contract: contract}, NCPListTransactor: NCPListTransactor{contract: contract}, NCPListFilterer: NCPListFilterer{contract: contract}}, nil
}

// NCPList is an auto generated Go binding around an Ethereum contract.
type NCPList struct {
	NCPListCaller     // Read-only binding to the contract
	NCPListTransactor // Write-only binding to the contract
	NCPListFilterer   // Log filterer for contract events
}

// NCPListCaller is an auto generated read-only Go binding around an Ethereum contract.
type NCPListCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NCPListTransactor is an auto generated write-only Go binding around an Ethereum contract.
type NCPListTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NCPListFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type NCPListFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NCPListSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type NCPListSession struct {
	Contract     *NCPList          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// NCPListCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type NCPListCallerSession struct {
	Contract *NCPListCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// NCPListTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type NCPListTransactorSession struct {
	Contract     *NCPListTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// NCPListRaw is an auto generated low-level Go binding around an Ethereum contract.
type NCPListRaw struct {
	Contract *NCPList // Generic contract binding to access the raw methods on
}

// NCPListCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type NCPListCallerRaw struct {
	Contract *NCPListCaller // Generic read-only contract binding to access the raw methods on
}

// NCPListTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type NCPListTransactorRaw struct {
	Contract *NCPListTransactor // Generic write-only contract binding to access the raw methods on
}

// NewNCPList creates a new instance of NCPList, bound to a specific deployed contract.
func NewNCPList(address common.Address, backend bind.ContractBackend) (*NCPList, error) {
	contract, err := bindNCPList(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &NCPList{NCPListCaller: NCPListCaller{contract: contract}, NCPListTransactor: NCPListTransactor{contract: contract}, NCPListFilterer: NCPListFilterer{contract: contract}}, nil
}

// NewNCPListCaller creates a new read-only instance of NCPList, bound to a specific deployed contract.
func NewNCPListCaller(address common.Address, caller bind.ContractCaller) (*NCPListCaller, error) {
	contract, err := bindNCPList(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &NCPListCaller{contract: contract}, nil
}

// NewNCPListTransactor creates a new write-only instance of NCPList, bound to a specific deployed contract.
func NewNCPListTransactor(address common.Address, transactor bind.ContractTransactor) (*NCPListTransactor, error) {
	contract, err := bindNCPList(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &NCPListTransactor{contract: contract}, nil
}

// NewNCPListFilterer creates a new log filterer instance of NCPList, bound to a specific deployed contract.
func NewNCPListFilterer(address common.Address, filterer bind.ContractFilterer) (*NCPListFilterer, error) {
	contract, err := bindNCPList(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &NCPListFilterer{contract: contract}, nil
}

// bindNCPList binds a generic wrapper to an already deployed contract.
func bindNCPList(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := NCPListMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NCPList *NCPListRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NCPList.Contract.NCPListCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NCPList *NCPListRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NCPList.Contract.NCPListTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NCPList *NCPListRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NCPList.Contract.NCPListTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NCPList *NCPListCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NCPList.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NCPList *NCPListTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NCPList.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NCPList *NCPListTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NCPList.Contract.contract.Transact(opts, method, params...)
}

// VOTINGPERIOD is a free data retrieval call binding the contract method 0xb1610d7e.
//
// Solidity: function VOTING_PERIOD() view returns(uint256)
func (_NCPList *NCPListCaller) VOTINGPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _NCPList.contract.Call(opts, &out, "VOTING_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VOTINGPERIOD is a free data retrieval call binding the contract method 0xb1610d7e.
//
// Solidity: function VOTING_PERIOD() view returns(uint256)
func (_NCPList *NCPListSession) VOTINGPERIOD() (*big.Int, error) {
	return _NCPList.Contract.VOTINGPERIOD(&_NCPList.CallOpts)
}

// VOTINGPERIOD is a free data retrieval call binding the contract method 0xb1610d7e.
//
// Solidity: function VOTING_PERIOD() view returns(uint256)
func (_NCPList *NCPListCallerSession) VOTINGPERIOD() (*big.Int, error) {
	return _NCPList.Contract.VOTINGPERIOD(&_NCPList.CallOpts)
}

// CurrentProposalID is a free data retrieval call binding the contract method 0x63966190.
//
// Solidity: function currentProposalID() view returns(uint256)
func (_NCPList *NCPListCaller) CurrentProposalID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _NCPList.contract.Call(opts, &out, "currentProposalID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentProposalID is a free data retrieval call binding the contract method 0x63966190.
//
// Solidity: function currentProposalID() view returns(uint256)
func (_NCPList *NCPListSession) CurrentProposalID() (*big.Int, error) {
	return _NCPList.Contract.CurrentProposalID(&_NCPList.CallOpts)
}

// CurrentProposalID is a free data retrieval call binding the contract method 0x63966190.
//
// Solidity: function currentProposalID() view returns(uint256)
func (_NCPList *NCPListCallerSession) CurrentProposalID() (*big.Int, error) {
	return _NCPList.Contract.CurrentProposalID(&_NCPList.CallOpts)
}

// IsNCP is a free data retrieval call binding the contract method 0x701693e3.
//
// Solidity: function isNCP(address _ncp) view returns(bool)
func (_NCPList *NCPListCaller) IsNCP(opts *bind.CallOpts, _ncp common.Address) (bool, error) {
	var out []interface{}
	err := _NCPList.contract.Call(opts, &out, "isNCP", _ncp)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsNCP is a free data retrieval call binding the contract method 0x701693e3.
//
// Solidity: function isNCP(address _ncp) view returns(bool)
func (_NCPList *NCPListSession) IsNCP(_ncp common.Address) (bool, error) {
	return _NCPList.Contract.IsNCP(&_NCPList.CallOpts, _ncp)
}

// IsNCP is a free data retrieval call binding the contract method 0x701693e3.
//
// Solidity: function isNCP(address _ncp) view returns(bool)
func (_NCPList *NCPListCallerSession) IsNCP(_ncp common.Address) (bool, error) {
	return _NCPList.Contract.IsNCP(&_NCPList.CallOpts, _ncp)
}

// NcpList is a free data retrieval call binding the contract method 0x34afb7bd.
//
// Solidity: function ncpList() view returns(address[])
func (_NCPList *NCPListCaller) NcpList(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _NCPList.contract.Call(opts, &out, "ncpList")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// NcpList is a free data retrieval call binding the contract method 0x34afb7bd.
//
// Solidity: function ncpList() view returns(address[])
func (_NCPList *NCPListSession) NcpList() ([]common.Address, error) {
	return _NCPList.Contract.NcpList(&_NCPList.CallOpts)
}

// NcpList is a free data retrieval call binding the contract method 0x34afb7bd.
//
// Solidity: function ncpList() view returns(address[])
func (_NCPList *NCPListCallerSession) NcpList() ([]common.Address, error) {
	return _NCPList.Contract.NcpList(&_NCPList.CallOpts)
}

// CancelVote is a paid mutator transaction binding the contract method 0xbacbe2da.
//
// Solidity: function cancelVote(uint256 _proposalID) returns()
func (_NCPList *NCPListTransactor) CancelVote(opts *bind.TransactOpts, _proposalID *big.Int) (*types.Transaction, error) {
	return _NCPList.contract.Transact(opts, "cancelVote", _proposalID)
}

// CancelVote is a paid mutator transaction binding the contract method 0xbacbe2da.
//
// Solidity: function cancelVote(uint256 _proposalID) returns()
func (_NCPList *NCPListSession) CancelVote(_proposalID *big.Int) (*types.Transaction, error) {
	return _NCPList.Contract.CancelVote(&_NCPList.TransactOpts, _proposalID)
}

// CancelVote is a paid mutator transaction binding the contract method 0xbacbe2da.
//
// Solidity: function cancelVote(uint256 _proposalID) returns()
func (_NCPList *NCPListTransactorSession) CancelVote(_proposalID *big.Int) (*types.Transaction, error) {
	return _NCPList.Contract.CancelVote(&_NCPList.TransactOpts, _proposalID)
}

// NewProposalToAddNCP is a paid mutator transaction binding the contract method 0xa998cb3c.
//
// Solidity: function newProposalToAddNCP(address _newNCP) returns()
func (_NCPList *NCPListTransactor) NewProposalToAddNCP(opts *bind.TransactOpts, _newNCP common.Address) (*types.Transaction, error) {
	return _NCPList.contract.Transact(opts, "newProposalToAddNCP", _newNCP)
}

// NewProposalToAddNCP is a paid mutator transaction binding the contract method 0xa998cb3c.
//
// Solidity: function newProposalToAddNCP(address _newNCP) returns()
func (_NCPList *NCPListSession) NewProposalToAddNCP(_newNCP common.Address) (*types.Transaction, error) {
	return _NCPList.Contract.NewProposalToAddNCP(&_NCPList.TransactOpts, _newNCP)
}

// NewProposalToAddNCP is a paid mutator transaction binding the contract method 0xa998cb3c.
//
// Solidity: function newProposalToAddNCP(address _newNCP) returns()
func (_NCPList *NCPListTransactorSession) NewProposalToAddNCP(_newNCP common.Address) (*types.Transaction, error) {
	return _NCPList.Contract.NewProposalToAddNCP(&_NCPList.TransactOpts, _newNCP)
}

// NewProposalToRemoveNCP is a paid mutator transaction binding the contract method 0xfba44fe0.
//
// Solidity: function newProposalToRemoveNCP(address _ncp) returns()
func (_NCPList *NCPListTransactor) NewProposalToRemoveNCP(opts *bind.TransactOpts, _ncp common.Address) (*types.Transaction, error) {
	return _NCPList.contract.Transact(opts, "newProposalToRemoveNCP", _ncp)
}

// NewProposalToRemoveNCP is a paid mutator transaction binding the contract method 0xfba44fe0.
//
// Solidity: function newProposalToRemoveNCP(address _ncp) returns()
func (_NCPList *NCPListSession) NewProposalToRemoveNCP(_ncp common.Address) (*types.Transaction, error) {
	return _NCPList.Contract.NewProposalToRemoveNCP(&_NCPList.TransactOpts, _ncp)
}

// NewProposalToRemoveNCP is a paid mutator transaction binding the contract method 0xfba44fe0.
//
// Solidity: function newProposalToRemoveNCP(address _ncp) returns()
func (_NCPList *NCPListTransactorSession) NewProposalToRemoveNCP(_ncp common.Address) (*types.Transaction, error) {
	return _NCPList.Contract.NewProposalToRemoveNCP(&_NCPList.TransactOpts, _ncp)
}

// Vote is a paid mutator transaction binding the contract method 0xc9d27afe.
//
// Solidity: function vote(uint256 _proposalID, bool _accept) returns()
func (_NCPList *NCPListTransactor) Vote(opts *bind.TransactOpts, _proposalID *big.Int, _accept bool) (*types.Transaction, error) {
	return _NCPList.contract.Transact(opts, "vote", _proposalID, _accept)
}

// Vote is a paid mutator transaction binding the contract method 0xc9d27afe.
//
// Solidity: function vote(uint256 _proposalID, bool _accept) returns()
func (_NCPList *NCPListSession) Vote(_proposalID *big.Int, _accept bool) (*types.Transaction, error) {
	return _NCPList.Contract.Vote(&_NCPList.TransactOpts, _proposalID, _accept)
}

// Vote is a paid mutator transaction binding the contract method 0xc9d27afe.
//
// Solidity: function vote(uint256 _proposalID, bool _accept) returns()
func (_NCPList *NCPListTransactorSession) Vote(_proposalID *big.Int, _accept bool) (*types.Transaction, error) {
	return _NCPList.Contract.Vote(&_NCPList.TransactOpts, _proposalID, _accept)
}

// NCPListNCPAddedIterator is returned from FilterNCPAdded and is used to iterate over the raw logs and unpacked data for NCPAdded events raised by the NCPList contract.
type NCPListNCPAddedIterator struct {
	Event *NCPListNCPAdded // Event containing the contract specifics and raw log

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
func (it *NCPListNCPAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NCPListNCPAdded)
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
		it.Event = new(NCPListNCPAdded)
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
func (it *NCPListNCPAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NCPListNCPAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NCPListNCPAdded represents a NCPAdded event raised by the NCPList contract.
type NCPListNCPAdded struct {
	Ncp common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterNCPAdded is a free log retrieval operation binding the contract event 0xfdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a9329.
//
// Solidity: event NCPAdded(address indexed ncp)
func (_NCPList *NCPListFilterer) FilterNCPAdded(opts *bind.FilterOpts, ncp []common.Address) (*NCPListNCPAddedIterator, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _NCPList.contract.FilterLogs(opts, "NCPAdded", ncpRule)
	if err != nil {
		return nil, err
	}
	return &NCPListNCPAddedIterator{contract: _NCPList.contract, event: "NCPAdded", logs: logs, sub: sub}, nil
}

// WatchNCPAdded is a free log subscription operation binding the contract event 0xfdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a9329.
//
// Solidity: event NCPAdded(address indexed ncp)
func (_NCPList *NCPListFilterer) WatchNCPAdded(opts *bind.WatchOpts, sink chan<- *NCPListNCPAdded, ncp []common.Address) (event.Subscription, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _NCPList.contract.WatchLogs(opts, "NCPAdded", ncpRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NCPListNCPAdded)
				if err := _NCPList.contract.UnpackLog(event, "NCPAdded", log); err != nil {
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

// ParseNCPAdded is a log parse operation binding the contract event 0xfdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a9329.
//
// Solidity: event NCPAdded(address indexed ncp)
func (_NCPList *NCPListFilterer) ParseNCPAdded(log types.Log) (*NCPListNCPAdded, error) {
	event := new(NCPListNCPAdded)
	if err := _NCPList.contract.UnpackLog(event, "NCPAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NCPListNCPRemovedIterator is returned from FilterNCPRemoved and is used to iterate over the raw logs and unpacked data for NCPRemoved events raised by the NCPList contract.
type NCPListNCPRemovedIterator struct {
	Event *NCPListNCPRemoved // Event containing the contract specifics and raw log

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
func (it *NCPListNCPRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NCPListNCPRemoved)
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
		it.Event = new(NCPListNCPRemoved)
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
func (it *NCPListNCPRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NCPListNCPRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NCPListNCPRemoved represents a NCPRemoved event raised by the NCPList contract.
type NCPListNCPRemoved struct {
	Ncp common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterNCPRemoved is a free log retrieval operation binding the contract event 0x9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c.
//
// Solidity: event NCPRemoved(address indexed ncp)
func (_NCPList *NCPListFilterer) FilterNCPRemoved(opts *bind.FilterOpts, ncp []common.Address) (*NCPListNCPRemovedIterator, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _NCPList.contract.FilterLogs(opts, "NCPRemoved", ncpRule)
	if err != nil {
		return nil, err
	}
	return &NCPListNCPRemovedIterator{contract: _NCPList.contract, event: "NCPRemoved", logs: logs, sub: sub}, nil
}

// WatchNCPRemoved is a free log subscription operation binding the contract event 0x9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c.
//
// Solidity: event NCPRemoved(address indexed ncp)
func (_NCPList *NCPListFilterer) WatchNCPRemoved(opts *bind.WatchOpts, sink chan<- *NCPListNCPRemoved, ncp []common.Address) (event.Subscription, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _NCPList.contract.WatchLogs(opts, "NCPRemoved", ncpRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NCPListNCPRemoved)
				if err := _NCPList.contract.UnpackLog(event, "NCPRemoved", log); err != nil {
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

// ParseNCPRemoved is a log parse operation binding the contract event 0x9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c.
//
// Solidity: event NCPRemoved(address indexed ncp)
func (_NCPList *NCPListFilterer) ParseNCPRemoved(log types.Log) (*NCPListNCPRemoved, error) {
	event := new(NCPListNCPRemoved)
	if err := _NCPList.contract.UnpackLog(event, "NCPRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NCPListNewProposalIterator is returned from FilterNewProposal and is used to iterate over the raw logs and unpacked data for NewProposal events raised by the NCPList contract.
type NCPListNewProposalIterator struct {
	Event *NCPListNewProposal // Event containing the contract specifics and raw log

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
func (it *NCPListNewProposalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NCPListNewProposal)
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
		it.Event = new(NCPListNewProposal)
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
func (it *NCPListNewProposalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NCPListNewProposalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NCPListNewProposal represents a NewProposal event raised by the NCPList contract.
type NCPListNewProposal struct {
	Id           *big.Int
	ProposalType *big.Int
	Ncp          common.Address
	Proposer     common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterNewProposal is a free log retrieval operation binding the contract event 0xe160e5422f314166cb563967f23c23e17326d41e9a52bc5b565759ebdef4f220.
//
// Solidity: event NewProposal(uint256 indexed id, uint256 proposalType, address ncp, address proposer)
func (_NCPList *NCPListFilterer) FilterNewProposal(opts *bind.FilterOpts, id []*big.Int) (*NCPListNewProposalIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _NCPList.contract.FilterLogs(opts, "NewProposal", idRule)
	if err != nil {
		return nil, err
	}
	return &NCPListNewProposalIterator{contract: _NCPList.contract, event: "NewProposal", logs: logs, sub: sub}, nil
}

// WatchNewProposal is a free log subscription operation binding the contract event 0xe160e5422f314166cb563967f23c23e17326d41e9a52bc5b565759ebdef4f220.
//
// Solidity: event NewProposal(uint256 indexed id, uint256 proposalType, address ncp, address proposer)
func (_NCPList *NCPListFilterer) WatchNewProposal(opts *bind.WatchOpts, sink chan<- *NCPListNewProposal, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _NCPList.contract.WatchLogs(opts, "NewProposal", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NCPListNewProposal)
				if err := _NCPList.contract.UnpackLog(event, "NewProposal", log); err != nil {
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

// ParseNewProposal is a log parse operation binding the contract event 0xe160e5422f314166cb563967f23c23e17326d41e9a52bc5b565759ebdef4f220.
//
// Solidity: event NewProposal(uint256 indexed id, uint256 proposalType, address ncp, address proposer)
func (_NCPList *NCPListFilterer) ParseNewProposal(log types.Log) (*NCPListNewProposal, error) {
	event := new(NCPListNewProposal)
	if err := _NCPList.contract.UnpackLog(event, "NewProposal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NCPListProposalCanceledIterator is returned from FilterProposalCanceled and is used to iterate over the raw logs and unpacked data for ProposalCanceled events raised by the NCPList contract.
type NCPListProposalCanceledIterator struct {
	Event *NCPListProposalCanceled // Event containing the contract specifics and raw log

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
func (it *NCPListProposalCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NCPListProposalCanceled)
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
		it.Event = new(NCPListProposalCanceled)
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
func (it *NCPListProposalCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NCPListProposalCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NCPListProposalCanceled represents a ProposalCanceled event raised by the NCPList contract.
type NCPListProposalCanceled struct {
	ProposalID *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalCanceled is a free log retrieval operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 indexed _proposalID)
func (_NCPList *NCPListFilterer) FilterProposalCanceled(opts *bind.FilterOpts, _proposalID []*big.Int) (*NCPListProposalCanceledIterator, error) {

	var _proposalIDRule []interface{}
	for _, _proposalIDItem := range _proposalID {
		_proposalIDRule = append(_proposalIDRule, _proposalIDItem)
	}

	logs, sub, err := _NCPList.contract.FilterLogs(opts, "ProposalCanceled", _proposalIDRule)
	if err != nil {
		return nil, err
	}
	return &NCPListProposalCanceledIterator{contract: _NCPList.contract, event: "ProposalCanceled", logs: logs, sub: sub}, nil
}

// WatchProposalCanceled is a free log subscription operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 indexed _proposalID)
func (_NCPList *NCPListFilterer) WatchProposalCanceled(opts *bind.WatchOpts, sink chan<- *NCPListProposalCanceled, _proposalID []*big.Int) (event.Subscription, error) {

	var _proposalIDRule []interface{}
	for _, _proposalIDItem := range _proposalID {
		_proposalIDRule = append(_proposalIDRule, _proposalIDItem)
	}

	logs, sub, err := _NCPList.contract.WatchLogs(opts, "ProposalCanceled", _proposalIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NCPListProposalCanceled)
				if err := _NCPList.contract.UnpackLog(event, "ProposalCanceled", log); err != nil {
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

// ParseProposalCanceled is a log parse operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 indexed _proposalID)
func (_NCPList *NCPListFilterer) ParseProposalCanceled(log types.Log) (*NCPListProposalCanceled, error) {
	event := new(NCPListProposalCanceled)
	if err := _NCPList.contract.UnpackLog(event, "ProposalCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NCPListProposalFinalizedIterator is returned from FilterProposalFinalized and is used to iterate over the raw logs and unpacked data for ProposalFinalized events raised by the NCPList contract.
type NCPListProposalFinalizedIterator struct {
	Event *NCPListProposalFinalized // Event containing the contract specifics and raw log

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
func (it *NCPListProposalFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NCPListProposalFinalized)
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
		it.Event = new(NCPListProposalFinalized)
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
func (it *NCPListProposalFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NCPListProposalFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NCPListProposalFinalized represents a ProposalFinalized event raised by the NCPList contract.
type NCPListProposalFinalized struct {
	ProposalID *big.Int
	Accepted   bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalFinalized is a free log retrieval operation binding the contract event 0xb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a7.
//
// Solidity: event ProposalFinalized(uint256 indexed proposalID, bool accepted)
func (_NCPList *NCPListFilterer) FilterProposalFinalized(opts *bind.FilterOpts, proposalID []*big.Int) (*NCPListProposalFinalizedIterator, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _NCPList.contract.FilterLogs(opts, "ProposalFinalized", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return &NCPListProposalFinalizedIterator{contract: _NCPList.contract, event: "ProposalFinalized", logs: logs, sub: sub}, nil
}

// WatchProposalFinalized is a free log subscription operation binding the contract event 0xb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a7.
//
// Solidity: event ProposalFinalized(uint256 indexed proposalID, bool accepted)
func (_NCPList *NCPListFilterer) WatchProposalFinalized(opts *bind.WatchOpts, sink chan<- *NCPListProposalFinalized, proposalID []*big.Int) (event.Subscription, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _NCPList.contract.WatchLogs(opts, "ProposalFinalized", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NCPListProposalFinalized)
				if err := _NCPList.contract.UnpackLog(event, "ProposalFinalized", log); err != nil {
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

// ParseProposalFinalized is a log parse operation binding the contract event 0xb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a7.
//
// Solidity: event ProposalFinalized(uint256 indexed proposalID, bool accepted)
func (_NCPList *NCPListFilterer) ParseProposalFinalized(log types.Log) (*NCPListProposalFinalized, error) {
	event := new(NCPListProposalFinalized)
	if err := _NCPList.contract.UnpackLog(event, "ProposalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NCPListVoteIterator is returned from FilterVote and is used to iterate over the raw logs and unpacked data for Vote events raised by the NCPList contract.
type NCPListVoteIterator struct {
	Event *NCPListVote // Event containing the contract specifics and raw log

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
func (it *NCPListVoteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NCPListVote)
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
		it.Event = new(NCPListVote)
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
func (it *NCPListVoteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NCPListVoteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NCPListVote represents a Vote event raised by the NCPList contract.
type NCPListVote struct {
	ProposalID *big.Int
	Voter      common.Address
	Accept     bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterVote is a free log retrieval operation binding the contract event 0xcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26.
//
// Solidity: event Vote(uint256 indexed proposalID, address voter, bool accept)
func (_NCPList *NCPListFilterer) FilterVote(opts *bind.FilterOpts, proposalID []*big.Int) (*NCPListVoteIterator, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _NCPList.contract.FilterLogs(opts, "Vote", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return &NCPListVoteIterator{contract: _NCPList.contract, event: "Vote", logs: logs, sub: sub}, nil
}

// WatchVote is a free log subscription operation binding the contract event 0xcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26.
//
// Solidity: event Vote(uint256 indexed proposalID, address voter, bool accept)
func (_NCPList *NCPListFilterer) WatchVote(opts *bind.WatchOpts, sink chan<- *NCPListVote, proposalID []*big.Int) (event.Subscription, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _NCPList.contract.WatchLogs(opts, "Vote", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NCPListVote)
				if err := _NCPList.contract.UnpackLog(event, "Vote", log); err != nil {
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

// ParseVote is a log parse operation binding the contract event 0xcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26.
//
// Solidity: event Vote(uint256 indexed proposalID, address voter, bool accept)
func (_NCPList *NCPListFilterer) ParseVote(log types.Log) (*NCPListVote, error) {
	event := new(NCPListVote)
	if err := _NCPList.contract.UnpackLog(event, "Vote", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
