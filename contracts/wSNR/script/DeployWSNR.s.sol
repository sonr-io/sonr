// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Script, console2} from "forge-std/Script.sol";
import {WSNR} from "../src/WSNR.sol";

contract DeployWSNR is Script {
    function run() external returns (WSNR) {
        // Get deployer private key from environment
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        
        // Get chain ID for logging
        uint256 chainId = block.chainid;
        
        console2.log("Deploying WSNR to chain ID:", chainId);
        console2.log("Deployer address:", vm.addr(deployerPrivateKey));
        console2.log("Deployer balance:", vm.addr(deployerPrivateKey).balance);
        
        // Start broadcasting transactions
        vm.startBroadcast(deployerPrivateKey);
        
        // Deploy WSNR contract
        WSNR wsnr = new WSNR();
        
        console2.log("WSNR deployed at:", address(wsnr));
        console2.log("Contract name:", wsnr.name());
        console2.log("Contract symbol:", wsnr.symbol());
        console2.log("Contract decimals:", wsnr.decimals());
        
        vm.stopBroadcast();
        
        // Write deployment info to file for reference
        string memory deploymentInfo = string(
            abi.encodePacked(
                "{\n",
                '  "contractName": "WSNR",\n',
                '  "address": "', vm.toString(address(wsnr)), '",\n',
                '  "chainId": ', vm.toString(chainId), ',\n',
                '  "deployer": "', vm.toString(vm.addr(deployerPrivateKey)), '",\n',
                '  "deploymentBlock": ', vm.toString(block.number), ',\n',
                '  "timestamp": ', vm.toString(block.timestamp), '\n',
                "}"
            )
        );
        
        // Save deployment info
        vm.writeFile(
            string(abi.encodePacked("deployments/", vm.toString(chainId), "-WSNR.json")),
            deploymentInfo
        );
        
        return wsnr;
    }
}