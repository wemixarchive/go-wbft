// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import "@openzeppelin/contracts/proxy/Proxy.sol";

contract GovRewarder is Proxy {
    address public immutable implementation;

    constructor(address _imp) {
        implementation = _imp;
    }

    function _implementation() internal view virtual override returns (address) {
        return implementation;
    }
}
