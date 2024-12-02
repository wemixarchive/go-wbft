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

// GovStakingMetaData contains all meta data concerning the GovStaking contract.
var GovStakingMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Delegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"credentialID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unbonding\",\"type\":\"uint256\"}],\"name\":\"NewCredential\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Undelegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"reward\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"}],\"name\":\"ValidatorRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"}],\"name\":\"ValidatorRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"credentialID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrew\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAXIMUM_STAKING\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINIMUM_STAKING\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UNBONDING_PERIOD_DELEGATOR\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UNBONDING_PERIOD_VALIDATOR\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"credentialCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"credentials\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"requestTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"withdrawableTime\",\"type\":\"uint256\"},{\"internalType\":\"enumGovStaking.WithdrawalStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"delegateTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"isStakerOrReward\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"isValidator\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_reward\",\"type\":\"address\"}],\"name\":\"registerValidator\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalStaking\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"undelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validatorByReward\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validatorByStaker\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validatorInfo\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"reward\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"delegated\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validatorLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_cid\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"129060ab": "MAXIMUM_STAKING()",
		"ba631d3f": "MINIMUM_STAKING()",
		"840c1771": "UNBONDING_PERIOD_DELEGATOR()",
		"f90aa6ca": "UNBONDING_PERIOD_VALIDATOR()",
		"cd0e35b9": "credentialCount()",
		"e0574e3f": "credentials(uint256)",
		"026e402b": "delegate(address,uint256)",
		"438bb7e5": "delegateTo(address,address)",
		"db463214": "isStakerOrReward(address)",
		"facd743b": "isValidator(address)",
		"0cd4b40f": "registerValidator(uint256,address,address)",
		"a694fc3a": "stake(uint256)",
		"165defa4": "totalStaking()",
		"4d99dd16": "undelegate(address,uint256)",
		"2e17de78": "unstake(uint256)",
		"9f7f7ed2": "validatorByReward(address)",
		"eff39866": "validatorByStaker(address)",
		"4f1811dd": "validatorInfo(address)",
		"aed1d403": "validatorLength()",
		"ca1e7819": "validators()",
		"2e1a7d4d": "withdraw(uint256)",
	},
	Bin: "0x608060405234801561001057600080fd5b506118ae806100206000396000f3fe60806040526004361061012a5760003560e01c80639f7f7ed2116100ab578063cd0e35b91161006f578063cd0e35b91461037c578063db46321414610392578063e0574e3f146103c2578063eff398661461042a578063f90aa6ca14610460578063facd743b1461047657600080fd5b80639f7f7ed2146102c6578063a694fc3a14610314578063aed1d40314610327578063ba631d3f1461033c578063ca1e78191461035a57600080fd5b80632e1a7d4d116100f25780632e1a7d4d146101bb578063438bb7e5146101db5780634d99dd16146102135780634f1811dd14610233578063840c1771146102af57600080fd5b8063026e402b1461012f5780630cd4b40f14610144578063129060ab14610157578063165defa4146101855780632e17de781461019b575b600080fd5b61014261013d366004611634565b610496565b005b61014261015236600461165e565b6105f4565b34801561016357600080fd5b506101726001600160801b0381565b6040519081526020015b60405180910390f35b34801561019157600080fd5b5061017260005481565b3480156101a757600080fd5b506101426101b636600461169a565b6109fc565b3480156101c757600080fd5b506101426101d636600461169a565b610cb3565b3480156101e757600080fd5b506101726101f63660046116b3565b600660209081526000928352604080842090915290825290205481565b34801561021f57600080fd5b5061014261022e366004611634565b610e3b565b34801561023f57600080fd5b5061028461024e3660046116e6565b600360208190526000918252604090912080546001820154600283015492909301546001600160a01b0391821693909116919084565b604080516001600160a01b03958616815294909316602085015291830152606082015260800161017c565b3480156102bb57600080fd5b506101726203f48081565b3480156102d257600080fd5b506102fc6102e13660046116e6565b6005602052600090815260409020546001600160a01b031681565b6040516001600160a01b03909116815260200161017c565b61014261032236600461169a565b610fb4565b34801561033357600080fd5b50610172611043565b34801561034857600080fd5b506101726969e10de76676d080000081565b34801561036657600080fd5b5061036f611054565b60405161017c9190611701565b34801561038857600080fd5b5061017260075481565b34801561039e57600080fd5b506103b26103ad3660046116e6565b611060565b604051901515815260200161017c565b3480156103ce57600080fd5b506104196103dd36600461169a565b600860205260009081526040902080546001820154600283015460038401546004909401546001600160a01b0390931693919290919060ff1685565b60405161017c959493929190611764565b34801561043657600080fd5b506102fc6104453660046116e6565b6004602052600090815260409020546001600160a01b031681565b34801561046c57600080fd5b50610172610e1081565b34801561048257600080fd5b506103b26104913660046116e6565b6110a9565b808034146104bf5760405162461bcd60e51b81526004016104b6906117b7565b60405180910390fd5b6104c8336110a9565b156105155760405162461bcd60e51b815260206004820152601960248201527f76616c696461746f722063616e6e6f742064656c65676174650000000000000060448201526064016104b6565b61051e33611060565b1561056b5760405162461bcd60e51b815260206004820152601e60248201527f7374616b657228726577617264292063616e6e6f742064656c6567617465000060448201526064016104b6565b610577838360016110b6565b3360009081526006602090815260408083206001600160a01b0387168452909152812080548492906105aa908490611804565b90915550506040518281526001600160a01b0384169033907fe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b9060200160405180910390a3505050565b828034146106145760405162461bcd60e51b81526004016104b6906117b7565b6969e10de76676d0800000841015801561063557506001600160801b038411155b6106715760405162461bcd60e51b815260206004820152600d60248201526c6f7574206f6620626f756e647360981b60448201526064016104b6565b336001600160a01b038416148015906106935750336001600160a01b03831614155b6106eb5760405162461bcd60e51b8152602060048201526024808201527f7374616b65722063616e6e6f742062652076616c696461746f72206f722072656044820152631dd85c9960e21b60648201526084016104b6565b6001600160a01b0383161580159061070b57506001600160a01b03821615155b6107465760405162461bcd60e51b815260206004820152600c60248201526b7a65726f206164647265737360a01b60448201526064016104b6565b816001600160a01b0316836001600160a01b0316036107a75760405162461bcd60e51b815260206004820152601a60248201527f76616c696461746f722063616e6e6f742062652072657761726400000000000060448201526064016104b6565b6107b033611060565b156107fd5760405162461bcd60e51b815260206004820152601c60248201527f7374616b657220697320616c726561647920726567697374657265640000000060448201526064016104b6565b61080683611060565b156108535760405162461bcd60e51b815260206004820152601f60248201527f76616c696461746f7220697320616c726561647920726567697374657265640060448201526064016104b6565b61085c82611060565b156108a95760405162461bcd60e51b815260206004820152601c60248201527f72657761726420697320616c726561647920726567697374657265640000000060448201526064016104b6565b6108b46001846111d1565b6108f35760405162461bcd60e51b815260206004820152601060248201526f76616c696461746f722065786973747360801b60448201526064016104b6565b60408051608081018252338082526001600160a01b0385811660208085018281528587018b81526000606088018181528c871680835260038087528b84209a518b54908a166001600160a01b0319918216178c55955160018c01805491909a1690871617909855925160028a01555197909501969096559383526004815285832080548516861790559082526005905292832080549091169091179055805485919081906109a2908490611804565b9091555050604080513381526001600160a01b038481166020830152918101869052908416907f08dc906197021397d84c73d31dcffd0296c7c044af3bb331a3091e122d0b58db906060015b60405180910390a250505050565b336000908152600460205260409020546001600160a01b031680610a5b5760405162461bcd60e51b81526020600482015260166024820152753ab73932b3b4b9ba32b932b2103b30b634b230ba37b960511b60448201526064016104b6565b60008211610a9c5760405162461bcd60e51b815260206004820152600e60248201526d616d6f756e74206973207a65726f60901b60448201526064016104b6565b6001600160a01b03811660009081526003602081905260408220908101546002820154919291610acc919061181c565b905083811015610b155760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b60448201526064016104b6565b6969e10de76676d0800000610b2a858361181c565b1015610c3b57838114610b955760405162461bcd60e51b815260206004820152602d60248201527f616d6f756e74206d75737420657175616c2062616c616e636520746f2072656d60448201526c37bb32903b30b634b230ba37b960991b60648201526084016104b6565b610ba06001846111ed565b5033600090815260046020908152604080832080546001600160a01b03199081169091556001868101546001600160a01b0390811686526005855283862080548416905588168086526003948590528386208054841681559182018054909316909255600281018590559092018390555190917fe1434e25d6611e0db941968fdc97811c982ac1602e951637d206f5fdda9dd8f191a2610c55565b83826002016000828254610c4f919061181c565b90915550505b83600080828254610c66919061181c565b90915550610c78905084610e10611202565b826001600160a01b03167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75856040516109ee91815260200190565b60008181526008602052604090206001600482015460ff166002811115610cdc57610cdc61174e565b14610d1e5760405162461bcd60e51b81526020600482015260126024820152711a5b9d985b1a590818dc9959195b9d1a585b60721b60448201526064016104b6565b80546001600160a01b03163314610d775760405162461bcd60e51b815260206004820152601b60248201527f6d73672e73656e646572206973206e6f7420726571756573746572000000000060448201526064016104b6565b8060030154421015610dcb5760405162461bcd60e51b815260206004820152601860248201527f6e6f74207965742074696d6520746f207769746864726177000000000000000060448201526064016104b6565b60018101548154610de7916001600160a01b0390911690611323565b60048101805460ff19166002179055600181015460408051338152602081019290925283917ff4c56bb5a568ebcb1087fa42400137fd62e67119180ddfe97906e108ecc0e633910160405180910390a25050565b3360009081526006602090815260408083206001600160a01b0386168452909152902054811115610ea55760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b60448201526064016104b6565b610eae826110a9565b15610f15576001600160a01b03821660009081526003602081905260408220908101805491928492610ee190849061181c565b9250508190555081816002016000828254610efc919061181c565b90915550610f0f9050826203f480611202565b50610f1f565b610f1f3382611323565b3360009081526006602090815260408083206001600160a01b038616845290915281208054839290610f5290849061181c565b9250508190555080600080828254610f6a919061181c565b90915550506040518181526001600160a01b0383169033907f4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c906020015b60405180910390a35050565b80803414610fd45760405162461bcd60e51b81526004016104b6906117b7565b336000908152600460205260408120546001600160a01b031690610ffb90829085906110b6565b806001600160a01b03167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d8460405161103691815260200190565b60405180910390a2505050565b600061104f6001611441565b905090565b606061104f600161144b565b6001600160a01b038181166000908152600460205260408120549091161515806110a357506001600160a01b038281166000908152600560205260409020541615155b92915050565b60006110a3600183611458565b6110bf836110a9565b6111045760405162461bcd60e51b81526020600482015260166024820152753ab73932b3b4b9ba32b932b2103b30b634b230ba37b960511b60448201526064016104b6565b6001600160a01b038316600090815260036020526040902060028101546001600160801b0390611135908590611804565b111561117a5760405162461bcd60e51b8152602060048201526014602482015273657863656564656420746865206d6178696d756d60601b60448201526064016104b6565b8260008082825461118b9190611804565b92505081905550828160020160008282546111a69190611804565b909155505081156111cb57828160030160008282546111c59190611804565b90915550505b50505050565b60006111e6836001600160a01b03841661147a565b9392505050565b60006111e6836001600160a01b0384166114c9565b6040518060a00160405280336001600160a01b0316815260200183815260200142815260200182426112349190611804565b815260200160018152506008600060076000815461125190611833565b919050819055815260200190815260200160002060008201518160000160006101000a8154816001600160a01b0302191690836001600160a01b0316021790555060208201518160010155604082015181600201556060820151816003015560808201518160040160006101000a81548160ff021916908360028111156112da576112da61174e565b021790555050600754604080518581524260208201529081018490523392507f4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d90606001610fa8565b804710156113735760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a20696e73756666696369656e742062616c616e636500000060448201526064016104b6565b6000826001600160a01b03168260405160006040518083038185875af1925050503d80600081146113c0576040519150601f19603f3d011682016040523d82523d6000602084013e6113c5565b606091505b505090508061143c5760405162461bcd60e51b815260206004820152603a60248201527f416464726573733a20756e61626c6520746f2073656e642076616c75652c207260448201527f6563697069656e74206d6179206861766520726576657274656400000000000060648201526084016104b6565b505050565b60006110a3825490565b606060006111e6836115bc565b6001600160a01b038116600090815260018301602052604081205415156111e6565b60008181526001830160205260408120546114c1575081546001818101845560008481526020808220909301849055845484825282860190935260409020919091556110a3565b5060006110a3565b600081815260018301602052604081205480156115b25760006114ed60018361181c565b85549091506000906115019060019061181c565b90508181146115665760008660000182815481106115215761152161184c565b90600052602060002001549050808760000184815481106115445761154461184c565b6000918252602080832090910192909255918252600188019052604090208390555b855486908061157757611577611862565b6001900381819060005260206000200160009055905585600101600086815260200190815260200160002060009055600193505050506110a3565b60009150506110a3565b60608160000180548060200260200160405190810160405280929190818152602001828054801561160c57602002820191906000526020600020905b8154815260200190600101908083116115f8575b50505050509050919050565b80356001600160a01b038116811461162f57600080fd5b919050565b6000806040838503121561164757600080fd5b61165083611618565b946020939093013593505050565b60008060006060848603121561167357600080fd5b8335925061168360208501611618565b915061169160408501611618565b90509250925092565b6000602082840312156116ac57600080fd5b5035919050565b600080604083850312156116c657600080fd5b6116cf83611618565b91506116dd60208401611618565b90509250929050565b6000602082840312156116f857600080fd5b6111e682611618565b6020808252825182820181905260009190848201906040850190845b818110156117425783516001600160a01b03168352928401929184019160010161171d565b50909695505050505050565b634e487b7160e01b600052602160045260246000fd5b6001600160a01b038616815260208101859052604081018490526060810183905260a08101600383106117a757634e487b7160e01b600052602160045260246000fd5b8260808301529695505050505050565b6020808252601d908201527f616d6f756e7420616e64206d73672e76616c7565206d69736d61746368000000604082015260600190565b634e487b7160e01b600052601160045260246000fd5b60008219821115611817576118176117ee565b500190565b60008282101561182e5761182e6117ee565b500390565b600060018201611845576118456117ee565b5060010190565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fdfea26469706673582212203bd80e766352862638ca6353224df865a247a3f1148edb5800fdc315ba47785264736f6c634300080e0033",
}

// GovStakingABI is the input ABI used to generate the binding from.
// Deprecated: Use GovStakingMetaData.ABI instead.
var GovStakingABI = GovStakingMetaData.ABI

// Deprecated: Use GovStakingMetaData.Sigs instead.
// GovStakingFuncSigs maps the 4-byte function signature to its string representation.
var GovStakingFuncSigs = GovStakingMetaData.Sigs

// GovStakingBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GovStakingMetaData.Bin instead.
var GovStakingBin = GovStakingMetaData.Bin

// DeployGovStaking deploys a new Ethereum contract, binding an instance of GovStaking to it.
func DeployGovStaking(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GovStaking, error) {
	parsed, err := GovStakingMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GovStakingBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GovStaking{GovStakingCaller: GovStakingCaller{contract: contract}, GovStakingTransactor: GovStakingTransactor{contract: contract}, GovStakingFilterer: GovStakingFilterer{contract: contract}}, nil
}

// GovStaking is an auto generated Go binding around an Ethereum contract.
type GovStaking struct {
	GovStakingCaller     // Read-only binding to the contract
	GovStakingTransactor // Write-only binding to the contract
	GovStakingFilterer   // Log filterer for contract events
}

// GovStakingCaller is an auto generated read-only Go binding around an Ethereum contract.
type GovStakingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovStakingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GovStakingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovStakingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GovStakingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovStakingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GovStakingSession struct {
	Contract     *GovStaking       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GovStakingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GovStakingCallerSession struct {
	Contract *GovStakingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// GovStakingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GovStakingTransactorSession struct {
	Contract     *GovStakingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// GovStakingRaw is an auto generated low-level Go binding around an Ethereum contract.
type GovStakingRaw struct {
	Contract *GovStaking // Generic contract binding to access the raw methods on
}

// GovStakingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GovStakingCallerRaw struct {
	Contract *GovStakingCaller // Generic read-only contract binding to access the raw methods on
}

// GovStakingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GovStakingTransactorRaw struct {
	Contract *GovStakingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGovStaking creates a new instance of GovStaking, bound to a specific deployed contract.
func NewGovStaking(address common.Address, backend bind.ContractBackend) (*GovStaking, error) {
	contract, err := bindGovStaking(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GovStaking{GovStakingCaller: GovStakingCaller{contract: contract}, GovStakingTransactor: GovStakingTransactor{contract: contract}, GovStakingFilterer: GovStakingFilterer{contract: contract}}, nil
}

// NewGovStakingCaller creates a new read-only instance of GovStaking, bound to a specific deployed contract.
func NewGovStakingCaller(address common.Address, caller bind.ContractCaller) (*GovStakingCaller, error) {
	contract, err := bindGovStaking(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GovStakingCaller{contract: contract}, nil
}

// NewGovStakingTransactor creates a new write-only instance of GovStaking, bound to a specific deployed contract.
func NewGovStakingTransactor(address common.Address, transactor bind.ContractTransactor) (*GovStakingTransactor, error) {
	contract, err := bindGovStaking(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GovStakingTransactor{contract: contract}, nil
}

// NewGovStakingFilterer creates a new log filterer instance of GovStaking, bound to a specific deployed contract.
func NewGovStakingFilterer(address common.Address, filterer bind.ContractFilterer) (*GovStakingFilterer, error) {
	contract, err := bindGovStaking(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GovStakingFilterer{contract: contract}, nil
}

// bindGovStaking binds a generic wrapper to an already deployed contract.
func bindGovStaking(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GovStakingMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovStaking *GovStakingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovStaking.Contract.GovStakingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovStaking *GovStakingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovStaking.Contract.GovStakingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovStaking *GovStakingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovStaking.Contract.GovStakingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovStaking *GovStakingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovStaking.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovStaking *GovStakingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovStaking.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovStaking *GovStakingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovStaking.Contract.contract.Transact(opts, method, params...)
}

// MAXIMUMSTAKING is a free data retrieval call binding the contract method 0x129060ab.
//
// Solidity: function MAXIMUM_STAKING() view returns(uint256)
func (_GovStaking *GovStakingCaller) MAXIMUMSTAKING(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "MAXIMUM_STAKING")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXIMUMSTAKING is a free data retrieval call binding the contract method 0x129060ab.
//
// Solidity: function MAXIMUM_STAKING() view returns(uint256)
func (_GovStaking *GovStakingSession) MAXIMUMSTAKING() (*big.Int, error) {
	return _GovStaking.Contract.MAXIMUMSTAKING(&_GovStaking.CallOpts)
}

// MAXIMUMSTAKING is a free data retrieval call binding the contract method 0x129060ab.
//
// Solidity: function MAXIMUM_STAKING() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) MAXIMUMSTAKING() (*big.Int, error) {
	return _GovStaking.Contract.MAXIMUMSTAKING(&_GovStaking.CallOpts)
}

// MINIMUMSTAKING is a free data retrieval call binding the contract method 0xba631d3f.
//
// Solidity: function MINIMUM_STAKING() view returns(uint256)
func (_GovStaking *GovStakingCaller) MINIMUMSTAKING(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "MINIMUM_STAKING")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINIMUMSTAKING is a free data retrieval call binding the contract method 0xba631d3f.
//
// Solidity: function MINIMUM_STAKING() view returns(uint256)
func (_GovStaking *GovStakingSession) MINIMUMSTAKING() (*big.Int, error) {
	return _GovStaking.Contract.MINIMUMSTAKING(&_GovStaking.CallOpts)
}

// MINIMUMSTAKING is a free data retrieval call binding the contract method 0xba631d3f.
//
// Solidity: function MINIMUM_STAKING() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) MINIMUMSTAKING() (*big.Int, error) {
	return _GovStaking.Contract.MINIMUMSTAKING(&_GovStaking.CallOpts)
}

// UNBONDINGPERIODDELEGATOR is a free data retrieval call binding the contract method 0x840c1771.
//
// Solidity: function UNBONDING_PERIOD_DELEGATOR() view returns(uint256)
func (_GovStaking *GovStakingCaller) UNBONDINGPERIODDELEGATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "UNBONDING_PERIOD_DELEGATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UNBONDINGPERIODDELEGATOR is a free data retrieval call binding the contract method 0x840c1771.
//
// Solidity: function UNBONDING_PERIOD_DELEGATOR() view returns(uint256)
func (_GovStaking *GovStakingSession) UNBONDINGPERIODDELEGATOR() (*big.Int, error) {
	return _GovStaking.Contract.UNBONDINGPERIODDELEGATOR(&_GovStaking.CallOpts)
}

// UNBONDINGPERIODDELEGATOR is a free data retrieval call binding the contract method 0x840c1771.
//
// Solidity: function UNBONDING_PERIOD_DELEGATOR() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) UNBONDINGPERIODDELEGATOR() (*big.Int, error) {
	return _GovStaking.Contract.UNBONDINGPERIODDELEGATOR(&_GovStaking.CallOpts)
}

// UNBONDINGPERIODVALIDATOR is a free data retrieval call binding the contract method 0xf90aa6ca.
//
// Solidity: function UNBONDING_PERIOD_VALIDATOR() view returns(uint256)
func (_GovStaking *GovStakingCaller) UNBONDINGPERIODVALIDATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "UNBONDING_PERIOD_VALIDATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UNBONDINGPERIODVALIDATOR is a free data retrieval call binding the contract method 0xf90aa6ca.
//
// Solidity: function UNBONDING_PERIOD_VALIDATOR() view returns(uint256)
func (_GovStaking *GovStakingSession) UNBONDINGPERIODVALIDATOR() (*big.Int, error) {
	return _GovStaking.Contract.UNBONDINGPERIODVALIDATOR(&_GovStaking.CallOpts)
}

// UNBONDINGPERIODVALIDATOR is a free data retrieval call binding the contract method 0xf90aa6ca.
//
// Solidity: function UNBONDING_PERIOD_VALIDATOR() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) UNBONDINGPERIODVALIDATOR() (*big.Int, error) {
	return _GovStaking.Contract.UNBONDINGPERIODVALIDATOR(&_GovStaking.CallOpts)
}

// CredentialCount is a free data retrieval call binding the contract method 0xcd0e35b9.
//
// Solidity: function credentialCount() view returns(uint256)
func (_GovStaking *GovStakingCaller) CredentialCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "credentialCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CredentialCount is a free data retrieval call binding the contract method 0xcd0e35b9.
//
// Solidity: function credentialCount() view returns(uint256)
func (_GovStaking *GovStakingSession) CredentialCount() (*big.Int, error) {
	return _GovStaking.Contract.CredentialCount(&_GovStaking.CallOpts)
}

// CredentialCount is a free data retrieval call binding the contract method 0xcd0e35b9.
//
// Solidity: function credentialCount() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) CredentialCount() (*big.Int, error) {
	return _GovStaking.Contract.CredentialCount(&_GovStaking.CallOpts)
}

// Credentials is a free data retrieval call binding the contract method 0xe0574e3f.
//
// Solidity: function credentials(uint256 ) view returns(address requester, uint256 amount, uint256 requestTime, uint256 withdrawableTime, uint8 status)
func (_GovStaking *GovStakingCaller) Credentials(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Requester        common.Address
	Amount           *big.Int
	RequestTime      *big.Int
	WithdrawableTime *big.Int
	Status           uint8
}, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "credentials", arg0)

	outstruct := new(struct {
		Requester        common.Address
		Amount           *big.Int
		RequestTime      *big.Int
		WithdrawableTime *big.Int
		Status           uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Requester = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.RequestTime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.WithdrawableTime = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Status = *abi.ConvertType(out[4], new(uint8)).(*uint8)

	return *outstruct, err

}

// Credentials is a free data retrieval call binding the contract method 0xe0574e3f.
//
// Solidity: function credentials(uint256 ) view returns(address requester, uint256 amount, uint256 requestTime, uint256 withdrawableTime, uint8 status)
func (_GovStaking *GovStakingSession) Credentials(arg0 *big.Int) (struct {
	Requester        common.Address
	Amount           *big.Int
	RequestTime      *big.Int
	WithdrawableTime *big.Int
	Status           uint8
}, error) {
	return _GovStaking.Contract.Credentials(&_GovStaking.CallOpts, arg0)
}

// Credentials is a free data retrieval call binding the contract method 0xe0574e3f.
//
// Solidity: function credentials(uint256 ) view returns(address requester, uint256 amount, uint256 requestTime, uint256 withdrawableTime, uint8 status)
func (_GovStaking *GovStakingCallerSession) Credentials(arg0 *big.Int) (struct {
	Requester        common.Address
	Amount           *big.Int
	RequestTime      *big.Int
	WithdrawableTime *big.Int
	Status           uint8
}, error) {
	return _GovStaking.Contract.Credentials(&_GovStaking.CallOpts, arg0)
}

// DelegateTo is a free data retrieval call binding the contract method 0x438bb7e5.
//
// Solidity: function delegateTo(address , address ) view returns(uint256)
func (_GovStaking *GovStakingCaller) DelegateTo(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "delegateTo", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DelegateTo is a free data retrieval call binding the contract method 0x438bb7e5.
//
// Solidity: function delegateTo(address , address ) view returns(uint256)
func (_GovStaking *GovStakingSession) DelegateTo(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _GovStaking.Contract.DelegateTo(&_GovStaking.CallOpts, arg0, arg1)
}

// DelegateTo is a free data retrieval call binding the contract method 0x438bb7e5.
//
// Solidity: function delegateTo(address , address ) view returns(uint256)
func (_GovStaking *GovStakingCallerSession) DelegateTo(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _GovStaking.Contract.DelegateTo(&_GovStaking.CallOpts, arg0, arg1)
}

// IsStakerOrReward is a free data retrieval call binding the contract method 0xdb463214.
//
// Solidity: function isStakerOrReward(address _addr) view returns(bool)
func (_GovStaking *GovStakingCaller) IsStakerOrReward(opts *bind.CallOpts, _addr common.Address) (bool, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "isStakerOrReward", _addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsStakerOrReward is a free data retrieval call binding the contract method 0xdb463214.
//
// Solidity: function isStakerOrReward(address _addr) view returns(bool)
func (_GovStaking *GovStakingSession) IsStakerOrReward(_addr common.Address) (bool, error) {
	return _GovStaking.Contract.IsStakerOrReward(&_GovStaking.CallOpts, _addr)
}

// IsStakerOrReward is a free data retrieval call binding the contract method 0xdb463214.
//
// Solidity: function isStakerOrReward(address _addr) view returns(bool)
func (_GovStaking *GovStakingCallerSession) IsStakerOrReward(_addr common.Address) (bool, error) {
	return _GovStaking.Contract.IsStakerOrReward(&_GovStaking.CallOpts, _addr)
}

// IsValidator is a free data retrieval call binding the contract method 0xfacd743b.
//
// Solidity: function isValidator(address _validator) view returns(bool)
func (_GovStaking *GovStakingCaller) IsValidator(opts *bind.CallOpts, _validator common.Address) (bool, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "isValidator", _validator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidator is a free data retrieval call binding the contract method 0xfacd743b.
//
// Solidity: function isValidator(address _validator) view returns(bool)
func (_GovStaking *GovStakingSession) IsValidator(_validator common.Address) (bool, error) {
	return _GovStaking.Contract.IsValidator(&_GovStaking.CallOpts, _validator)
}

// IsValidator is a free data retrieval call binding the contract method 0xfacd743b.
//
// Solidity: function isValidator(address _validator) view returns(bool)
func (_GovStaking *GovStakingCallerSession) IsValidator(_validator common.Address) (bool, error) {
	return _GovStaking.Contract.IsValidator(&_GovStaking.CallOpts, _validator)
}

// TotalStaking is a free data retrieval call binding the contract method 0x165defa4.
//
// Solidity: function totalStaking() view returns(uint256)
func (_GovStaking *GovStakingCaller) TotalStaking(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "totalStaking")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalStaking is a free data retrieval call binding the contract method 0x165defa4.
//
// Solidity: function totalStaking() view returns(uint256)
func (_GovStaking *GovStakingSession) TotalStaking() (*big.Int, error) {
	return _GovStaking.Contract.TotalStaking(&_GovStaking.CallOpts)
}

// TotalStaking is a free data retrieval call binding the contract method 0x165defa4.
//
// Solidity: function totalStaking() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) TotalStaking() (*big.Int, error) {
	return _GovStaking.Contract.TotalStaking(&_GovStaking.CallOpts)
}

// ValidatorByReward is a free data retrieval call binding the contract method 0x9f7f7ed2.
//
// Solidity: function validatorByReward(address ) view returns(address)
func (_GovStaking *GovStakingCaller) ValidatorByReward(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "validatorByReward", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ValidatorByReward is a free data retrieval call binding the contract method 0x9f7f7ed2.
//
// Solidity: function validatorByReward(address ) view returns(address)
func (_GovStaking *GovStakingSession) ValidatorByReward(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.ValidatorByReward(&_GovStaking.CallOpts, arg0)
}

// ValidatorByReward is a free data retrieval call binding the contract method 0x9f7f7ed2.
//
// Solidity: function validatorByReward(address ) view returns(address)
func (_GovStaking *GovStakingCallerSession) ValidatorByReward(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.ValidatorByReward(&_GovStaking.CallOpts, arg0)
}

// ValidatorByStaker is a free data retrieval call binding the contract method 0xeff39866.
//
// Solidity: function validatorByStaker(address ) view returns(address)
func (_GovStaking *GovStakingCaller) ValidatorByStaker(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "validatorByStaker", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ValidatorByStaker is a free data retrieval call binding the contract method 0xeff39866.
//
// Solidity: function validatorByStaker(address ) view returns(address)
func (_GovStaking *GovStakingSession) ValidatorByStaker(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.ValidatorByStaker(&_GovStaking.CallOpts, arg0)
}

// ValidatorByStaker is a free data retrieval call binding the contract method 0xeff39866.
//
// Solidity: function validatorByStaker(address ) view returns(address)
func (_GovStaking *GovStakingCallerSession) ValidatorByStaker(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.ValidatorByStaker(&_GovStaking.CallOpts, arg0)
}

// ValidatorInfo is a free data retrieval call binding the contract method 0x4f1811dd.
//
// Solidity: function validatorInfo(address ) view returns(address staker, address reward, uint256 staking, uint256 delegated)
func (_GovStaking *GovStakingCaller) ValidatorInfo(opts *bind.CallOpts, arg0 common.Address) (struct {
	Staker    common.Address
	Reward    common.Address
	Staking   *big.Int
	Delegated *big.Int
}, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "validatorInfo", arg0)

	outstruct := new(struct {
		Staker    common.Address
		Reward    common.Address
		Staking   *big.Int
		Delegated *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Staker = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Reward = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Staking = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Delegated = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// ValidatorInfo is a free data retrieval call binding the contract method 0x4f1811dd.
//
// Solidity: function validatorInfo(address ) view returns(address staker, address reward, uint256 staking, uint256 delegated)
func (_GovStaking *GovStakingSession) ValidatorInfo(arg0 common.Address) (struct {
	Staker    common.Address
	Reward    common.Address
	Staking   *big.Int
	Delegated *big.Int
}, error) {
	return _GovStaking.Contract.ValidatorInfo(&_GovStaking.CallOpts, arg0)
}

// ValidatorInfo is a free data retrieval call binding the contract method 0x4f1811dd.
//
// Solidity: function validatorInfo(address ) view returns(address staker, address reward, uint256 staking, uint256 delegated)
func (_GovStaking *GovStakingCallerSession) ValidatorInfo(arg0 common.Address) (struct {
	Staker    common.Address
	Reward    common.Address
	Staking   *big.Int
	Delegated *big.Int
}, error) {
	return _GovStaking.Contract.ValidatorInfo(&_GovStaking.CallOpts, arg0)
}

// ValidatorLength is a free data retrieval call binding the contract method 0xaed1d403.
//
// Solidity: function validatorLength() view returns(uint256)
func (_GovStaking *GovStakingCaller) ValidatorLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "validatorLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ValidatorLength is a free data retrieval call binding the contract method 0xaed1d403.
//
// Solidity: function validatorLength() view returns(uint256)
func (_GovStaking *GovStakingSession) ValidatorLength() (*big.Int, error) {
	return _GovStaking.Contract.ValidatorLength(&_GovStaking.CallOpts)
}

// ValidatorLength is a free data retrieval call binding the contract method 0xaed1d403.
//
// Solidity: function validatorLength() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) ValidatorLength() (*big.Int, error) {
	return _GovStaking.Contract.ValidatorLength(&_GovStaking.CallOpts)
}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address[])
func (_GovStaking *GovStakingCaller) Validators(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "validators")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address[])
func (_GovStaking *GovStakingSession) Validators() ([]common.Address, error) {
	return _GovStaking.Contract.Validators(&_GovStaking.CallOpts)
}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address[])
func (_GovStaking *GovStakingCallerSession) Validators() ([]common.Address, error) {
	return _GovStaking.Contract.Validators(&_GovStaking.CallOpts)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address _validator, uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactor) Delegate(opts *bind.TransactOpts, _validator common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "delegate", _validator, _amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address _validator, uint256 _amount) payable returns()
func (_GovStaking *GovStakingSession) Delegate(_validator common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Delegate(&_GovStaking.TransactOpts, _validator, _amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address _validator, uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactorSession) Delegate(_validator common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Delegate(&_GovStaking.TransactOpts, _validator, _amount)
}

// RegisterValidator is a paid mutator transaction binding the contract method 0x0cd4b40f.
//
// Solidity: function registerValidator(uint256 _amount, address _validator, address _reward) payable returns()
func (_GovStaking *GovStakingTransactor) RegisterValidator(opts *bind.TransactOpts, _amount *big.Int, _validator common.Address, _reward common.Address) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "registerValidator", _amount, _validator, _reward)
}

// RegisterValidator is a paid mutator transaction binding the contract method 0x0cd4b40f.
//
// Solidity: function registerValidator(uint256 _amount, address _validator, address _reward) payable returns()
func (_GovStaking *GovStakingSession) RegisterValidator(_amount *big.Int, _validator common.Address, _reward common.Address) (*types.Transaction, error) {
	return _GovStaking.Contract.RegisterValidator(&_GovStaking.TransactOpts, _amount, _validator, _reward)
}

// RegisterValidator is a paid mutator transaction binding the contract method 0x0cd4b40f.
//
// Solidity: function registerValidator(uint256 _amount, address _validator, address _reward) payable returns()
func (_GovStaking *GovStakingTransactorSession) RegisterValidator(_amount *big.Int, _validator common.Address, _reward common.Address) (*types.Transaction, error) {
	return _GovStaking.Contract.RegisterValidator(&_GovStaking.TransactOpts, _amount, _validator, _reward)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactor) Stake(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "stake", _amount)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 _amount) payable returns()
func (_GovStaking *GovStakingSession) Stake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Stake(&_GovStaking.TransactOpts, _amount)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactorSession) Stake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Stake(&_GovStaking.TransactOpts, _amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4d99dd16.
//
// Solidity: function undelegate(address _validator, uint256 _amount) returns()
func (_GovStaking *GovStakingTransactor) Undelegate(opts *bind.TransactOpts, _validator common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "undelegate", _validator, _amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4d99dd16.
//
// Solidity: function undelegate(address _validator, uint256 _amount) returns()
func (_GovStaking *GovStakingSession) Undelegate(_validator common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Undelegate(&_GovStaking.TransactOpts, _validator, _amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4d99dd16.
//
// Solidity: function undelegate(address _validator, uint256 _amount) returns()
func (_GovStaking *GovStakingTransactorSession) Undelegate(_validator common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Undelegate(&_GovStaking.TransactOpts, _validator, _amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 _amount) returns()
func (_GovStaking *GovStakingTransactor) Unstake(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "unstake", _amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 _amount) returns()
func (_GovStaking *GovStakingSession) Unstake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Unstake(&_GovStaking.TransactOpts, _amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 _amount) returns()
func (_GovStaking *GovStakingTransactorSession) Unstake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Unstake(&_GovStaking.TransactOpts, _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _cid) returns()
func (_GovStaking *GovStakingTransactor) Withdraw(opts *bind.TransactOpts, _cid *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "withdraw", _cid)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _cid) returns()
func (_GovStaking *GovStakingSession) Withdraw(_cid *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Withdraw(&_GovStaking.TransactOpts, _cid)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _cid) returns()
func (_GovStaking *GovStakingTransactorSession) Withdraw(_cid *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Withdraw(&_GovStaking.TransactOpts, _cid)
}

// GovStakingDelegatedIterator is returned from FilterDelegated and is used to iterate over the raw logs and unpacked data for Delegated events raised by the GovStaking contract.
type GovStakingDelegatedIterator struct {
	Event *GovStakingDelegated // Event containing the contract specifics and raw log

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
func (it *GovStakingDelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingDelegated)
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
		it.Event = new(GovStakingDelegated)
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
func (it *GovStakingDelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingDelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingDelegated represents a Delegated event raised by the GovStaking contract.
type GovStakingDelegated struct {
	Delegator common.Address
	Validator common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegated is a free log retrieval operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterDelegated(opts *bind.FilterOpts, delegator []common.Address, validator []common.Address) (*GovStakingDelegatedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Delegated", delegatorRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingDelegatedIterator{contract: _GovStaking.contract, event: "Delegated", logs: logs, sub: sub}, nil
}

// WatchDelegated is a free log subscription operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchDelegated(opts *bind.WatchOpts, sink chan<- *GovStakingDelegated, delegator []common.Address, validator []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Delegated", delegatorRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingDelegated)
				if err := _GovStaking.contract.UnpackLog(event, "Delegated", log); err != nil {
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

// ParseDelegated is a log parse operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseDelegated(log types.Log) (*GovStakingDelegated, error) {
	event := new(GovStakingDelegated)
	if err := _GovStaking.contract.UnpackLog(event, "Delegated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingNewCredentialIterator is returned from FilterNewCredential and is used to iterate over the raw logs and unpacked data for NewCredential events raised by the GovStaking contract.
type GovStakingNewCredentialIterator struct {
	Event *GovStakingNewCredential // Event containing the contract specifics and raw log

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
func (it *GovStakingNewCredentialIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingNewCredential)
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
		it.Event = new(GovStakingNewCredential)
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
func (it *GovStakingNewCredentialIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingNewCredentialIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingNewCredential represents a NewCredential event raised by the GovStaking contract.
type GovStakingNewCredential struct {
	CredentialID *big.Int
	Requester    common.Address
	Amount       *big.Int
	Time         *big.Int
	Unbonding    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterNewCredential is a free log retrieval operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) FilterNewCredential(opts *bind.FilterOpts, credentialID []*big.Int, requester []common.Address) (*GovStakingNewCredentialIterator, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "NewCredential", credentialIDRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingNewCredentialIterator{contract: _GovStaking.contract, event: "NewCredential", logs: logs, sub: sub}, nil
}

// WatchNewCredential is a free log subscription operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) WatchNewCredential(opts *bind.WatchOpts, sink chan<- *GovStakingNewCredential, credentialID []*big.Int, requester []common.Address) (event.Subscription, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "NewCredential", credentialIDRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingNewCredential)
				if err := _GovStaking.contract.UnpackLog(event, "NewCredential", log); err != nil {
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

// ParseNewCredential is a log parse operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) ParseNewCredential(log types.Log) (*GovStakingNewCredential, error) {
	event := new(GovStakingNewCredential)
	if err := _GovStaking.contract.UnpackLog(event, "NewCredential", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the GovStaking contract.
type GovStakingStakedIterator struct {
	Event *GovStakingStaked // Event containing the contract specifics and raw log

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
func (it *GovStakingStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingStaked)
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
		it.Event = new(GovStakingStaked)
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
func (it *GovStakingStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingStaked represents a Staked event raised by the GovStaking contract.
type GovStakingStaked struct {
	Validator common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterStaked(opts *bind.FilterOpts, validator []common.Address) (*GovStakingStakedIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Staked", validatorRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingStakedIterator{contract: _GovStaking.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *GovStakingStaked, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Staked", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingStaked)
				if err := _GovStaking.contract.UnpackLog(event, "Staked", log); err != nil {
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

// ParseStaked is a log parse operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseStaked(log types.Log) (*GovStakingStaked, error) {
	event := new(GovStakingStaked)
	if err := _GovStaking.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingUndelegatedIterator is returned from FilterUndelegated and is used to iterate over the raw logs and unpacked data for Undelegated events raised by the GovStaking contract.
type GovStakingUndelegatedIterator struct {
	Event *GovStakingUndelegated // Event containing the contract specifics and raw log

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
func (it *GovStakingUndelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingUndelegated)
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
		it.Event = new(GovStakingUndelegated)
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
func (it *GovStakingUndelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingUndelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingUndelegated represents a Undelegated event raised by the GovStaking contract.
type GovStakingUndelegated struct {
	Delegator common.Address
	Validator common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUndelegated is a free log retrieval operation binding the contract event 0x4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c.
//
// Solidity: event Undelegated(address indexed delegator, address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterUndelegated(opts *bind.FilterOpts, delegator []common.Address, validator []common.Address) (*GovStakingUndelegatedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Undelegated", delegatorRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingUndelegatedIterator{contract: _GovStaking.contract, event: "Undelegated", logs: logs, sub: sub}, nil
}

// WatchUndelegated is a free log subscription operation binding the contract event 0x4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c.
//
// Solidity: event Undelegated(address indexed delegator, address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchUndelegated(opts *bind.WatchOpts, sink chan<- *GovStakingUndelegated, delegator []common.Address, validator []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Undelegated", delegatorRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingUndelegated)
				if err := _GovStaking.contract.UnpackLog(event, "Undelegated", log); err != nil {
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

// ParseUndelegated is a log parse operation binding the contract event 0x4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c.
//
// Solidity: event Undelegated(address indexed delegator, address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseUndelegated(log types.Log) (*GovStakingUndelegated, error) {
	event := new(GovStakingUndelegated)
	if err := _GovStaking.contract.UnpackLog(event, "Undelegated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingUnstakedIterator is returned from FilterUnstaked and is used to iterate over the raw logs and unpacked data for Unstaked events raised by the GovStaking contract.
type GovStakingUnstakedIterator struct {
	Event *GovStakingUnstaked // Event containing the contract specifics and raw log

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
func (it *GovStakingUnstakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingUnstaked)
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
		it.Event = new(GovStakingUnstaked)
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
func (it *GovStakingUnstakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingUnstakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingUnstaked represents a Unstaked event raised by the GovStaking contract.
type GovStakingUnstaked struct {
	Validator common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUnstaked is a free log retrieval operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterUnstaked(opts *bind.FilterOpts, validator []common.Address) (*GovStakingUnstakedIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Unstaked", validatorRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingUnstakedIterator{contract: _GovStaking.contract, event: "Unstaked", logs: logs, sub: sub}, nil
}

// WatchUnstaked is a free log subscription operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchUnstaked(opts *bind.WatchOpts, sink chan<- *GovStakingUnstaked, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Unstaked", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingUnstaked)
				if err := _GovStaking.contract.UnpackLog(event, "Unstaked", log); err != nil {
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

// ParseUnstaked is a log parse operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed validator, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseUnstaked(log types.Log) (*GovStakingUnstaked, error) {
	event := new(GovStakingUnstaked)
	if err := _GovStaking.contract.UnpackLog(event, "Unstaked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingValidatorRegisteredIterator is returned from FilterValidatorRegistered and is used to iterate over the raw logs and unpacked data for ValidatorRegistered events raised by the GovStaking contract.
type GovStakingValidatorRegisteredIterator struct {
	Event *GovStakingValidatorRegistered // Event containing the contract specifics and raw log

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
func (it *GovStakingValidatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingValidatorRegistered)
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
		it.Event = new(GovStakingValidatorRegistered)
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
func (it *GovStakingValidatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingValidatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingValidatorRegistered represents a ValidatorRegistered event raised by the GovStaking contract.
type GovStakingValidatorRegistered struct {
	Validator common.Address
	Staker    common.Address
	Reward    common.Address
	Staking   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterValidatorRegistered is a free log retrieval operation binding the contract event 0x08dc906197021397d84c73d31dcffd0296c7c044af3bb331a3091e122d0b58db.
//
// Solidity: event ValidatorRegistered(address indexed validator, address staker, address reward, uint256 staking)
func (_GovStaking *GovStakingFilterer) FilterValidatorRegistered(opts *bind.FilterOpts, validator []common.Address) (*GovStakingValidatorRegisteredIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "ValidatorRegistered", validatorRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingValidatorRegisteredIterator{contract: _GovStaking.contract, event: "ValidatorRegistered", logs: logs, sub: sub}, nil
}

// WatchValidatorRegistered is a free log subscription operation binding the contract event 0x08dc906197021397d84c73d31dcffd0296c7c044af3bb331a3091e122d0b58db.
//
// Solidity: event ValidatorRegistered(address indexed validator, address staker, address reward, uint256 staking)
func (_GovStaking *GovStakingFilterer) WatchValidatorRegistered(opts *bind.WatchOpts, sink chan<- *GovStakingValidatorRegistered, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "ValidatorRegistered", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingValidatorRegistered)
				if err := _GovStaking.contract.UnpackLog(event, "ValidatorRegistered", log); err != nil {
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

// ParseValidatorRegistered is a log parse operation binding the contract event 0x08dc906197021397d84c73d31dcffd0296c7c044af3bb331a3091e122d0b58db.
//
// Solidity: event ValidatorRegistered(address indexed validator, address staker, address reward, uint256 staking)
func (_GovStaking *GovStakingFilterer) ParseValidatorRegistered(log types.Log) (*GovStakingValidatorRegistered, error) {
	event := new(GovStakingValidatorRegistered)
	if err := _GovStaking.contract.UnpackLog(event, "ValidatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingValidatorRemovedIterator is returned from FilterValidatorRemoved and is used to iterate over the raw logs and unpacked data for ValidatorRemoved events raised by the GovStaking contract.
type GovStakingValidatorRemovedIterator struct {
	Event *GovStakingValidatorRemoved // Event containing the contract specifics and raw log

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
func (it *GovStakingValidatorRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingValidatorRemoved)
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
		it.Event = new(GovStakingValidatorRemoved)
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
func (it *GovStakingValidatorRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingValidatorRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingValidatorRemoved represents a ValidatorRemoved event raised by the GovStaking contract.
type GovStakingValidatorRemoved struct {
	Validator common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterValidatorRemoved is a free log retrieval operation binding the contract event 0xe1434e25d6611e0db941968fdc97811c982ac1602e951637d206f5fdda9dd8f1.
//
// Solidity: event ValidatorRemoved(address indexed validator)
func (_GovStaking *GovStakingFilterer) FilterValidatorRemoved(opts *bind.FilterOpts, validator []common.Address) (*GovStakingValidatorRemovedIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "ValidatorRemoved", validatorRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingValidatorRemovedIterator{contract: _GovStaking.contract, event: "ValidatorRemoved", logs: logs, sub: sub}, nil
}

// WatchValidatorRemoved is a free log subscription operation binding the contract event 0xe1434e25d6611e0db941968fdc97811c982ac1602e951637d206f5fdda9dd8f1.
//
// Solidity: event ValidatorRemoved(address indexed validator)
func (_GovStaking *GovStakingFilterer) WatchValidatorRemoved(opts *bind.WatchOpts, sink chan<- *GovStakingValidatorRemoved, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "ValidatorRemoved", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingValidatorRemoved)
				if err := _GovStaking.contract.UnpackLog(event, "ValidatorRemoved", log); err != nil {
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

// ParseValidatorRemoved is a log parse operation binding the contract event 0xe1434e25d6611e0db941968fdc97811c982ac1602e951637d206f5fdda9dd8f1.
//
// Solidity: event ValidatorRemoved(address indexed validator)
func (_GovStaking *GovStakingFilterer) ParseValidatorRemoved(log types.Log) (*GovStakingValidatorRemoved, error) {
	event := new(GovStakingValidatorRemoved)
	if err := _GovStaking.contract.UnpackLog(event, "ValidatorRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingWithdrewIterator is returned from FilterWithdrew and is used to iterate over the raw logs and unpacked data for Withdrew events raised by the GovStaking contract.
type GovStakingWithdrewIterator struct {
	Event *GovStakingWithdrew // Event containing the contract specifics and raw log

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
func (it *GovStakingWithdrewIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingWithdrew)
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
		it.Event = new(GovStakingWithdrew)
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
func (it *GovStakingWithdrewIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingWithdrewIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingWithdrew represents a Withdrew event raised by the GovStaking contract.
type GovStakingWithdrew struct {
	CredentialID *big.Int
	Requester    common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterWithdrew is a free log retrieval operation binding the contract event 0xf4c56bb5a568ebcb1087fa42400137fd62e67119180ddfe97906e108ecc0e633.
//
// Solidity: event Withdrew(uint256 indexed credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterWithdrew(opts *bind.FilterOpts, credentialID []*big.Int) (*GovStakingWithdrewIterator, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Withdrew", credentialIDRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingWithdrewIterator{contract: _GovStaking.contract, event: "Withdrew", logs: logs, sub: sub}, nil
}

// WatchWithdrew is a free log subscription operation binding the contract event 0xf4c56bb5a568ebcb1087fa42400137fd62e67119180ddfe97906e108ecc0e633.
//
// Solidity: event Withdrew(uint256 indexed credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchWithdrew(opts *bind.WatchOpts, sink chan<- *GovStakingWithdrew, credentialID []*big.Int) (event.Subscription, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Withdrew", credentialIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingWithdrew)
				if err := _GovStaking.contract.UnpackLog(event, "Withdrew", log); err != nil {
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

// ParseWithdrew is a log parse operation binding the contract event 0xf4c56bb5a568ebcb1087fa42400137fd62e67119180ddfe97906e108ecc0e633.
//
// Solidity: event Withdrew(uint256 indexed credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseWithdrew(log types.Log) (*GovStakingWithdrew, error) {
	event := new(GovStakingWithdrew)
	if err := _GovStaking.contract.UnpackLog(event, "Withdrew", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
