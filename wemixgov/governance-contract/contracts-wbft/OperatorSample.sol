// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import {IFeeRecipient, IERC165} from "./IFeeRecipient.sol";
import "./IMultiSigWallet.sol";


contract OperatorSample is IMultiSigWallet, IFeeRecipient {

    /* =========== STATE VARIABLES ===========*/

    // multiSig variables
    uint256 constant public MAX_OWNER_COUNT = 50;    // number of max owner
    uint256 public quorum;  // minimum required confirmation number
    Transaction[] public transactions; // transaction struct set
    mapping(bytes32 => uint256) public proposalHashToTxId; // mapping from hash(proposer,blockNumber) =>  transaction id
    address[] public owners;  // owners address set
    mapping(address => bool) private _isOwner; // mapping from owner => bool
    mapping(uint256 => mapping(address => bool)) public isConfirmed;   // mapping from transaction id => owner => bool
    mapping(address => bool) private __claimers;  // mapping from claimer => bool

    // wbft gov variables
    address public constant GOV_STAKING = address(0x1001);
    bytes4 private constant _WITHDRAW_FEE_AMOUNT_SELECTOR = bytes4(keccak256("withdrawFeeAmount(address)"));
    bytes4 private constant _WITHDRAW_REWARD_AMOUNT_SELECTOR = bytes4(keccak256("withdrawRewardAmount(address)"));
    bytes4 private constant _WITHDRAW_UNSTAKED_AMOUNT_SELECTOR = bytes4(keccak256("withdrawUnstakedAmount(address)"));
    bytes4 private constant _REGISTER_STAKER_SELECTOR = bytes4(keccak256("registerStaker(uint256,address,address,uint256,bytes)"));
    bytes4 private constant _STAKE_SELECTOR = bytes4(keccak256("stake(uint256)"));
    bool private _receivingRewardStat;
    uint256 private _unstakedAmount;
    uint256 private _rewardAmount;
    uint256 private _feeAmount;

    // reentrancy guard variables
    uint8 private constant _NOT_ENTERED = 1;
    uint8 private constant _ENTERED = 2;
    uint8 private _status;

    /* =========== EVENTS  ===========*/

    event ReceivedUnstaked(address indexed from, uint256 amount);
    event ReceivedReward(address indexed from, uint256 amount);
    event SentUnstakedAmount(address indexed to, uint256 amount);
    event SentRewardAmount(address indexed to, uint256 amount);


    /* =========== MODIFIERS  ===========*/

    modifier onlyOwner() {
        require(_isOwner[msg.sender], "MultiSig: Only Owner can access.");
        _;
    }

    modifier isOneOfOwner(address _owner) {
        require(_isOwner[_owner], "MultiSig: Only Owner can access.");
        _;
    }

    modifier isNotOneOfOwner(address _owner) {
        require(!_isOwner[_owner], "MultiSig: Owner can not access.");
        _;
    }

    modifier onlyWalletOrSingleOwner() {
        // if contract has single owner, multiSig wallet is not activated
        // owner can execute this function right away
        if (owners.length == 1 ) {
            require(msg.sender == owners[0], "SingleOwner: Only Owner can access.");
        } else {
            require(msg.sender == address(this), "MultiSig: Only Wallet can access.");
        }
        _;
    }

    modifier notNull(address _address) {
        require(_address != address(0), "MultiSig: Owner cannot be 0.");
        _;
    }

    modifier isTransactionExist(uint256 _transactionId) {
        require(_transactionId < transactions.length, "MultiSig: Transaction does not exist.");
        _;
    }

    modifier notExecuted(uint256 _transactionId) {
        require(!transactions[_transactionId].executed, "MultiSig: Transaction is already executed.");
        _;
    }

    modifier notConfirmed(uint256 _transactionId) {
        require(!isConfirmed[_transactionId][msg.sender], "MultiSig: Transaction is already confirmed");
        _;
    }

    modifier validRequirement(uint256 _ownerCount, uint256 _quorum) {
        require(
            _ownerCount <= MAX_OWNER_COUNT
            && _quorum <= _ownerCount
            && _quorum != 0
            && _ownerCount != 0,
            "MultiSig: Invalid Requirement."
        );
        _;
    }

    modifier onlyClaimerOrOwner(address _addr) {
        require(__claimers[_addr] || isOwner(_addr), "Operator: Only claimer or owner can execute");
        _;
    }

    modifier nonReentrant() {
        require(_status != _ENTERED, "ReentrancyGuard: reentrant call");
        _status = _ENTERED;
        _;
        _status = _NOT_ENTERED;
    }

    constructor(address[] memory _owners, uint256 _quorum) {
        require(_owners.length > 0, "MultiSig: Owners length must be at least 1");

        if (_owners.length > 2) {
            require(_quorum > 1 && _quorum <= _owners.length, "MultiSig: Number of confirmations does not satisfy quorum.");
            quorum = _quorum;
        } else {
            // if owners.length is less than 3, set quorum as 1
            quorum = 1;
        }

        for (uint256 i = 0; i < _owners.length; i++) {
            address owner = _owners[i];

            require(owner != address(0), "MultiSig: Owner cannot be 0.");
            require(!_isOwner[owner], "MultiSig: Owner Address is duplicated.");

            _isOwner[owner] = true;
            owners.push(owner);
        }
        _status = _NOT_ENTERED;
    }

    // @notice Function is to receive reward, unstaked, or ether that eoa sent to the contract.
    // @dev When _receivingRewardStat is true, means the received ether is reward value so increase the _rewardAmount.
    // Else, increase the _unstakedAmount
    receive() external payable {
        if (_receivingRewardStat) {
            _rewardAmount += msg.value;
            emit ReceivedReward(msg.sender, msg.value);
            _receivingRewardStat = false;
        }else {
            _unstakedAmount += msg.value;
            emit ReceivedUnstaked(msg.sender, msg.value);
        }
        emit Deposit(msg.sender, msg.value, address(this).balance);
    }

    /* ========== EXTERNAL FUNCTION ========== */

    // implements IERC165
    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return
            interfaceId == type(IFeeRecipient).interfaceId ||
            interfaceId == type(IERC165).interfaceId;
    }

    // implements IFeeRecipient
    function receiveFee(uint256 _amount) external payable {
        require(_amount == msg.value, "FeeRecipient: FeeAmount and value sent mismatched");
        _feeAmount += _amount;
        emit ReceivedFee(msg.sender, _amount);
    }

    // implements IFeeRecipient
    // @notice Function is for sending _to a fee amount
    function withdrawFeeAmount(address _to, uint256 _amount) external onlyWalletOrSingleOwner notNull(_to) nonReentrant {
        require(_amount <= _feeAmount, "FeeRecipient: Withdraw Amount is greater than fee amount");
        _feeAmount -= _amount;
        (bool success, ) = payable(_to).call{value:_amount}("");
        require(success, "failed to send fee amount");
        emit SentFeeAmount(_to, _amount);
    }

    // @notice Function is for sending _to a unstaked amount
    function withdrawUnstakedAmount(address _to, uint256 _amount) external onlyWalletOrSingleOwner notNull(_to) nonReentrant {
        require(_amount <= _unstakedAmount, "Operator: Withdraw Amount is greater than unstaked amount");
        _unstakedAmount -= _amount;
        (bool success, ) = payable(_to).call{value:_amount}("");
        emit SentUnstakedAmount(_to, _amount);
        require(success, "failed to send unstaked amount");
    }

    // @notice Function is for sending _to a reward amount
    function withdrawRewardAmount(address _to, uint256 _amount) external onlyWalletOrSingleOwner notNull(_to) nonReentrant {
        require(_amount <= _rewardAmount, "Operator: Withdraw Amount is greater than reward amount");
        _rewardAmount -= _amount;
        (bool success, ) = payable(_to).call{value:_amount}("");
        require(success, "failed to send reward amount");
        emit SentRewardAmount(_to, _amount);
    }

    // @notice Function calls the registerStaker method of the GovStaking contract with Ether value
    // You cannot directly submit a transaction to GovStaking's registerStaker method since it's blocked for tracking amounts.
    // Instead, you must submit a transaction that calls this wrapper function, or just call this function directly if there's only one owner.
    function registerStaker(uint256 _amount, address _staker, address _feeRecipient, uint256 _feeRate, bytes calldata _blsPK) external onlyWalletOrSingleOwner nonReentrant {
        require(_amount <= _unstakedAmount, "Operator : Amount is greater than unstaked amount");
        _unstakedAmount -= _amount;
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
        emit SentUnstakedAmount(GOV_STAKING, _amount);
    }

    // @notice Function calls the stake method of the GovStaking contract with Ether value
    // You cannot directly submit a transaction to GovStaking's stake method since it's blocked for tracking amounts.
    // Instead, you must submit a transaction that calls this wrapper function, or just call this function directly if there's only one owner.
    function stake(uint256 _amount) external onlyWalletOrSingleOwner nonReentrant {
        require(_amount <= _unstakedAmount, "Operator : Amount is greater than unstaked amount");
        _unstakedAmount -= _amount;
        bytes memory data = abi.encodeWithSignature(
            "stake(uint256)",
            _amount
        );
        (bool success, ) = GOV_STAKING.call{value: _amount}(data);
        require(success, "stake tx failed");
        emit SentUnstakedAmount(GOV_STAKING, _amount);
    }

    // @notice Function calls the unstake method of the GovStaking contract with Ether value
    // You cannot directly submit a transaction to GovStaking's unstake method since it's blocked for tracking amounts.
    // Instead, you must submit a transaction that calls this wrapper function, or just call this function directly if there's only one owner.
    function unstake(uint256 _amount) external onlyWalletOrSingleOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature(
            "unstake(uint256)",
            _amount
        );
        (bool success, ) = GOV_STAKING.call(data);
        require(success, "unstake tx failed");
    }

    // @notice Function calls the claim method of the GovStaking contract
    // Owner or claime rcan directly call this method without multiSig signing.
    function claim(address _staker, bool _restake) external onlyClaimerOrOwner(msg.sender) {
        bytes memory data = abi.encodeWithSignature(
            "claim(address, bool)",
            _staker,
            _restake
        );
        _receivingRewardStat = true;
        (bool success, ) = GOV_STAKING.call(data);
        if (!success) {
            _receivingRewardStat = false;
            revert("claim tx failed");
        }
    }

    // @notice Function to add owner
    // Added owner will be automatically added to claimer
    function addOwner(address _newOwner) external onlyWalletOrSingleOwner notNull(_newOwner) isNotOneOfOwner(_newOwner) validRequirement(owners.length + 1, quorum){
        _isOwner[_newOwner] = true;
        owners.push(_newOwner);

        _addClaimer(_newOwner);
        emit AddOwner(_newOwner);
    }

    // @notice Function to add remove owner
    // Removed owner will be automatically removed from claimer
    function removeOwner(address _owner) external onlyWalletOrSingleOwner isOneOfOwner(_owner){
        _isOwner[_owner] = false;

        for (uint256 i=0; i<owners.length;) {
            if (owners[i] == _owner) {
                owners[i] = owners[owners.length - 1];
                owners.pop();
                break;
            }
            unchecked {
                i++;
            }
        }

        if (quorum > owners.length) {
            // changeQuorum(owners.length);
            quorum = owners.length;
            emit ChangeQuorum(owners.length);
        }
        _removeClaimer(_owner);
        emit RemoveOwner(_owner);
    }

    // @notice Function to replace owner
    // Claimer will also be replaced
    function replaceOwner(address _owner, address _newOwner) external onlyWalletOrSingleOwner isOneOfOwner(_owner) isNotOneOfOwner(_newOwner){
        for (uint i=0; i<owners.length;) {
            if (owners[i] == _owner) {
                owners[i] = _newOwner;
                break;
            }
            unchecked {
                i++;
            }
        }

        _isOwner[_owner] = false;
        _isOwner[_newOwner] = true;

        _removeClaimer(_owner);
        _addClaimer(_newOwner);

        emit RemoveOwner(_owner);
        emit AddOwner(_newOwner);
    }

    // @notice Function to change quorum
    function changeQuorum(uint256 _quorum) external onlyWalletOrSingleOwner validRequirement(owners.length, _quorum) {
        quorum = _quorum;
        emit ChangeQuorum(_quorum);
    }

    /* ========== PUBLIC FUNCTION ========== */

    // @notice Function to submit a transaction for multiSig signing.
    // Only owner can access.
    function submitTransaction(address _to, uint256 _value, bytes memory _data) public onlyOwner {
        bytes32 proposalHash = keccak256(abi.encodePacked( msg.sender, block.number));
        require(proposalHashToTxId[proposalHash] == 0, "MultiSig: Duplicate proposal in same block");

        // transaction id starts with 1
        uint256 transactionId = transactions.length + 1;
        proposalHashToTxId[proposalHash] = transactionId;

        if (_value > 0 ) {
            // if the transaction is to transfer value from this contract, force to use allowed method
            bytes4 selector = bytes4(_data);
            require( _to == address(this) && (
                selector == _WITHDRAW_FEE_AMOUNT_SELECTOR ||
                selector ==  _WITHDRAW_REWARD_AMOUNT_SELECTOR ||
                selector == _WITHDRAW_UNSTAKED_AMOUNT_SELECTOR ||
                selector == _REGISTER_STAKER_SELECTOR ||
                selector == _STAKE_SELECTOR),
                "Operator: Use proper withdraw/stake functions to transfer value from contract"
            );
        }
        transactions.push(
            Transaction(
                {
                    to: _to,
                    value: _value,
                    data: _data,
                    executed: false,
                    currentNumberOfConfirmations: 0
                }
            )
        );
        emit SubmitTransaction(msg.sender, transactionId, _to, _value, _data);
    }

    function executeTransaction(uint256 _transactionId) public payable onlyOwner isTransactionExist(_transactionId) notExecuted(_transactionId) {
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

    // @notice Function to confirm a transaction submitted.
    // Only owner can access.
    function confirmTransaction(uint256 _transactionId) public onlyOwner isTransactionExist(_transactionId) notExecuted(_transactionId) notConfirmed(_transactionId) {
        Transaction storage transaction = transactions[_transactionId];
        transaction.currentNumberOfConfirmations += 1;
        isConfirmed[_transactionId][msg.sender] = true;

        emit ConfirmTransaction(msg.sender, _transactionId);
    }

    // @notice Function to revoke confirmation.
    // Only owner who confirmed the transaction can access.
    function revokeConfirmation(uint256 _transactionId) public onlyOwner isTransactionExist(_transactionId) notExecuted(_transactionId) {
        Transaction storage transaction = transactions[_transactionId];

        require(isConfirmed[_transactionId][msg.sender], "MultiSig: Transaction is not confirmed");

        transaction.currentNumberOfConfirmations -= 1;
        isConfirmed[_transactionId][msg.sender] = false;

        emit RevokeConfirmation(msg.sender, _transactionId);
    }

    // @notice Function to get transaction id
    function getTransactionId(address _proposer, uint256 _blockNumber) public view returns (uint256) {
        return proposalHashToTxId[keccak256(abi.encodePacked(_proposer, _blockNumber))];
    }


    /* ========== INTERNAL FUNCTION ========== */

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

    function isOwner(address _owner) public view returns (bool) {
        return _isOwner[_owner];
    }

    /**
     * @notice Get Owners.
     */
    function getOwners() public view returns (address[] memory) {
        return owners;
    }

    /**
     * @notice Get Owners Count.
     */
    function getOwnerCount() public view returns (uint256) {
        return owners.length;
    }

    /**
     * @notice Get Transaction.
     * @return to Target address.
     * @return value Ether value.
     * @return data Transaction data.
     * @return executed Execute or Not.
     * @return currentNumberOfConfirmations Number of Confirmations.
     */
    function getTransaction(uint256 _transactionId) public view returns (address to, uint256 value, bytes memory data, bool executed, uint256 currentNumberOfConfirmations) {
        Transaction storage transaction = transactions[_transactionId];

        return (
            transaction.to,
            transaction.value,
            transaction.data,
            transaction.executed,
            transaction.currentNumberOfConfirmations
        );
    }

    /**
     * @notice Get Transaction Count.
     */
    function getTransactionCount() public view returns (uint256) {
        return transactions.length;
    }


}