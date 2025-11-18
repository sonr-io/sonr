#!/bin/bash
# USDC Swap E2E Test Setup Script
# This script helps set up the environment for USDC swap E2E tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SONR_RPC="${SONR_RPC:-http://localhost:26657}"
SONR_REST="${SONR_REST:-http://localhost:1317}"
FAUCET_URL="${FAUCET_URL:-http://localhost:8000}"
NOBLE_CONNECTION="${NOBLE_CONNECTION:-connection-0}"

echo -e "${BLUE}=== USDC Swap E2E Test Setup ===${NC}\n"

# Function to print status
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Step 1: Check prerequisites
echo -e "${BLUE}Step 1: Checking prerequisites...${NC}"

if ! command_exists go; then
    print_error "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi
print_status "Go installed: $(go version | awk '{print $3}')"

if ! command_exists snrd; then
    print_error "snrd is not installed. Please install Sonr daemon."
    exit 1
fi
print_status "snrd installed"

if ! command_exists jq; then
    print_warning "jq is not installed. Some features may not work. Install with: brew install jq"
else
    print_status "jq installed"
fi

if ! command_exists curl; then
    print_error "curl is not installed."
    exit 1
fi
print_status "curl installed"

echo ""

# Step 2: Check network status
echo -e "${BLUE}Step 2: Checking network status...${NC}"

if curl -s "${SONR_REST}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
    CHAIN_ID=$(curl -s "${SONR_REST}/cosmos/base/tendermint/v1beta1/node_info" | jq -r '.default_node_info.network')
    print_status "Sonr testnet is running (Chain ID: $CHAIN_ID)"
else
    print_error "Sonr testnet is not accessible at ${SONR_REST}"
    echo "  Start testnet with: make testnet"
    exit 1
fi

if curl -s "${FAUCET_URL}/health" >/dev/null 2>&1; then
    print_status "Faucet is running"
else
    print_warning "Faucet is not accessible. Tests may fail without funded accounts."
fi

echo ""

# Step 3: Check IBC connections
echo -e "${BLUE}Step 3: Checking IBC connections...${NC}"

CONNECTIONS=$(snrd query ibc connection connections --node ${SONR_RPC} --output json 2>/dev/null | jq -r '.connections // [] | length')

if [ "$CONNECTIONS" -eq 0 ]; then
    print_warning "No IBC connections found"
    echo "  You need to establish a connection to Noble testnet"
    echo "  This typically requires:"
    echo "    1. Setting up an IBC relayer"
    echo "    2. Creating a client on each chain"
    echo "    3. Creating a connection"
    echo ""
    echo "  For testing without Noble, some tests will be skipped."
else
    print_status "Found $CONNECTIONS IBC connection(s)"

    # Check for Noble connection
    if snrd query ibc connection end "$NOBLE_CONNECTION" --node ${SONR_RPC} --output json 2>/dev/null | jq -e '.connection.state == "STATE_OPEN"' >/dev/null; then
        print_status "Noble connection ($NOBLE_CONNECTION) is OPEN"
    else
        print_warning "Noble connection ($NOBLE_CONNECTION) not found or not open"
        echo "  Update NOBLE_CONNECTION in config or create the connection"
    fi
fi

echo ""

# Step 4: Check IBC channels
echo -e "${BLUE}Step 4: Checking IBC channels...${NC}"

CHANNELS=$(snrd query ibc channel channels --node ${SONR_RPC} --output json 2>/dev/null | jq -r '.channels // [] | length')

if [ "$CHANNELS" -eq 0 ]; then
    print_warning "No IBC channels found"
else
    print_status "Found $CHANNELS IBC channel(s)"

    # List channels
    if command_exists jq; then
        echo "  Channels:"
        snrd query ibc channel channels --node ${SONR_RPC} --output json 2>/dev/null | \
            jq -r '.channels[] | "    \(.port_id)/\(.channel_id): \(.state)"'
    fi
fi

echo ""

# Step 5: Check DEX module
echo -e "${BLUE}Step 5: Checking DEX module...${NC}"

if snrd query dex params --node ${SONR_RPC} --output json 2>/dev/null | jq -e '.params.enabled == true' >/dev/null; then
    print_status "DEX module is enabled"

    if command_exists jq; then
        MAX_ACCOUNTS=$(snrd query dex params --node ${SONR_RPC} --output json 2>/dev/null | jq -r '.params.max_accounts_per_did // "N/A"')
        echo "  Max accounts per DID: $MAX_ACCOUNTS"
    fi
else
    print_error "DEX module is not enabled"
    exit 1
fi

echo ""

# Step 6: Check DID module
echo -e "${BLUE}Step 6: Checking DID module...${NC}"

if snrd query did params --node ${SONR_RPC} 2>/dev/null >/dev/null; then
    print_status "DID module is available"
else
    print_warning "DID module query failed (module may not be enabled)"
fi

echo ""

# Step 7: Create test configuration
echo -e "${BLUE}Step 7: Setting up test configuration...${NC}"

if [ ! -f "config.yaml" ]; then
    if [ -f "config.example.yaml" ]; then
        cp config.example.yaml config.yaml

        # Update connection ID if different from default
        if [ -n "$NOBLE_CONNECTION" ]; then
            sed -i.bak "s/connection-0/$NOBLE_CONNECTION/g" config.yaml
            rm -f config.yaml.bak
        fi

        print_status "Created config.yaml from example"
        echo "  Review and update config.yaml with your connection details"
    else
        print_warning "config.example.yaml not found"
    fi
else
    print_status "config.yaml already exists"
fi

echo ""

# Step 8: Install test dependencies
echo -e "${BLUE}Step 8: Installing test dependencies...${NC}"

if go mod download 2>/dev/null; then
    print_status "Downloaded Go dependencies"
else
    print_warning "Failed to download dependencies (may already be cached)"
fi

echo ""

# Step 9: Fund test accounts (optional)
echo -e "${BLUE}Step 9: Test account setup...${NC}"

read -p "Do you want to create and fund a test account? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Generate test account
    TEST_ACCOUNT=$(snrd keys add test_e2e_user --keyring-backend test --output json 2>/dev/null | jq -r '.address' || echo "")

    if [ -n "$TEST_ACCOUNT" ]; then
        print_status "Created test account: $TEST_ACCOUNT"

        # Try to fund via faucet
        if curl -s "${FAUCET_URL}/health" >/dev/null 2>&1; then
            FUND_RESPONSE=$(curl -s -X POST "${FAUCET_URL}/credit" \
                -H "Content-Type: application/json" \
                -d "{\"address\":\"$TEST_ACCOUNT\",\"denom\":\"snr\"}" || echo "")

            if echo "$FUND_RESPONSE" | grep -q "success"; then
                print_status "Funded test account with SNR tokens"
            else
                print_warning "Failed to fund account via faucet"
                echo "  Fund manually with: snrd tx bank send validator $TEST_ACCOUNT 100000000usnr"
            fi
        fi
    else
        print_warning "Failed to create test account"
    fi
fi

echo ""

# Step 10: Run quick verification
echo -e "${BLUE}Step 10: Running quick verification...${NC}"

if make check-prereqs 2>/dev/null; then
    print_status "All prerequisites verified"
else
    print_warning "Some verification checks failed"
fi

echo ""

# Final summary
echo -e "${BLUE}=== Setup Complete ===${NC}\n"
echo "Environment Summary:"
echo "  Sonr RPC:        $SONR_RPC"
echo "  Sonr REST:       $SONR_REST"
echo "  Faucet:          $FAUCET_URL"
echo "  Chain ID:        $CHAIN_ID"
echo "  IBC Connections: $CONNECTIONS"
echo "  IBC Channels:    $CHANNELS"
echo ""
echo "Next steps:"
echo "  1. Review and update config.yaml if needed"
echo "  2. Ensure Noble IBC connection is established"
echo "  3. Run tests with: make test"
echo ""
echo "Useful commands:"
echo "  make test              - Run all tests"
echo "  make test-verbose      - Run with detailed output"
echo "  make check-ibc         - Check IBC status"
echo "  make help              - Show all available commands"
echo ""

# Check if ready to run tests
if [ "$CONNECTIONS" -gt 0 ] && [ "$CHANNELS" -gt 0 ]; then
    echo -e "${GREEN}✓ System appears ready for E2E tests!${NC}"
    echo ""
    read -p "Run tests now? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        make test
    fi
else
    echo -e "${YELLOW}⚠ IBC setup incomplete. Some tests may be skipped.${NC}"
    echo "  Set up IBC connections to Noble for full test coverage."
fi
