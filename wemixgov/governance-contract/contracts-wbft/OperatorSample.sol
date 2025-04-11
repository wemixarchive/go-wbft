// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import {MultiSigWallet} from "./MultiSigWallet.sol";
import {IFeeRecipient, IERC165} from "./IFeeRecipient.sol";


contract OperatorSample is MultiSigWallet, IFeeRecipient {
    address public constant GOV_STAKING = address(0x1001);
    bytes4 private constant _WITHDRAW_FEE_AMOUNT_SELECTOR = bytes4(keccak256("withdrawFeeAmount(address)"));
    bytes4 private constant _WITHDRAW_REWARD_AMOUNT_SELECTOR = bytes4(keccak256("withdrawRewardAmount(address)"));
    bytes4 private constant _WITHDRAW_UNSTAKED_AMOUNT_SELECTOR = bytes4(keccak256("withdrawUnstakedAmount(address)"));

    mapping(address => bool) private __claimers;
    uint256 public unstakedAmount;
    uint256 public rewardAmount;
    uint256 public feeAmount;

    uint8 private constant _NOT_ENTERED = 1;
    uint8 private constant _ENTERED = 2;
    uint8 private _status;

    bool private _receivingRewardStat;

    event ReceivedUnstaked(address indexed from, uint256 amount);
    event ReceivedReward(address indexed from, uint256 amount);
    event SentUnstakedAmount(address indexed to, uint256 amount);
    event SentRewardAmount(address indexed to, uint256 amount);


    modifier onlyClaimerOrOwner(address _addr) {
        require(__claimers[_addr] || isOwner(_addr), "only claimer or owner can execute");
        _;
    }

    modifier nonReentrant() {
        require(_status != _ENTERED, "ReentrancyGuard: reentrant call");
        _status = _ENTERED;
        _;
        _status = _NOT_ENTERED;
    }

    constructor(address[] memory _owners, uint256 _quorum) MultiSigWallet(_owners, _quorum) {
        _status = _NOT_ENTERED;
    }

    // receive() function is to receive reward, unstaked, or ether that eoa sent to the contract
    // when _receivingRewardStat is true, means the received ether is reward
    // else, increase the unstakedAmount
    receive() external payable override {
        if (_receivingRewardStat) {
            rewardAmount += msg.value;
            emit ReceivedReward(msg.sender, msg.value);
            _receivingRewardStat = false;
        }else {
            unstakedAmount += msg.value;
            emit ReceivedUnstaked(msg.sender, msg.value);
        }
    }

    /* ========== EXTERNAL FUNCTION ========== */

    // implements IERC165
    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return
            interfaceId == type(IFeeRecipient).interfaceId ||
            interfaceId == type(IERC165).interfaceId;
    }

    // implement IFeeRecipient
    function receiveFee(uint256 amount) external payable {
        require(amount == msg.value, "FeeRecipeint : FeeAmount and value sent mismatched");
        feeAmount += amount;
        emit ReceivedFee(msg.sender, amount);
    }

    function withdrawUnstakedAmount(address _to) external onlyWallet notNull(_to) nonReentrant {
        uint256 amount = unstakedAmount;
        unstakedAmount = 0;
        (bool success, ) = payable(_to).call{value:amount}("");
        require(success, "failed to send unstaked amount");
        emit SentUnstakedAmount(_to, amount);
    }

    function withdrawRewardAmount(address _to) external onlyWallet notNull(_to) nonReentrant {
        uint256 amount = rewardAmount;
        rewardAmount = 0;
        (bool success, ) = payable(_to).call{value:amount}("");
        require(success, "failed to send reward amount");
        emit SentRewardAmount(_to, amount);
    }

    // implement IFeeRecipient
    function withdrawFeeAmount(address _to) external onlyWallet notNull(_to) nonReentrant {
        uint256 amount = feeAmount;
        feeAmount = 0;
        (bool success, ) = payable(_to).call{value:amount}("");
        require(success, "failed to send fee amount");
        emit SentFeeAmount(_to, amount);
    }

    // call registerStaker without multiSig confirmation
    function registerStaker(uint256 _amount, address _staker, address _feeRecipient, uint256 _feeRate, bytes calldata _blsPK) external onlyOwner isSingleOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature(
            "registerStaker(uint256,address,address,uint256,bytes)",
            _amount,
            _staker,
            _feeRecipient,
            _feeRate,
            _blsPK
        );
        (bool success, ) = GOV_STAKING.call{value: _amount}(data);
        require(success, "registerStaker tx failed");
    }

    // call stake without multiSig confirmation
    function stake(uint256 _amount) external onlyOwner isSingleOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature(
            "stake(uint256)",
            _amount
        );
        (bool success, ) = GOV_STAKING.call{value: _amount}(data);
        require(success, "stake tx failed");
    }

    // call unstake without multiSig confirmation
    function unstake(uint256 _amount) external onlyOwner isSingleOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature(
            "unstake(uint256)",
            _amount
        );
        (bool success, ) = GOV_STAKING.call(data);
        require(success, "unstake tx failed");
    }

    //claim
    function claim(address _staker, bool _restake) external onlyClaimerOrOwner(msg.sender) {
        bytes memory data = abi.encodeWithSignature(
            "claim(address, bool)",
            _staker,
            _restake
        );
        (bool success, ) = GOV_STAKING.call(data);
        require(success, "claim tx failed");
    }


    /* ========== PUBLIC FUNCTION ========== */

    function submitTransaction(address _to, uint256 _value, bytes memory _data) public override onlyOwner {
        if (_to == address(this)) {
            bytes4 selector = bytes4(_data);
            bool isTransferFunction = _isEtherTransferFunction(selector);

            if (isTransferFunction) {
                require(
                    selector == _WITHDRAW_FEE_AMOUNT_SELECTOR ||
                    selector ==  _WITHDRAW_REWARD_AMOUNT_SELECTOR ||
                    selector == _WITHDRAW_UNSTAKED_AMOUNT_SELECTOR,
                    "Use proper withdraw functions to transfer value from contract"
                );
            }
        }
        super.submitTransaction(_to, _value, _data);
    }


    function executeTransaction(uint256 _transactionId) public payable override onlyOwner isTransactionExist(_transactionId) notExecuted(_transactionId) {
        Transaction storage transaction = transactions[_transactionId];
        require(transaction.currentNumberOfConfirmations >= quorum, "MultiSig: Current Number Of Confirmations must be greater than or equal to quorum.");

        if (transaction.to == address(GOV_STAKING)) {
            (bool isClaimCall, address staker, bool restake) = _getClaimParameters(transaction);
            if (isClaimCall && !restake) {
                _receivingRewardStat = true;
                transaction.executed = true;
            }
        }
        transaction.executed = true;
        (bool success, ) = transaction.to.call{value: transaction.value}(transaction.data);
        if (!success && _receivingRewardStat) {
            _receivingRewardStat = false;
        }
        require(success, "MultiSig: Transaction failed.");

        emit ExecuteTransaction(msg.sender, _transactionId);
    }


    /* ========== INTERNAL FUNCTION ========== */

    function _isEtherTransferFunction(bytes4 _selector) internal pure returns (bool) {
        return (
            _selector == bytes4(keccak256("transfer(address,uint256)")) ||
            _selector == bytes4(keccak256("send(address,uint256)")) ||
            _selector == bytes4(keccak256("call(bytes)"))
        );
    }

    function _getClaimParameters(Transaction storage transaction) internal  view returns (bool isClaimCall, address staker, bool restake) {
        if (transaction.data.length < 68) return (false, address(0), false);

        bytes memory txData = transaction.data;

        bytes4 signature;
        assembly {
            signature := mload(add(txData, 32))
        }

        bytes4 claimSelector = bytes4(keccak256("claim(address,bool)"));
        if (signature != claimSelector) return (false, address(0), false);

        assembly {
            staker := mload(add(txData, 64))
            restake := mload(add(txData, 96))
        }

        return (true, staker, restake);
    }

    /* ========== PRIVATE FUNCTION ========== */

    function _addClaimer(address _newClaimer) private {
        require(!__claimers[_newClaimer], "already registered claimer");
        __claimers[_newClaimer] = true;
    }

    function _removeClaimer(address _claimerToRemove) private {
        require(__claimers[_claimerToRemove], "claimer is not registered");
        __claimers[_claimerToRemove] = false;
    }

}