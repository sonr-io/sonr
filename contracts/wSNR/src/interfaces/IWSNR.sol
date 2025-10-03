// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";

/**
 * @title IWSNR
 * @notice Interface for Wrapped SNR (WSNR) token contract
 * @dev Extends ERC20 with deposit and withdraw functionality for wrapping native SNR tokens
 */
interface IWSNR is IERC20 {
    /**
     * @notice Emitted when native SNR is deposited and WSNR is minted
     * @param from Address that deposited SNR
     * @param amount Amount of SNR deposited (and WSNR minted)
     */
    event Deposit(address indexed from, uint256 amount);

    /**
     * @notice Emitted when WSNR is burned and native SNR is withdrawn
     * @param to Address that received SNR
     * @param amount Amount of WSNR burned (and SNR withdrawn)
     */
    event Withdrawal(address indexed to, uint256 amount);

    /**
     * @notice Deposit native SNR and receive WSNR tokens
     * @dev Mints WSNR tokens equal to the amount of SNR sent
     */
    function deposit() external payable;

    /**
     * @notice Withdraw native SNR by burning WSNR tokens
     * @param amount Amount of WSNR to burn and SNR to receive
     * @dev Burns WSNR tokens and sends equivalent native SNR
     */
    function withdraw(uint256 amount) external;

    /**
     * @notice Deposit native SNR to a specific address
     * @param to Address to receive the WSNR tokens
     * @dev Allows depositing on behalf of another address
     */
    function depositTo(address to) external payable;

    /**
     * @notice Withdraw native SNR to a specific address
     * @param to Address to receive the native SNR
     * @param amount Amount of WSNR to burn
     * @dev Allows withdrawing to a different address
     */
    function withdrawTo(address to, uint256 amount) external;
}