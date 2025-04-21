// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./dummy_ERC20.sol";

contract Token is ERC20 {
    constructor() ERC20("MyToken", "MTK") {}
}
