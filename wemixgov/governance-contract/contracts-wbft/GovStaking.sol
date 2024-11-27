// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import "@openzeppelin/contracts/utils/Address.sol";

contract GovStaking {
    using EnumerableSet for EnumerableSet.AddressSet;
    using Address for address payable;

    struct Validator {
        address staker;
        address reward;
        uint256 staking;
        uint256 delegated;
    }

    struct WithdrawalCredential {
        address requester;
        uint256 amount;
        uint256 requestTime;
        uint256 withdrawableTime;
        WithdrawalStatus status;
    }

    enum WithdrawalStatus {
        None,
        Requested,
        Withdrew
    }
    event NewValidator(address indexed validator, address staker, address reward, uint256 staking);
    event Staked(address indexed validator, uint256 amount);
    event Unstaked(address indexed validator, uint256 amount);
    event ValidatorRemoved(address indexed validator);
    event Delegated(address indexed delegator, address indexed validator, uint256 amount);
    event Undelegated(address indexed delegator, address indexed validator, uint256 amount);
    event NewCredential(uint256 indexed id, address indexed requester, uint256 amount, uint256 time, uint256 unbonding);
    event Withdrew(uint256 credentialID, address requester, uint256 amount);

    uint256 public constant MINIMUM_STAKING = 500000e18;
    uint256 public constant MAXIMUM_STAKING = type(uint128).max;
    uint256 public constant UNBONDING_PERIOD_VALIDATOR = 1 hours;
    uint256 public constant UNBONDING_PERIOD_USER = 72 hours;

    uint256 public totalStaking; // 0x0

    // Validator
    EnumerableSet.AddressSet private __validatorSet; // 0x1, 0x2
    mapping(address => Validator) public validatorInfo; // 0x3
    mapping(address => address) public validatorByStaker; // 0x4
    mapping(address => address) public validatorByReward; // 0x5

    // Delegate
    mapping(address => mapping(address => uint256)) public delegateTo; // 0x5

    // Withdrawal Credential
    WithdrawalCredential[] public credentials; // 0x6

    modifier checkAmount(uint256 _amount) {
        require(msg.value == _amount, "amount and msg.value mismatch");
        _;
    }

    function isValidator(address _validator) public view returns (bool) {
        return __validatorSet.contains(_validator);
    }

    function isStakerOrReward(address _addr) public view returns (bool) {
        return validatorByStaker[_addr] != address(0) || validatorByReward[_addr] != address(0);
    }

    function validatorLength() external view returns (uint256) {
        return __validatorSet.length();
    }

    function validators() external view returns (address[] memory) {
        return __validatorSet.values();
    }

    function newValidator(uint256 _amount, address _validator, address _reward) external payable checkAmount(_amount) {
        require(_amount >= MINIMUM_STAKING && _amount <= MAXIMUM_STAKING, "out of bounds");
        require(msg.sender != _validator && msg.sender != _reward, "staker cannot be validator or reward");
        require(_validator != address(0) && _reward != address(0), "zero address");
        require(_validator != _reward, "validator cannot be reward");
        require(!isStakerOrReward(msg.sender), "staker is already registered");
        require(!isStakerOrReward(_validator), "validator is already registered");
        require(!isStakerOrReward(_reward), "reward is already registered");

        require(__validatorSet.add(_validator), "validator exists");
        validatorInfo[_validator] = Validator({ staker: msg.sender, reward: _reward, staking: _amount, delegated: 0 });

        validatorByStaker[msg.sender] = _validator;
        validatorByReward[_reward] = _validator;

        totalStaking += _amount;

        emit NewValidator(_validator, msg.sender, _reward, _amount);
    }

    function stake(uint256 _amount) external payable checkAmount(_amount) {
        address _validator = validatorByStaker[msg.sender];
        require(_validator != address(0), "unregistered validator");

        Validator storage _validatorInfo = validatorInfo[_validator];
        require(_validatorInfo.staking + _amount <= MAXIMUM_STAKING, "exceeded the maximum");

        _validatorInfo.staking += _amount;
        totalStaking += _amount;

        emit Staked(_validator, _amount);
    }

    function delegate(address _validator, uint256 _amount) external payable checkAmount(_amount) {
        require(isValidator(_validator), "unregistered validator");
        require(!isValidator(msg.sender), "validator cannot delegate");
        require(!isStakerOrReward(msg.sender), "staker(reward) cannot delegate");

        Validator storage _validatorInfo = validatorInfo[_validator];
        require(_validatorInfo.staking + _amount <= MAXIMUM_STAKING, "exceeded the maximum");

        delegateTo[msg.sender][_validator] += _amount;
        _validatorInfo.staking += _amount;
        _validatorInfo.delegated += _amount;

        totalStaking += _amount;

        emit Delegated(msg.sender, _validator, _amount);
    }

    function unstake(uint256 _amount) external {
        address _validator = validatorByStaker[msg.sender];
        require(_validator != address(0), "unregistered validator");
        require(_amount > 0, "amount is zero");

        Validator storage _validatorInfo = validatorInfo[_validator];
        uint256 _validatorStaking = _validatorInfo.staking - _validatorInfo.delegated;

        require(_validatorStaking >= _amount, "insufficient balance");
        if (_validatorStaking - _amount < MINIMUM_STAKING) {
            require(_validatorStaking == _amount, "amount must equal balance to remove validator");

            __validatorSet.remove(_validator);
            delete validatorByStaker[msg.sender];
            delete validatorByReward[_validatorInfo.reward];
            delete validatorInfo[_validator];

            emit ValidatorRemoved(_validator);
        } else {
            _validatorInfo.staking -= _amount;
        }

        totalStaking -= _amount;
        _newCredential(_amount, UNBONDING_PERIOD_VALIDATOR);

        emit Unstaked(_validator, _amount);
    }

    function undelegate(address _validator, uint256 _amount) external {
        require(delegateTo[msg.sender][_validator] >= _amount, "insufficient balance");

        delegateTo[msg.sender][_validator] -= _amount;
        if (isValidator(_validator)) {
            validatorInfo[_validator].delegated -= _amount;
            _newCredential(_amount, UNBONDING_PERIOD_USER);
        } else {
            payable(msg.sender).sendValue(_amount);
        }

        emit Undelegated(msg.sender, _validator, _amount);
    }

    function _newCredential(uint256 _amount, uint256 _unbondingPeriod) private {
        credentials.push(
            WithdrawalCredential({
                requester: msg.sender,
                amount: _amount,
                requestTime: block.timestamp,
                withdrawableTime: block.timestamp + _unbondingPeriod,
                status: WithdrawalStatus.Requested
            })
        );

        emit NewCredential(credentials.length - 1, msg.sender, _amount, block.timestamp, _unbondingPeriod);
    }

    function withdraw(uint256 _cid) external {
        WithdrawalCredential storage _credential = credentials[_cid];
        require(_credential.status == WithdrawalStatus.Requested, "invalid credential");
        require(_credential.requester != msg.sender, "msg.sender is not requester");
        require(block.timestamp >= _credential.withdrawableTime, "not yet time to withdraw");

        payable(_credential.requester).sendValue(_credential.amount);
        _credential.status = WithdrawalStatus.Withdrew;

        emit Withdrew(_cid, msg.sender, _credential.amount);
    }
}
