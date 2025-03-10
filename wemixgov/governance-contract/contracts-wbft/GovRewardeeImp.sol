// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import {GovConst} from "./GovConst.sol";
import {GovStaking} from "./GovStaking.sol";

contract GovRewardeeImp {
    GovStaking public constant GOV_STAKING = GovStaking(payable(address(0x1001)));

    event RewardPaid(address indexed recipient, uint256 amount);

    receive() external payable {}

    modifier onlyGovStaking() {
        require(msg.sender == address(GOV_STAKING), "GovRewardee: caller is not the GovStaking contract");
        _;
    }

    function sendRewardTo(address payable recipient, uint256 amount) onlyGovStaking external {
        require(recipient != address(0), "GovRewardee: claimer is the zero address");
        require(amount > 0, "GovRewardee: amount is zero");
        require(amount <= address(this).balance, "GovRewardee: insufficient balance");

        recipient.transfer(amount);

        emit RewardPaid(recipient, amount);
    }
}
