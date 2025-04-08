// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import {MultiSigWallet} from "./MultiSigWallet.sol";


contract OperatorSample is MultiSigWallet {
    address public constant GOV_STAKING = address(0x1001);

    mapping(address => bool) private __claimers;

    // registerStaker
    function registerStaker(uint256 _amount, address _staker, address _feeRecipient, uint256 _feeRate, bytes calldata _blsPK) external onlyOwner isSingleOwner {
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

    //stake
    function stake(uint256 _amount) external onlyOwner isSingleOwner {
        bytes memory data = abi.encodeWithSignature(
        "stake(uint256)",
        _amount
        );
        (bool success, ) = GOV_STAKING.call{value: _amount}(data);
        require(success, "stake tx failed");
    }

    //unstake
    function unstake(uint256 _amount) external onlyOwner isSingleOwner {
        bytes memory data = abi.encodeWithSignature(
            "unstake(uint256)",
            _amount
        );
        (bool success, ) = GOV_STAKING.call(data);
        require(success, "unstake tx failed");
    }

    //claim
    function claim(address _staker, bool _restake) onlyClaimerOrOwner {
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
        require(!__claimers[_claimerToRemove], "claimer is not registered");
        __claimers[_claimerToRemove] = false;
    }

    modifier onlyClaimerOrOwner(address _addr) {
        require(__claimers[_addr] || isOwner(_addr), "only claimer or owner can execute");
        _;
    }
}