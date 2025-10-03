#!/bin/bash

# Test script for Identity DAO IBC Integration
# Validates cross-chain DID verification between Cosmos Hub and Sonr

set -e

# Configuration
DEPLOYMENT_FILE="${1:-cosmos-hub-deployment.json}"
TEST_DID="did:sonr:test123"
TEST_ADDRESS="cosmos1test..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Identity DAO IBC Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Load deployment info
if [ ! -f "$DEPLOYMENT_FILE" ]; then
    echo -e "${RED}Error: Deployment file not found: $DEPLOYMENT_FILE${NC}"
    exit 1
fi

# Extract contract addresses
VOTING_ADDR=$(jq -r '.contracts.voting.address' "$DEPLOYMENT_FILE")
PROPOSALS_ADDR=$(jq -r '.contracts.proposals.address' "$DEPLOYMENT_FILE")
VOTING_CHANNEL=$(jq -r '.contracts.voting.ibc_channel' "$DEPLOYMENT_FILE")
CHAIN_ID=$(jq -r '.chain_id' "$DEPLOYMENT_FILE")

echo "Using contracts from deployment:"
echo "  Voting: $VOTING_ADDR (Channel: $VOTING_CHANNEL)"
echo "  Proposals: $PROPOSALS_ADDR"
echo ""

# Function to run test
run_test() {
    local test_name=$1
    local test_cmd=$2
    
    echo -ne "${YELLOW}Testing: ${test_name}...${NC}"
    
    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e " ${GREEN}✓${NC}"
        return 0
    else
        echo -e " ${RED}✗${NC}"
        return 1
    fi
}

# Test 1: Check IBC channel status
echo -e "${BLUE}1. Checking IBC Channel Status${NC}"

CHANNEL_STATUS=$(hermes query channel end \
    --chain "$CHAIN_ID" \
    --port "wasm.$VOTING_ADDR" \
    --channel "$VOTING_CHANNEL" \
    2>/dev/null | jq -r '.state')

if [ "$CHANNEL_STATUS" = "Open" ]; then
    echo -e "${GREEN}✓ Channel is OPEN${NC}"
else
    echo -e "${RED}✗ Channel status: $CHANNEL_STATUS${NC}"
    exit 1
fi

# Test 2: Query Sonr chain for DIDs
echo -e "${BLUE}2. Querying Sonr DIDs via IBC${NC}"

# Send IBC packet to query DID
QUERY_MSG='{
    "send_did_query": {
        "did": "'$TEST_DID'",
        "channel": "'$VOTING_CHANNEL'"
    }
}'

TX_HASH=$(gaiad tx wasm execute "$VOTING_ADDR" "$QUERY_MSG" \
    --from deployer \
    --chain-id "$CHAIN_ID" \
    --gas-prices 0.025uatom \
    --gas auto \
    --gas-adjustment 1.5 \
    --yes \
    --output json 2>/dev/null | jq -r '.txhash')

echo "Query transaction: $TX_HASH"
sleep 6

# Check if packet was acknowledged
PACKET_ACK=$(hermes query packet acks \
    --chain "$CHAIN_ID" \
    --port "wasm.$VOTING_ADDR" \
    --channel "$VOTING_CHANNEL" \
    2>/dev/null | jq -r '.acks | length')

if [ "$PACKET_ACK" -gt 0 ]; then
    echo -e "${GREEN}✓ IBC packet acknowledged${NC}"
else
    echo -e "${YELLOW}⚠ No acknowledgment yet (may be pending)${NC}"
fi

# Test 3: Update voter with DID verification
echo -e "${BLUE}3. Testing Voter Update with DID${NC}"

UPDATE_MSG='{
    "update_voter": {
        "did": "'$TEST_DID'",
        "address": "'$TEST_ADDRESS'"
    }
}'

UPDATE_TX=$(gaiad tx wasm execute "$VOTING_ADDR" "$UPDATE_MSG" \
    --from deployer \
    --chain-id "$CHAIN_ID" \
    --gas-prices 0.025uatom \
    --gas auto \
    --gas-adjustment 1.5 \
    --yes \
    --output json 2>/dev/null | jq -r '.txhash')

echo "Update transaction: $UPDATE_TX"
sleep 6

# Query voter info
VOTER_QUERY='{
    "voter_info": {
        "did": "'$TEST_DID'"
    }
}'

VOTER_INFO=$(gaiad query wasm contract-state smart "$VOTING_ADDR" "$VOTER_QUERY" \
    --output json 2>/dev/null | jq -r '.data')

if [ "$VOTER_INFO" != "null" ]; then
    echo -e "${GREEN}✓ Voter registered successfully${NC}"
    echo "Voter info: $VOTER_INFO"
else
    echo -e "${RED}✗ Voter not found${NC}"
fi

# Test 4: Create proposal through IBC
echo -e "${BLUE}4. Testing Proposal Creation${NC}"

PROPOSAL_MSG='{
    "propose": {
        "title": "Test IBC Proposal",
        "description": "Testing cross-chain governance",
        "msgs": [],
        "proposer_did": "'$TEST_DID'"
    }
}'

PROPOSAL_TX=$(gaiad tx wasm execute "$PROPOSALS_ADDR" "$PROPOSAL_MSG" \
    --from deployer \
    --chain-id "$CHAIN_ID" \
    --gas-prices 0.025uatom \
    --gas auto \
    --gas-adjustment 1.5 \
    --yes \
    --output json 2>/dev/null | jq -r '.txhash')

echo "Proposal transaction: $PROPOSAL_TX"
sleep 6

# Query proposals
PROPOSALS_QUERY='{"list_proposals": {"limit": 10}}'

PROPOSALS=$(gaiad query wasm contract-state smart "$PROPOSALS_ADDR" "$PROPOSALS_QUERY" \
    --output json 2>/dev/null | jq -r '.data.proposals | length')

if [ "$PROPOSALS" -gt 0 ]; then
    echo -e "${GREEN}✓ Proposal created successfully${NC}"
    echo "Total proposals: $PROPOSALS"
else
    echo -e "${RED}✗ No proposals found${NC}"
fi

# Test 5: Check relayer metrics
echo -e "${BLUE}5. Checking Relayer Health${NC}"

RELAYER_STATUS=$(hermes health-check 2>/dev/null | grep -c "OK" || true)

if [ "$RELAYER_STATUS" -gt 0 ]; then
    echo -e "${GREEN}✓ Relayer is healthy${NC}"
else
    echo -e "${YELLOW}⚠ Relayer may need attention${NC}"
fi

# Test 6: Packet flow statistics
echo -e "${BLUE}6. IBC Packet Statistics${NC}"

echo "Channel: $VOTING_CHANNEL"

# Get packet commitments
PENDING_PACKETS=$(hermes query packet commitments \
    --chain "$CHAIN_ID" \
    --port "wasm.$VOTING_ADDR" \
    --channel "$VOTING_CHANNEL" \
    2>/dev/null | jq -r '.commitments | length')

echo "Pending packets: $PENDING_PACKETS"

# Get packet acknowledgments
TOTAL_ACKS=$(hermes query packet acks \
    --chain "$CHAIN_ID" \
    --port "wasm.$VOTING_ADDR" \
    --channel "$VOTING_CHANNEL" \
    2>/dev/null | jq -r '.acks | length')

echo "Total acknowledgments: $TOTAL_ACKS"

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"

TESTS_PASSED=0
TESTS_TOTAL=6

[ "$CHANNEL_STATUS" = "Open" ] && ((TESTS_PASSED++))
[ "$PACKET_ACK" -gt 0 ] && ((TESTS_PASSED++))
[ "$VOTER_INFO" != "null" ] && ((TESTS_PASSED++))
[ "$PROPOSALS" -gt 0 ] && ((TESTS_PASSED++))
[ "$RELAYER_STATUS" -gt 0 ] && ((TESTS_PASSED++))
[ "$PENDING_PACKETS" -eq 0 ] && ((TESTS_PASSED++))

if [ "$TESTS_PASSED" -eq "$TESTS_TOTAL" ]; then
    echo -e "${GREEN}✓ All tests passed! ($TESTS_PASSED/$TESTS_TOTAL)${NC}"
else
    echo -e "${YELLOW}⚠ Some tests failed ($TESTS_PASSED/$TESTS_TOTAL)${NC}"
fi

echo ""
echo "Next steps:"
echo "1. Monitor packet relay: hermes query packet pending --chain $CHAIN_ID"
echo "2. Check channel balance: hermes query channel balance --chain $CHAIN_ID"
echo "3. View relayer logs: hermes start --log-level debug"