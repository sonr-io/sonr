// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Test, console2} from "forge-std/Test.sol";
import {WSNR} from "../src/WSNR.sol";
import {ReentrancyGuard} from "openzeppelin-contracts/contracts/utils/ReentrancyGuard.sol";

contract WSNRTest is Test {
    WSNR public wsnr;
    address public alice = address(0x1);
    address public bob = address(0x2);
    address public charlie = address(0x3);

    event Deposit(address indexed from, uint256 amount);
    event Withdrawal(address indexed to, uint256 amount);
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    function setUp() public {
        wsnr = new WSNR();
        
        // Fund test accounts
        vm.deal(alice, 100 ether);
        vm.deal(bob, 100 ether);
        vm.deal(charlie, 100 ether);
    }

    function testInitialState() public {
        assertEq(wsnr.name(), "Wrapped SNR");
        assertEq(wsnr.symbol(), "WSNR");
        assertEq(wsnr.decimals(), 18);
        assertEq(wsnr.totalSupply(), 0);
        assertEq(wsnr.getReserve(), 0);
        assertTrue(wsnr.isFullyCollateralized());
    }

    function testDeposit() public {
        uint256 depositAmount = 10 ether;
        
        vm.startPrank(alice);
        
        // Test deposit event
        vm.expectEmit(true, false, false, true);
        emit Deposit(alice, depositAmount);
        
        // Deposit SNR
        wsnr.deposit{value: depositAmount}();
        
        // Check balances
        assertEq(wsnr.balanceOf(alice), depositAmount);
        assertEq(wsnr.totalSupply(), depositAmount);
        assertEq(wsnr.getReserve(), depositAmount);
        assertEq(address(wsnr).balance, depositAmount);
        assertTrue(wsnr.isFullyCollateralized());
        
        vm.stopPrank();
    }

    function testDepositZeroAmount() public {
        vm.startPrank(alice);
        vm.expectRevert("WSNR: deposit amount must be greater than 0");
        wsnr.deposit{value: 0}();
        vm.stopPrank();
    }

    function testDepositTo() public {
        uint256 depositAmount = 5 ether;
        
        vm.startPrank(alice);
        
        // Test deposit event
        vm.expectEmit(true, false, false, true);
        emit Deposit(bob, depositAmount);
        
        // Deposit SNR to bob's account
        wsnr.depositTo{value: depositAmount}(bob);
        
        // Check balances
        assertEq(wsnr.balanceOf(alice), 0);
        assertEq(wsnr.balanceOf(bob), depositAmount);
        assertEq(wsnr.totalSupply(), depositAmount);
        assertEq(address(wsnr).balance, depositAmount);
        
        vm.stopPrank();
    }

    function testDepositToZeroAddress() public {
        vm.startPrank(alice);
        vm.expectRevert("WSNR: cannot deposit to zero address");
        wsnr.depositTo{value: 1 ether}(address(0));
        vm.stopPrank();
    }

    function testReceiveFallback() public {
        uint256 depositAmount = 3 ether;
        
        vm.startPrank(alice);
        
        // Test deposit event through fallback
        vm.expectEmit(true, false, false, true);
        emit Deposit(alice, depositAmount);
        
        // Send SNR directly to contract
        (bool success,) = address(wsnr).call{value: depositAmount}("");
        assertTrue(success);
        
        // Check balances
        assertEq(wsnr.balanceOf(alice), depositAmount);
        assertEq(wsnr.totalSupply(), depositAmount);
        
        vm.stopPrank();
    }

    function testWithdraw() public {
        uint256 depositAmount = 10 ether;
        uint256 withdrawAmount = 6 ether;
        
        vm.startPrank(alice);
        
        // First deposit
        wsnr.deposit{value: depositAmount}();
        uint256 aliceBalanceBefore = alice.balance;
        
        // Test withdrawal event
        vm.expectEmit(true, false, false, true);
        emit Withdrawal(alice, withdrawAmount);
        
        // Withdraw
        wsnr.withdraw(withdrawAmount);
        
        // Check balances
        assertEq(wsnr.balanceOf(alice), depositAmount - withdrawAmount);
        assertEq(wsnr.totalSupply(), depositAmount - withdrawAmount);
        assertEq(address(wsnr).balance, depositAmount - withdrawAmount);
        assertEq(alice.balance, aliceBalanceBefore + withdrawAmount);
        assertTrue(wsnr.isFullyCollateralized());
        
        vm.stopPrank();
    }

    function testWithdrawAll() public {
        uint256 depositAmount = 10 ether;
        
        vm.startPrank(alice);
        
        // Deposit and withdraw all
        wsnr.deposit{value: depositAmount}();
        uint256 aliceBalanceBefore = alice.balance;
        
        wsnr.withdraw(depositAmount);
        
        // Check everything is back to zero
        assertEq(wsnr.balanceOf(alice), 0);
        assertEq(wsnr.totalSupply(), 0);
        assertEq(address(wsnr).balance, 0);
        assertEq(alice.balance, aliceBalanceBefore + depositAmount);
        
        vm.stopPrank();
    }

    function testWithdrawZeroAmount() public {
        vm.startPrank(alice);
        wsnr.deposit{value: 1 ether}();
        
        vm.expectRevert("WSNR: withdrawal amount must be greater than 0");
        wsnr.withdraw(0);
        vm.stopPrank();
    }

    function testWithdrawInsufficientBalance() public {
        vm.startPrank(alice);
        wsnr.deposit{value: 5 ether}();
        
        vm.expectRevert("WSNR: insufficient balance");
        wsnr.withdraw(10 ether);
        vm.stopPrank();
    }

    function testWithdrawTo() public {
        uint256 depositAmount = 10 ether;
        uint256 withdrawAmount = 4 ether;
        
        vm.startPrank(alice);
        
        // Deposit from alice
        wsnr.deposit{value: depositAmount}();
        uint256 bobBalanceBefore = bob.balance;
        
        // Test withdrawal event
        vm.expectEmit(true, false, false, true);
        emit Withdrawal(bob, withdrawAmount);
        
        // Withdraw to bob
        wsnr.withdrawTo(bob, withdrawAmount);
        
        // Check balances
        assertEq(wsnr.balanceOf(alice), depositAmount - withdrawAmount);
        assertEq(bob.balance, bobBalanceBefore + withdrawAmount);
        
        vm.stopPrank();
    }

    function testWithdrawToZeroAddress() public {
        vm.startPrank(alice);
        wsnr.deposit{value: 1 ether}();
        
        vm.expectRevert("WSNR: cannot withdraw to zero address");
        wsnr.withdrawTo(address(0), 1 ether);
        vm.stopPrank();
    }

    function testERC20Transfer() public {
        uint256 depositAmount = 10 ether;
        uint256 transferAmount = 3 ether;
        
        vm.startPrank(alice);
        wsnr.deposit{value: depositAmount}();
        
        // Test transfer event
        vm.expectEmit(true, true, false, true);
        emit Transfer(alice, bob, transferAmount);
        
        // Transfer WSNR tokens
        assertTrue(wsnr.transfer(bob, transferAmount));
        
        // Check balances
        assertEq(wsnr.balanceOf(alice), depositAmount - transferAmount);
        assertEq(wsnr.balanceOf(bob), transferAmount);
        assertEq(wsnr.totalSupply(), depositAmount); // Total supply unchanged
        
        vm.stopPrank();
    }

    function testERC20Approve() public {
        uint256 depositAmount = 10 ether;
        uint256 approveAmount = 5 ether;
        
        vm.startPrank(alice);
        wsnr.deposit{value: depositAmount}();
        
        // Test approval event
        vm.expectEmit(true, true, false, true);
        emit Approval(alice, bob, approveAmount);
        
        // Approve bob to spend alice's WSNR
        assertTrue(wsnr.approve(bob, approveAmount));
        assertEq(wsnr.allowance(alice, bob), approveAmount);
        
        vm.stopPrank();
    }

    function testERC20TransferFrom() public {
        uint256 depositAmount = 10 ether;
        uint256 approveAmount = 6 ether;
        uint256 transferAmount = 4 ether;
        
        // Alice deposits and approves bob
        vm.startPrank(alice);
        wsnr.deposit{value: depositAmount}();
        wsnr.approve(bob, approveAmount);
        vm.stopPrank();
        
        // Bob transfers from alice to charlie
        vm.startPrank(bob);
        
        // Test transfer event
        vm.expectEmit(true, true, false, true);
        emit Transfer(alice, charlie, transferAmount);
        
        assertTrue(wsnr.transferFrom(alice, charlie, transferAmount));
        vm.stopPrank();
        
        // Check balances and allowance
        assertEq(wsnr.balanceOf(alice), depositAmount - transferAmount);
        assertEq(wsnr.balanceOf(charlie), transferAmount);
        assertEq(wsnr.allowance(alice, bob), approveAmount - transferAmount);
    }

    function testMultipleUsersDepositWithdraw() public {
        // Multiple users deposit
        vm.prank(alice);
        wsnr.deposit{value: 5 ether}();
        
        vm.prank(bob);
        wsnr.deposit{value: 3 ether}();
        
        vm.prank(charlie);
        wsnr.deposit{value: 2 ether}();
        
        // Check total supply and reserves
        assertEq(wsnr.totalSupply(), 10 ether);
        assertEq(wsnr.getReserve(), 10 ether);
        assertTrue(wsnr.isFullyCollateralized());
        
        // Users withdraw
        vm.prank(alice);
        wsnr.withdraw(2 ether);
        
        vm.prank(bob);
        wsnr.withdraw(1 ether);
        
        // Check final state
        assertEq(wsnr.totalSupply(), 7 ether);
        assertEq(wsnr.getReserve(), 7 ether);
        assertEq(wsnr.balanceOf(alice), 3 ether);
        assertEq(wsnr.balanceOf(bob), 2 ether);
        assertEq(wsnr.balanceOf(charlie), 2 ether);
        assertTrue(wsnr.isFullyCollateralized());
    }

    // Fuzz testing
    function testFuzzDeposit(uint256 amount) public {
        vm.assume(amount > 0 && amount <= 100 ether);
        
        vm.deal(alice, amount);
        vm.prank(alice);
        wsnr.deposit{value: amount}();
        
        assertEq(wsnr.balanceOf(alice), amount);
        assertEq(wsnr.totalSupply(), amount);
        assertEq(address(wsnr).balance, amount);
    }

    function testFuzzWithdraw(uint256 depositAmount, uint256 withdrawAmount) public {
        vm.assume(depositAmount > 0 && depositAmount <= 100 ether);
        vm.assume(withdrawAmount > 0 && withdrawAmount <= depositAmount);
        
        vm.deal(alice, depositAmount);
        vm.startPrank(alice);
        
        wsnr.deposit{value: depositAmount}();
        wsnr.withdraw(withdrawAmount);
        
        assertEq(wsnr.balanceOf(alice), depositAmount - withdrawAmount);
        assertEq(address(wsnr).balance, depositAmount - withdrawAmount);
        
        vm.stopPrank();
    }

    function testReentrancyProtection() public {
        ReentrantAttacker attacker = new ReentrantAttacker(wsnr);
        vm.deal(address(attacker), 10 ether);
        
        // The reentrancy guard prevents the second withdraw, which causes
        // the ETH transfer to fail, resulting in "SNR transfer failed" error
        vm.expectRevert("WSNR: SNR transfer failed");
        attacker.attack{value: 2 ether}();
    }
}

// Reentrancy test helper contract
contract ReentrantAttacker {
    WSNR public wsnr;
    uint256 public attackCount;
    
    constructor(WSNR _wsnr) {
        wsnr = _wsnr;
    }
    
    receive() external payable {
        attackCount++;
        if (attackCount < 2 && address(wsnr).balance >= 1 ether) {
            wsnr.withdraw(1 ether);
        }
    }
    
    function attack() external payable {
        wsnr.deposit{value: msg.value}();
        wsnr.withdraw(1 ether);
    }
}