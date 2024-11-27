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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Delegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unbonding\",\"type\":\"uint256\"}],\"name\":\"NewCredential\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"reward\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"}],\"name\":\"NewValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Undelegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"}],\"name\":\"ValidatorRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"credentialID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrew\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAXIMUM_STAKING\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINIMUM_STAKING\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UNBONDING_PERIOD_USER\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UNBONDING_PERIOD_VALIDATOR\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"credentials\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"requestTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"withdrawableTime\",\"type\":\"uint256\"},{\"internalType\":\"enumGovStaking.WithdrawalStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"delegateTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"isStakerOrReward\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"isValidator\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_reward\",\"type\":\"address\"}],\"name\":\"newValidator\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalStaking\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_validator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"undelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validatorByReward\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validatorByStaker\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validatorInfo\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"reward\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"delegated\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validatorLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_cid\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"129060ab": "MAXIMUM_STAKING()",
		"ba631d3f": "MINIMUM_STAKING()",
		"96b403bb": "UNBONDING_PERIOD_USER()",
		"f90aa6ca": "UNBONDING_PERIOD_VALIDATOR()",
		"e0574e3f": "credentials(uint256)",
		"026e402b": "delegate(address,uint256)",
		"438bb7e5": "delegateTo(address,address)",
		"db463214": "isStakerOrReward(address)",
		"facd743b": "isValidator(address)",
		"fcb16116": "newValidator(uint256,address,address)",
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
	Bin: "0x608060405234801561001057600080fd5b506118e1806100206000396000f3fe60806040526004361061011f5760003560e01c8063a694fc3a116100a0578063e0574e3f11610064578063e0574e3f1461038e578063eff39866146103bf578063f90aa6ca146103f5578063facd743b1461040b578063fcb161161461042b57600080fd5b8063a694fc3a146102f6578063aed1d40314610309578063ba631d3f1461031e578063ca1e78191461033c578063db4632141461035e57600080fd5b8063438bb7e5116100e7578063438bb7e5146101bd5780634d99dd16146101f55780634f1811dd1461021557806396b403bb146102915780639f7f7ed2146102a857600080fd5b8063026e402b14610124578063129060ab14610139578063165defa4146101675780632e17de781461017d5780632e1a7d4d1461019d575b600080fd5b610137610132366004611650565b61043e565b005b34801561014557600080fd5b506101546001600160801b0381565b6040519081526020015b60405180910390f35b34801561017357600080fd5b5061015460005481565b34801561018957600080fd5b5061013761019836600461167a565b61067a565b3480156101a957600080fd5b506101376101b836600461167a565b610916565b3480156101c957600080fd5b506101546101d8366004611693565b600660209081526000928352604080842090915290825290205481565b34801561020157600080fd5b50610137610210366004611650565b610abe565b34801561022157600080fd5b506102666102303660046116c6565b600360208190526000918252604090912080546001820154600283015492909301546001600160a01b0391821693909116919084565b604080516001600160a01b03958616815294909316602085015291830152606082015260800161015e565b34801561029d57600080fd5b506101546203f48081565b3480156102b457600080fd5b506102de6102c33660046116c6565b6005602052600090815260409020546001600160a01b031681565b6040516001600160a01b03909116815260200161015e565b61013761030436600461167a565b610c01565b34801561031557600080fd5b50610154610d39565b34801561032a57600080fd5b506101546969e10de76676d080000081565b34801561034857600080fd5b50610351610d4a565b60405161015e91906116e1565b34801561036a57600080fd5b5061037e6103793660046116c6565b610d56565b604051901515815260200161015e565b34801561039a57600080fd5b506103ae6103a936600461167a565b610d9f565b60405161015e959493929190611744565b3480156103cb57600080fd5b506102de6103da3660046116c6565b6004602052600090815260409020546001600160a01b031681565b34801561040157600080fd5b50610154610e1081565b34801561041757600080fd5b5061037e6104263660046116c6565b610ded565b610137610439366004611797565b610dfa565b808034146104675760405162461bcd60e51b815260040161045e906117d3565b60405180910390fd5b61047083610ded565b61048c5760405162461bcd60e51b815260040161045e9061180a565b61049533610ded565b156104e25760405162461bcd60e51b815260206004820152601960248201527f76616c696461746f722063616e6e6f742064656c656761746500000000000000604482015260640161045e565b6104eb33610d56565b156105385760405162461bcd60e51b815260206004820152601e60248201527f7374616b657228726577617264292063616e6e6f742064656c65676174650000604482015260640161045e565b6001600160a01b038316600090815260036020526040902060028101546001600160801b0390610569908590611850565b11156105ae5760405162461bcd60e51b8152602060048201526014602482015273657863656564656420746865206d6178696d756d60601b604482015260640161045e565b3360009081526006602090815260408083206001600160a01b0388168452909152812080548592906105e1908490611850565b92505081905550828160020160008282546105fc9190611850565b92505081905550828160030160008282546106179190611850565b925050819055508260008082825461062f9190611850565b90915550506040518381526001600160a01b0385169033907fe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b9060200160405180910390a350505050565b336000908152600460205260409020546001600160a01b0316806106b05760405162461bcd60e51b815260040161045e9061180a565b600082116106f15760405162461bcd60e51b815260206004820152600e60248201526d616d6f756e74206973207a65726f60901b604482015260640161045e565b6001600160a01b038116600090815260036020819052604082209081015460028201549192916107219190611868565b90508381101561076a5760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b604482015260640161045e565b6969e10de76676d080000061077f8583611868565b1015610890578381146107ea5760405162461bcd60e51b815260206004820152602d60248201527f616d6f756e74206d75737420657175616c2062616c616e636520746f2072656d60448201526c37bb32903b30b634b230ba37b960991b606482015260840161045e565b6107f56001846111f8565b5033600090815260046020908152604080832080546001600160a01b03199081169091556001868101546001600160a01b0390811686526005855283862080548416905588168086526003948590528386208054841681559182018054909316909255600281018590559092018390555190917fe1434e25d6611e0db941968fdc97811c982ac1602e951637d206f5fdda9dd8f191a26108aa565b838260020160008282546108a49190611868565b90915550505b836000808282546108bb9190611868565b909155506108cd905084610e10611214565b826001600160a01b03167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f758560405161090891815260200190565b60405180910390a250505050565b60006007828154811061092b5761092b61187f565b6000918252602090912060059091020190506001600482015460ff1660028111156109585761095861172e565b1461099a5760405162461bcd60e51b81526020600482015260126024820152711a5b9d985b1a590818dc9959195b9d1a585b60721b604482015260640161045e565b8054336001600160a01b03909116036109f55760405162461bcd60e51b815260206004820152601b60248201527f6d73672e73656e646572206973206e6f74207265717565737465720000000000604482015260640161045e565b8060030154421015610a495760405162461bcd60e51b815260206004820152601860248201527f6e6f74207965742074696d6520746f2077697468647261770000000000000000604482015260640161045e565b60018101548154610a65916001600160a01b039091169061132a565b60048101805460ff19166002179055600181015460408051848152336020820152908101919091527ff4c56bb5a568ebcb1087fa42400137fd62e67119180ddfe97906e108ecc0e6339060600160405180910390a15050565b3360009081526006602090815260408083206001600160a01b0386168452909152902054811115610b285760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b604482015260640161045e565b3360009081526006602090815260408083206001600160a01b038616845290915281208054839290610b5b908490611868565b90915550610b6a905082610ded565b15610bb2576001600160a01b03821660009081526003602081905260408220018054839290610b9a908490611868565b90915550610bad9050816203f480611214565b610bbc565b610bbc338261132a565b6040518181526001600160a01b0383169033907f4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c906020015b60405180910390a35050565b80803414610c215760405162461bcd60e51b815260040161045e906117d3565b336000908152600460205260409020546001600160a01b031680610c575760405162461bcd60e51b815260040161045e9061180a565b6001600160a01b038116600090815260036020526040902060028101546001600160801b0390610c88908690611850565b1115610ccd5760405162461bcd60e51b8152602060048201526014602482015273657863656564656420746865206d6178696d756d60601b604482015260640161045e565b83816002016000828254610ce19190611850565b9250508190555083600080828254610cf99190611850565b90915550506040518481526001600160a01b038316907f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d90602001610908565b6000610d456001611448565b905090565b6060610d456001611452565b6001600160a01b03818116600090815260046020526040812054909116151580610d9957506001600160a01b038281166000908152600560205260409020541615155b92915050565b60078181548110610daf57600080fd5b6000918252602090912060059091020180546001820154600283015460038401546004909401546001600160a01b0390931694509092909160ff1685565b6000610d9960018361145f565b82803414610e1a5760405162461bcd60e51b815260040161045e906117d3565b6969e10de76676d08000008410158015610e3b57506001600160801b038411155b610e775760405162461bcd60e51b815260206004820152600d60248201526c6f7574206f6620626f756e647360981b604482015260640161045e565b336001600160a01b03841614801590610e995750336001600160a01b03831614155b610ef15760405162461bcd60e51b8152602060048201526024808201527f7374616b65722063616e6e6f742062652076616c696461746f72206f722072656044820152631dd85c9960e21b606482015260840161045e565b6001600160a01b03831615801590610f1157506001600160a01b03821615155b610f4c5760405162461bcd60e51b815260206004820152600c60248201526b7a65726f206164647265737360a01b604482015260640161045e565b816001600160a01b0316836001600160a01b031603610fad5760405162461bcd60e51b815260206004820152601a60248201527f76616c696461746f722063616e6e6f7420626520726577617264000000000000604482015260640161045e565b610fb633610d56565b156110035760405162461bcd60e51b815260206004820152601c60248201527f7374616b657220697320616c7265616479207265676973746572656400000000604482015260640161045e565b61100c83610d56565b156110595760405162461bcd60e51b815260206004820152601f60248201527f76616c696461746f7220697320616c7265616479207265676973746572656400604482015260640161045e565b61106282610d56565b156110af5760405162461bcd60e51b815260206004820152601c60248201527f72657761726420697320616c7265616479207265676973746572656400000000604482015260640161045e565b6110ba600184611481565b6110f95760405162461bcd60e51b815260206004820152601060248201526f76616c696461746f722065786973747360801b604482015260640161045e565b60408051608081018252338082526001600160a01b0385811660208085018281528587018b81526000606088018181528c871680835260038087528b84209a518b54908a166001600160a01b0319918216178c55955160018c01805491909a1690871617909855925160028a01555197909501969096559383526004815285832080548516861790559082526005905292832080549091169091179055805485919081906111a8908490611850565b9091555050604080513381526001600160a01b038481166020830152918101869052908416907f44ed68c3423517a7680fbb45340f2c9c788d35002529c153856aa1611a9c655c90606001610908565b600061120d836001600160a01b038416611496565b9392505050565b60076040518060a00160405280336001600160a01b0316815260200184815260200142815260200183426112489190611850565b8152602001600190528154600180820184556000938452602093849020835160059093020180546001600160a01b0319166001600160a01b0390931692909217825592820151818401556040820151600280830191909155606083015160038301556080830151600483018054949593949193909260ff199092169184908111156112d5576112d561172e565b0217905550506007543391506112ed90600190611868565b604080518581524260208201529081018490527f4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d90606001610bf5565b8047101561137a5760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a20696e73756666696369656e742062616c616e6365000000604482015260640161045e565b6000826001600160a01b03168260405160006040518083038185875af1925050503d80600081146113c7576040519150601f19603f3d011682016040523d82523d6000602084013e6113cc565b606091505b50509050806114435760405162461bcd60e51b815260206004820152603a60248201527f416464726573733a20756e61626c6520746f2073656e642076616c75652c207260448201527f6563697069656e74206d61792068617665207265766572746564000000000000606482015260840161045e565b505050565b6000610d99825490565b6060600061120d83611589565b6001600160a01b0381166000908152600183016020526040812054151561120d565b600061120d836001600160a01b0384166115e5565b6000818152600183016020526040812054801561157f5760006114ba600183611868565b85549091506000906114ce90600190611868565b90508181146115335760008660000182815481106114ee576114ee61187f565b90600052602060002001549050808760000184815481106115115761151161187f565b6000918252602080832090910192909255918252600188019052604090208390555b855486908061154457611544611895565b600190038181906000526020600020016000905590558560010160008681526020019081526020016000206000905560019350505050610d99565b6000915050610d99565b6060816000018054806020026020016040519081016040528092919081815260200182805480156115d957602002820191906000526020600020905b8154815260200190600101908083116115c5575b50505050509050919050565b600081815260018301602052604081205461162c57508154600181810184556000848152602080822090930184905584548482528286019093526040902091909155610d99565b506000610d99565b80356001600160a01b038116811461164b57600080fd5b919050565b6000806040838503121561166357600080fd5b61166c83611634565b946020939093013593505050565b60006020828403121561168c57600080fd5b5035919050565b600080604083850312156116a657600080fd5b6116af83611634565b91506116bd60208401611634565b90509250929050565b6000602082840312156116d857600080fd5b61120d82611634565b6020808252825182820181905260009190848201906040850190845b818110156117225783516001600160a01b0316835292840192918401916001016116fd565b50909695505050505050565b634e487b7160e01b600052602160045260246000fd5b6001600160a01b038616815260208101859052604081018490526060810183905260a081016003831061178757634e487b7160e01b600052602160045260246000fd5b8260808301529695505050505050565b6000806000606084860312156117ac57600080fd5b833592506117bc60208501611634565b91506117ca60408501611634565b90509250925092565b6020808252601d908201527f616d6f756e7420616e64206d73672e76616c7565206d69736d61746368000000604082015260600190565b6020808252601690820152753ab73932b3b4b9ba32b932b2103b30b634b230ba37b960511b604082015260600190565b634e487b7160e01b600052601160045260246000fd5b600082198211156118635761186361183a565b500190565b60008282101561187a5761187a61183a565b500390565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fdfea26469706673582212202d93f0477d608444d720d7c8b49d018c7c04c04bfe5c956026f5bbdd1671c47964736f6c634300080e0033",
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

// UNBONDINGPERIODUSER is a free data retrieval call binding the contract method 0x96b403bb.
//
// Solidity: function UNBONDING_PERIOD_USER() view returns(uint256)
func (_GovStaking *GovStakingCaller) UNBONDINGPERIODUSER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "UNBONDING_PERIOD_USER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UNBONDINGPERIODUSER is a free data retrieval call binding the contract method 0x96b403bb.
//
// Solidity: function UNBONDING_PERIOD_USER() view returns(uint256)
func (_GovStaking *GovStakingSession) UNBONDINGPERIODUSER() (*big.Int, error) {
	return _GovStaking.Contract.UNBONDINGPERIODUSER(&_GovStaking.CallOpts)
}

// UNBONDINGPERIODUSER is a free data retrieval call binding the contract method 0x96b403bb.
//
// Solidity: function UNBONDING_PERIOD_USER() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) UNBONDINGPERIODUSER() (*big.Int, error) {
	return _GovStaking.Contract.UNBONDINGPERIODUSER(&_GovStaking.CallOpts)
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

// NewValidator is a paid mutator transaction binding the contract method 0xfcb16116.
//
// Solidity: function newValidator(uint256 _amount, address _validator, address _reward) payable returns()
func (_GovStaking *GovStakingTransactor) NewValidator(opts *bind.TransactOpts, _amount *big.Int, _validator common.Address, _reward common.Address) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "newValidator", _amount, _validator, _reward)
}

// NewValidator is a paid mutator transaction binding the contract method 0xfcb16116.
//
// Solidity: function newValidator(uint256 _amount, address _validator, address _reward) payable returns()
func (_GovStaking *GovStakingSession) NewValidator(_amount *big.Int, _validator common.Address, _reward common.Address) (*types.Transaction, error) {
	return _GovStaking.Contract.NewValidator(&_GovStaking.TransactOpts, _amount, _validator, _reward)
}

// NewValidator is a paid mutator transaction binding the contract method 0xfcb16116.
//
// Solidity: function newValidator(uint256 _amount, address _validator, address _reward) payable returns()
func (_GovStaking *GovStakingTransactorSession) NewValidator(_amount *big.Int, _validator common.Address, _reward common.Address) (*types.Transaction, error) {
	return _GovStaking.Contract.NewValidator(&_GovStaking.TransactOpts, _amount, _validator, _reward)
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
	Id        *big.Int
	Requester common.Address
	Amount    *big.Int
	Time      *big.Int
	Unbonding *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNewCredential is a free log retrieval operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed id, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) FilterNewCredential(opts *bind.FilterOpts, id []*big.Int, requester []common.Address) (*GovStakingNewCredentialIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "NewCredential", idRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingNewCredentialIterator{contract: _GovStaking.contract, event: "NewCredential", logs: logs, sub: sub}, nil
}

// WatchNewCredential is a free log subscription operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed id, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) WatchNewCredential(opts *bind.WatchOpts, sink chan<- *GovStakingNewCredential, id []*big.Int, requester []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "NewCredential", idRule, requesterRule)
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
// Solidity: event NewCredential(uint256 indexed id, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) ParseNewCredential(log types.Log) (*GovStakingNewCredential, error) {
	event := new(GovStakingNewCredential)
	if err := _GovStaking.contract.UnpackLog(event, "NewCredential", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingNewValidatorIterator is returned from FilterNewValidator and is used to iterate over the raw logs and unpacked data for NewValidator events raised by the GovStaking contract.
type GovStakingNewValidatorIterator struct {
	Event *GovStakingNewValidator // Event containing the contract specifics and raw log

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
func (it *GovStakingNewValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingNewValidator)
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
		it.Event = new(GovStakingNewValidator)
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
func (it *GovStakingNewValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingNewValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingNewValidator represents a NewValidator event raised by the GovStaking contract.
type GovStakingNewValidator struct {
	Validator common.Address
	Staker    common.Address
	Reward    common.Address
	Staking   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNewValidator is a free log retrieval operation binding the contract event 0x44ed68c3423517a7680fbb45340f2c9c788d35002529c153856aa1611a9c655c.
//
// Solidity: event NewValidator(address indexed validator, address staker, address reward, uint256 staking)
func (_GovStaking *GovStakingFilterer) FilterNewValidator(opts *bind.FilterOpts, validator []common.Address) (*GovStakingNewValidatorIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "NewValidator", validatorRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingNewValidatorIterator{contract: _GovStaking.contract, event: "NewValidator", logs: logs, sub: sub}, nil
}

// WatchNewValidator is a free log subscription operation binding the contract event 0x44ed68c3423517a7680fbb45340f2c9c788d35002529c153856aa1611a9c655c.
//
// Solidity: event NewValidator(address indexed validator, address staker, address reward, uint256 staking)
func (_GovStaking *GovStakingFilterer) WatchNewValidator(opts *bind.WatchOpts, sink chan<- *GovStakingNewValidator, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "NewValidator", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingNewValidator)
				if err := _GovStaking.contract.UnpackLog(event, "NewValidator", log); err != nil {
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

// ParseNewValidator is a log parse operation binding the contract event 0x44ed68c3423517a7680fbb45340f2c9c788d35002529c153856aa1611a9c655c.
//
// Solidity: event NewValidator(address indexed validator, address staker, address reward, uint256 staking)
func (_GovStaking *GovStakingFilterer) ParseNewValidator(log types.Log) (*GovStakingNewValidator, error) {
	event := new(GovStakingNewValidator)
	if err := _GovStaking.contract.UnpackLog(event, "NewValidator", log); err != nil {
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
// Solidity: event Withdrew(uint256 credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterWithdrew(opts *bind.FilterOpts) (*GovStakingWithdrewIterator, error) {

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Withdrew")
	if err != nil {
		return nil, err
	}
	return &GovStakingWithdrewIterator{contract: _GovStaking.contract, event: "Withdrew", logs: logs, sub: sub}, nil
}

// WatchWithdrew is a free log subscription operation binding the contract event 0xf4c56bb5a568ebcb1087fa42400137fd62e67119180ddfe97906e108ecc0e633.
//
// Solidity: event Withdrew(uint256 credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchWithdrew(opts *bind.WatchOpts, sink chan<- *GovStakingWithdrew) (event.Subscription, error) {

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Withdrew")
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
// Solidity: event Withdrew(uint256 credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseWithdrew(log types.Log) (*GovStakingWithdrew, error) {
	event := new(GovStakingWithdrew)
	if err := _GovStaking.contract.UnpackLog(event, "Withdrew", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
