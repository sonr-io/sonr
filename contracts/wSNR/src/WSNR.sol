// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ERC20} from "openzeppelin-contracts/contracts/token/ERC20/ERC20.sol";
import {ReentrancyGuard} from "openzeppelin-contracts/contracts/utils/ReentrancyGuard.sol";
import {IWSNR} from "./interfaces/IWSNR.sol";

/**
 * @title WSNR - Wrapped SNR
 * @notice ERC20 wrapper for native SNR tokens
 * @dev Implements a 1:1 wrapping mechanism for SNR tokens with deposit/withdraw functionality
 */
contract WSNR is IWSNR, ERC20, ReentrancyGuard {
    /**
     * @notice Initializes the WSNR token contract
     * @dev Sets token name as "Wrapped SNR" and symbol as "WSNR"
     */
    constructor() ERC20("Wrapped SNR", "WSNR") {}

    /**
     * @notice Fallback function to handle direct SNR transfers
     * @dev Automatically wraps sent SNR into WSNR tokens
     */
    receive() external payable {
        deposit();
    }

    /**
     * @notice Deposit native SNR and receive WSNR tokens
     * @dev Mints WSNR tokens equal to the amount of SNR sent
     */
    function deposit() public payable override nonReentrant {
        require(msg.value > 0, "WSNR: deposit amount must be greater than 0");
        
        _mint(msg.sender, msg.value);
        
        emit Deposit(msg.sender, msg.value);
    }

    /**
     * @notice Deposit native SNR to a specific address
     * @param to Address to receive the WSNR tokens
     * @dev Allows depositing on behalf of another address
     */
    function depositTo(address to) public payable override nonReentrant {
        require(msg.value > 0, "WSNR: deposit amount must be greater than 0");
        require(to != address(0), "WSNR: cannot deposit to zero address");
        
        _mint(to, msg.value);
        
        emit Deposit(to, msg.value);
    }

    /**
     * @notice Withdraw native SNR by burning WSNR tokens
     * @param amount Amount of WSNR to burn and SNR to receive
     * @dev Burns WSNR tokens and sends equivalent native SNR
     */
    function withdraw(uint256 amount) public override {
        withdrawTo(msg.sender, amount);
    }

    /**
     * @notice Withdraw native SNR to a specific address
     * @param to Address to receive the native SNR
     * @param amount Amount of WSNR to burn
     * @dev Allows withdrawing to a different address
     */
    function withdrawTo(address to, uint256 amount) public override nonReentrant {
        require(amount > 0, "WSNR: withdrawal amount must be greater than 0");
        require(to != address(0), "WSNR: cannot withdraw to zero address");
        require(balanceOf(msg.sender) >= amount, "WSNR: insufficient balance");
        
        _burn(msg.sender, amount);
        
        (bool success,) = to.call{value: amount}("");
        require(success, "WSNR: SNR transfer failed");
        
        emit Withdrawal(to, amount);
    }

    /**
     * @notice Get the total amount of SNR locked in the contract
     * @return The balance of native SNR held by the contract
     */
    function getReserve() public view returns (uint256) {
        return address(this).balance;
    }

    /**
     * @notice Verify that total supply equals contract balance
     * @dev This should always return true for proper 1:1 backing
     * @return Whether the contract is properly collateralized
     */
    function isFullyCollateralized() public view returns (bool) {
        return totalSupply() == address(this).balance;
    }
}
