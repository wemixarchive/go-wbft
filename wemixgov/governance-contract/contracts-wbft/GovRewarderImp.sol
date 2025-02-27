// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import {GovStaking} from "./GovStaking.sol";

contract GovRewarderImp {

    struct UserInfo {
        uint256 stakingAmount;
        uint256 accRewardPerStaking;
        uint256 calculatedReward;
        uint256 lastClaimed;
    }

    GovStaking public immutable staking; // 0x0
    address public immutable staker; // 0x1

    receive() external payable {}

    function claim(bool restake) external {
    }

    function update() public {
        uint256 accBalance = address(this).balance - lastBalance;
        lastBalance = address(this).balance;
        accRewardPerStaking = accRewardPerStaking + accBalance * REWARD_PRECISION / getTotalStaking();
    }

    function getTotalStaking() public view returns (uint256) {
        return staking.stakerInfo(staker).staking;
    }
}
