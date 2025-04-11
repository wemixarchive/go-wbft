// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import "./IFeeRecipient.sol";

contract GovRewardeeImp {
    address public constant GOV_STAKING = address(0x1001);

    event RewardPaid(address indexed recipient, uint256 amount);

    receive() external payable {}

    modifier onlyGovStaking() {
        require(msg.sender == address(GOV_STAKING), "GovRewardee: caller is not the GovStaking contract");
        _;
    }

    function sendRewardTo(address payable recipient, uint256 amount, bool ifSendingFee) onlyGovStaking external {
        require(recipient != address(0), "GovRewardee: recipient is the zero address");
        require(amount > 0, "GovRewardee: amount is zero");
        require(amount <= address(this).balance, "GovRewardee: insufficient balance");

        uint256 size;
        assembly {
            size := extcodesize(recipient)
        }
        if (size > 0 && ifSendingFee) {
            // if it is for sending fee to recipient contract, try to call receiveFee()
            try IERC165(recipient).supportsInterface(type(IFeeRecipient).interfaceId) returns (bool supported) {
                if (supported) {
                    // IFeeRecipient is implemented
                    try IFeeRecipient(recipient).receiveFee{value: amount}(amount) {
                        emit RewardPaid(recipient, amount);
                        return;
                    } catch {
                        revert("Fee recipient contract reverted");
                    }
                } else {
                    (bool success, ) = recipient.call{value: amount}("");
                    require(success, "Fee transfer failed");
                }
            } catch {
                // if receiveFee is not implemented, transfer ether directly
                (bool success, ) = recipient.call{value: amount}("");
                require(success, "Fee transfer failed");
            }
        } else {
            // if it is for sending reward or recipient is EOA, transfer ether directly
            (bool success, ) = recipient.call{value: amount}("");
            require(success, "Fee transfer failed");
        }
        emit RewardPaid(recipient, amount);
    }
}
