// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import {MultiSigWallet} from "./MultiSigWallet.sol";


contract OperatorSample is MultiSigWallet {
    address public constant GOV_STAKING = address(0x1001);
    bytes4 private constant SEND_FEE_SELECTOR = bytes4(keccak256("sendFee()"));
    bytes4 private constant SEND_REWARD_SELECTOR = bytes4(keccak256("sendReward()"));
    bytes4 private constant WITHDRAW_FEE_AMOUNT_SELECTOR = bytes4(keccak256("withdrawFeeAmount(address)"));
    bytes4 private constant WITHDRAW_REWARD_AMOUNT_SELECTOR = bytes4(keccak256("withdrawRewardAmount(address)"));
    bytes4 private constant WITHDRAW_UNSTAKED_AMOUNT_SELECTOR = bytes4(keccak256("withdrawUnstakedAmount(address)"));

    mapping(address => bool) private __claimers;
    uint256 public feeAmount;
    uint256 public unstakedAmount;
    uint256 public rewardAmount;

    uint8 private constant _NOT_ENTERED = 1;
    uint8 private constant _ENTERED = 2;
    uint8 private _status;


    event ReceivedUnstaked(address indexed from, uint256 amount);
    event ReceivedFee(address indexed from, uint256 amount);
    event ReceivedReward(address indexed from, uint256 amount);
    event SentUnstakedAmount(address indexed to, uint256 amount);
    event SentFeeAmount(address indexed to, uint256 amount);
    event SentRewardAmount(address indexed to, uint256 amount);


    // receive unstaked value by receive() function.
    // this includes value that eoa sends to operator contract for staking
    // and value from GovStaking contract that is unstaked.
    receive() external payable override {
        unstakedAmount += msg.value;
        emit ReceivedUnstaked(msg.sender, msg.value);
    }

    // receive fee or reward value by fallback() function
    fallback() external payable {
        require(msg.data.length >= 4, "Invalid call");
        if (bytes4(msg.data) == SEND_FEE_SELECTOR){
            feeAmount += msg.value;
            emit ReceivedFee(msg.sender, msg.value);
        } else if (bytes4(msg.data) == SEND_REWARD_SELECTOR) {
            rewardAmount += msg.value;
            emit ReceivedReward(msg.sender, msg.value);
        } else {
            revert("Invalid call");
        }
    }

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

    function submitTransaction(address _to, uint256 _value, bytes memory _data) public override onlyOwner {
        if (_to == address(this)) {
            bytes4 selector = bytes4(_data);
            bool isTransferFunction = isEtherTransferFunction(selector);

            if (isTransferFunction) {
                require(
                    selector == WITHDRAW_FEE_AMOUNT_SELECTOR ||
                    selector ==  WITHDRAW_REWARD_AMOUNT_SELECTOR ||
                    selector == WITHDRAW_UNSTAKED_AMOUNT_SELECTOR,
                    "Use proper withdraw functions to transfer value from contract"
                );
            }
        }
        super.submitTransaction(_to, _value, _data);
    }

    function isEtherTransferFunction(bytes4 _selector) internal pure returns (bool) {
        return (
            _selector == bytes4(keccak256("transfer(address,uint256)")) ||
            _selector == bytes4(keccak256("send(address,uint256)")) ||
            _selector == bytes4(keccak256("call(bytes)"))
        );
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

    // TODO : managing fee, distributing unstaked amount

    function _addClaimer(address _newClaimer) private {
        require(!__claimers[_newClaimer], "already registered claimer");
        __claimers[_newClaimer] = true;
    }

    function _removeClaimer(address _claimerToRemove) private {
        require(__claimers[_claimerToRemove], "claimer is not registered");
        __claimers[_claimerToRemove] = false;
    }

}