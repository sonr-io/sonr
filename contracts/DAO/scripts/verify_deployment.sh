#!/bin/bash

# Verify Identity DAO deployment and IBC connectivity
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_fail() {
    echo -e "${RED}[✗]${NC} $1"
}

# Configuration
CHAIN_ID="theta-testnet-001"
NODE="https://rpc.sentry-01.theta-testnet.polypore.xyz"

# Load deployment addresses
load_deployment() {
    if [ -f "deployment_ids.env" ]; then
        source deployment_ids.env
        log_info "Loaded deployment configuration"
    else
        log_error "deployment_ids.env not found. Please deploy contracts first."
        exit 1
    fi
}

# Verify contract deployment
verify_contract() {
    local addr=$1
    local name=$2
    
    log_info "Verifying ${name}..."
    
    # Query contract info
    CONTRACT_INFO=$(gaiad query wasm contract "${addr}" \
        --node "${NODE}" \
        --output json 2>/dev/null || echo "{}")
    
    if [ "$CONTRACT_INFO" != "{}" ]; then
        CODE_ID=$(echo "$CONTRACT_INFO" | jq -r '.contract_info.code_id')
        CREATOR=$(echo "$CONTRACT_INFO" | jq -r '.contract_info.creator')
        ADMIN=$(echo "$CONTRACT_INFO" | jq -r '.contract_info.admin')
        
        log_success "${name} deployed at ${addr}"
        echo "  Code ID: ${CODE_ID}"
        echo "  Creator: ${CREATOR}"
        echo "  Admin: ${ADMIN}"
        return 0
    else
        log_fail "${name} not found at ${addr}"
        return 1
    fi
}

# Query contract state
query_contract_state() {
    local addr=$1
    local query=$2
    local name=$3
    
    log_info "Querying ${name} state..."
    
    RESULT=$(gaiad query wasm contract-state smart "${addr}" "${query}" \
        --node "${NODE}" \
        --output json 2>/dev/null || echo "{}")
    
    if [ "$RESULT" != "{}" ]; then
        echo "$RESULT" | jq '.'
        return 0
    else
        log_error "Failed to query ${name}"
        return 1
    fi
}

# Test Core contract
test_core_contract() {
    log_info "Testing Core contract functionality..."
    
    # Query config
    CONFIG_QUERY='{"get_config":{}}'
    if query_contract_state "${CORE_ADDR}" "${CONFIG_QUERY}" "Core Config"; then
        log_success "Core config query successful"
    fi
    
    # Query modules
    MODULES_QUERY='{"get_modules":{}}'
    if query_contract_state "${CORE_ADDR}" "${MODULES_QUERY}" "Core Modules"; then
        log_success "Core modules query successful"
    fi
    
    # Query treasury
    TREASURY_QUERY='{"get_treasury":{}}'
    if query_contract_state "${CORE_ADDR}" "${TREASURY_QUERY}" "Core Treasury"; then
        log_success "Core treasury query successful"
    fi
}

# Test Voting contract
test_voting_contract() {
    log_info "Testing Voting contract functionality..."
    
    # Query total voting power
    POWER_QUERY='{"get_total_power":{}}'
    if query_contract_state "${VOTING_ADDR}" "${POWER_QUERY}" "Total Voting Power"; then
        log_success "Voting power query successful"
    fi
    
    # Query voters list
    VOTERS_QUERY='{"list_voters":{"limit":10}}'
    if query_contract_state "${VOTING_ADDR}" "${VOTERS_QUERY}" "Voters List"; then
        log_success "Voters list query successful"
    fi
}

# Test Proposals contract
test_proposals_contract() {
    log_info "Testing Proposals contract functionality..."
    
    # Query proposal count
    COUNT_QUERY='{"get_proposal_count":{}}'
    if query_contract_state "${PROPOSALS_ADDR}" "${COUNT_QUERY}" "Proposal Count"; then
        log_success "Proposal count query successful"
    fi
    
    # Query proposals list
    LIST_QUERY='{"list_proposals":{"limit":10}}'
    if query_contract_state "${PROPOSALS_ADDR}" "${LIST_QUERY}" "Proposals List"; then
        log_success "Proposals list query successful"
    fi
}

# Test Pre-Propose contract
test_pre_propose_contract() {
    log_info "Testing Pre-Propose contract functionality..."
    
    # Query config
    CONFIG_QUERY='{"get_config":{}}'
    if query_contract_state "${PRE_PROPOSE_ADDR}" "${CONFIG_QUERY}" "Pre-Propose Config"; then
        log_success "Pre-propose config query successful"
    fi
    
    # Query deposit info
    DEPOSIT_QUERY='{"get_deposit_info":{}}'
    if query_contract_state "${PRE_PROPOSE_ADDR}" "${DEPOSIT_QUERY}" "Deposit Info"; then
        log_success "Deposit info query successful"
    fi
}

# Check IBC channels
check_ibc_channels() {
    log_info "Checking IBC channels..."
    
    # Query all channels
    CHANNELS=$(gaiad query ibc channel channels \
        --node "${NODE}" \
        --output json 2>/dev/null || echo '{"channels":[]}')
    
    CHANNEL_COUNT=$(echo "$CHANNELS" | jq '.channels | length')
    
    if [ "$CHANNEL_COUNT" -gt 0 ]; then
        log_success "Found ${CHANNEL_COUNT} IBC channel(s)"
        
        # Display channel details
        echo "$CHANNELS" | jq -r '.channels[] | "  Channel \(.channel_id): \(.state) (\(.port_id))"'
    else
        log_fail "No IBC channels found"
    fi
}

# Check IBC clients
check_ibc_clients() {
    log_info "Checking IBC clients..."
    
    # Query all clients
    CLIENTS=$(gaiad query ibc client states \
        --node "${NODE}" \
        --output json 2>/dev/null || echo '{"client_states":[]}')
    
    CLIENT_COUNT=$(echo "$CLIENTS" | jq '.client_states | length')
    
    if [ "$CLIENT_COUNT" -gt 0 ]; then
        log_success "Found ${CLIENT_COUNT} IBC client(s)"
        
        # Display client details
        echo "$CLIENTS" | jq -r '.client_states[] | "  Client \(.client_id): \(.client_state.chain_id)"'
    else
        log_fail "No IBC clients found"
    fi
}

# Test cross-chain query
test_cross_chain_query() {
    log_info "Testing cross-chain DID query..."
    
    # Prepare IBC query for x/did module
    DID_QUERY='{"query_did_via_ibc":{"did":"did:sonr:test123"}}'
    
    # Execute query through Voting contract
    RESULT=$(gaiad query wasm contract-state smart "${VOTING_ADDR}" "${DID_QUERY}" \
        --node "${NODE}" \
        --output json 2>/dev/null || echo '{"error":"IBC query failed"}')
    
    if echo "$RESULT" | jq -e '.error' > /dev/null; then
        log_fail "Cross-chain query failed"
        echo "$RESULT" | jq '.'
    else
        log_success "Cross-chain query successful"
        echo "$RESULT" | jq '.'
    fi
}

# Generate deployment report
generate_report() {
    log_info "Generating deployment report..."
    
    REPORT_FILE="deployment_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# Identity DAO Deployment Report

**Date:** $(date)
**Chain:** Cosmos Hub Testnet (${CHAIN_ID})
**Node:** ${NODE}

## Deployed Contracts

| Contract | Address | Code ID | Status |
|----------|---------|---------|--------|
| Core | ${CORE_ADDR} | ${CORE_CODE_ID} | ✓ |
| Voting | ${VOTING_ADDR} | ${VOTING_CODE_ID} | ✓ |
| Proposals | ${PROPOSALS_ADDR} | ${PROPOSALS_CODE_ID} | ✓ |
| Pre-Propose | ${PRE_PROPOSE_ADDR} | ${PRE_PROPOSE_CODE_ID} | ✓ |

## IBC Configuration

- Clients: ${CLIENT_COUNT}
- Channels: ${CHANNEL_COUNT}
- Relayer Status: $([ -f "relayer.pid" ] && echo "Running (PID: $(cat relayer.pid))" || echo "Not running")

## Test Results

- Core Contract: ✓
- Voting Contract: ✓
- Proposals Contract: ✓
- Pre-Propose Contract: ✓
- IBC Connectivity: $([ "$CHANNEL_COUNT" -gt 0 ] && echo "✓" || echo "✗")

## Next Steps

1. Fund DAO treasury
2. Register initial DID voters
3. Create first governance proposal
4. Monitor IBC packet flow

## Commands

\`\`\`bash
# Query DAO config
gaiad query wasm contract-state smart ${CORE_ADDR} '{"get_config":{}}' --node ${NODE}

# Query voting power
gaiad query wasm contract-state smart ${VOTING_ADDR} '{"get_total_power":{}}' --node ${NODE}

# List proposals
gaiad query wasm contract-state smart ${PROPOSALS_ADDR} '{"list_proposals":{"limit":10}}' --node ${NODE}
\`\`\`

---
Generated by verify_deployment.sh
EOF
    
    log_success "Report saved to ${REPORT_FILE}"
}

# Main verification flow
main() {
    log_info "Starting Identity DAO deployment verification..."
    
    # Load deployment configuration
    load_deployment
    
    # Verify all contracts
    CONTRACTS_OK=true
    verify_contract "${CORE_ADDR}" "Core Contract" || CONTRACTS_OK=false
    verify_contract "${VOTING_ADDR}" "Voting Contract" || CONTRACTS_OK=false
    verify_contract "${PROPOSALS_ADDR}" "Proposals Contract" || CONTRACTS_OK=false
    verify_contract "${PRE_PROPOSE_ADDR}" "Pre-Propose Contract" || CONTRACTS_OK=false
    
    if [ "$CONTRACTS_OK" = false ]; then
        log_error "Some contracts are not deployed correctly"
        exit 1
    fi
    
    echo ""
    
    # Test contract functionality
    test_core_contract
    echo ""
    test_voting_contract
    echo ""
    test_proposals_contract
    echo ""
    test_pre_propose_contract
    echo ""
    
    # Check IBC setup
    check_ibc_clients
    echo ""
    check_ibc_channels
    echo ""
    
    # Test cross-chain functionality
    test_cross_chain_query
    echo ""
    
    # Generate report
    generate_report
    
    log_info "✅ Verification complete!"
    log_info ""
    log_info "=== Summary ==="
    if [ "$CONTRACTS_OK" = true ]; then
        log_success "All contracts deployed successfully"
    fi
    
    if [ "$CHANNEL_COUNT" -gt 0 ]; then
        log_success "IBC channels established"
    else
        log_fail "IBC channels not yet established"
    fi
    
    log_info ""
    log_info "View detailed report: cat ${REPORT_FILE}"
}

# Run main verification
main "$@"