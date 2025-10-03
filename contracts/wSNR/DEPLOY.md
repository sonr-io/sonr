# WSNR Deployment Scripts

This directory contains multiple deployment scripts for the WSNR (Wrapped SNR) smart contract.

## Prerequisites

1. **Compile the contract first**:

   ```bash
   cd contracts
   forge build
   ```

2. **Set up environment variables**:

   ```bash
   cd contracts
   cp .env.example .env
   # Edit .env and add your PRIVATE_KEY
   ```

3. **Ensure your Sonr node is running** with EVM enabled on `http://localhost:8545`

## Deployment Options

### Option 1: Foundry Script (Recommended)

The most robust option using Foundry's native scripting:

```bash
# Deploy to local Sonr node
./scripts/deploy-wsnr.sh

# Deploy to testnet
./scripts/deploy-wsnr.sh sonr-testnet
```

Features:

- Automatic chain detection
- Balance checking
- Contract verification
- Deployment info saved to `contracts/deployments/`

### Option 2: Simple Node.js Script

Requires Node.js and ethers.js:

```bash
# Install dependencies (if not already installed)
npm install ethers

# Deploy
node scripts/deploy-wsnr-simple.js [rpc-url]

# Example with custom RPC
node scripts/deploy-wsnr-simple.js http://192.168.1.100:8545
```

### Option 3: Python Script

Requires Python 3 and web3.py:

```bash
# Install dependencies (if not already installed)
pip3 install web3 eth-account

# Deploy
python3 scripts/deploy-wsnr.py [rpc-url]

# Example with custom RPC
python3 scripts/deploy-wsnr.py http://192.168.1.100:8545
```

### Option 4: Direct Foundry Command

For advanced users who want to customize deployment:

```bash
cd contracts
forge script script/DeployWSNR.s.sol:DeployWSNR \
    --rpc-url http://localhost:8545 \
    --broadcast \
    --private-key $PRIVATE_KEY
```

## Post-Deployment

After deployment, you'll receive:

- Contract address
- Transaction hash
- Block number
- Deployment info saved to `contracts/deployments/{chainId}-WSNR.json`

### Interacting with the Contract

**Using cast (Foundry)**:

```bash
# Deposit SNR to get WSNR
cast send <CONTRACT_ADDRESS> "deposit()" --value 1ether --rpc-url http://localhost:8545 --private-key $PRIVATE_KEY

# Check WSNR balance
cast call <CONTRACT_ADDRESS> "balanceOf(address)" <YOUR_ADDRESS> --rpc-url http://localhost:8545

# Withdraw SNR
cast send <CONTRACT_ADDRESS> "withdraw(uint256)" 1000000000000000000 --rpc-url http://localhost:8545 --private-key $PRIVATE_KEY
```

**Using web3 console**:

```javascript
// Connect to contract
const wsnr = new web3.eth.Contract(abi, contractAddress);

// Deposit
await wsnr.methods
  .deposit()
  .send({ from: account, value: web3.utils.toWei("1", "ether") });

// Check balance
const balance = await wsnr.methods.balanceOf(account).call();

// Withdraw
await wsnr.methods
  .withdraw(web3.utils.toWei("1", "ether"))
  .send({ from: account });
```

## Troubleshooting

### Cannot connect to node

- Ensure Sonr node is running: `make sh-testnet` or `docker-compose up sonr-node`
- Check if EVM is enabled with JSON-RPC on port 8545
- Try `curl http://localhost:8545` to test connectivity

### Insufficient balance

- Fund your deployer address with SNR tokens
- Check balance: `cast balance <YOUR_ADDRESS> --rpc-url http://localhost:8545`

### Contract not compiled

- Run `cd contracts && forge build` first
- Ensure you have Foundry installed: `curl -L https://foundry.paradigm.xyz | bash`

### Transaction fails

- Check gas prices and limits
- Ensure your account has enough SNR for gas
- Check if the network is synced
