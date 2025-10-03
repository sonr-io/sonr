#!/bin/bash
# Starship E2E Testing Setup Script
# This script manages the Starship deployment for E2E testing

set -e

# Configuration
STARSHIP_VERSION="${STARSHIP_VERSION:-v2.0.0}"
STARSHIP_CONFIG="${STARSHIP_CONFIG:-chains/e2e-test.json}"
NAMESPACE="${NAMESPACE:-starship}"
TIMEOUT="${TIMEOUT:-300}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check for kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is required but not installed"
        exit 1
    fi
    
    # Check for Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is required but not installed"
        exit 1
    fi
    
    # Check for Starship CLI
    if ! command -v starship &> /dev/null; then
        log_warning "Starship CLI not found, installing..."
        install_starship
    fi
    
    log_info "Prerequisites check completed"
}

# Install Starship CLI
install_starship() {
    log_info "Installing Starship CLI..."
    npm install -g @starship-ci/cli@${STARSHIP_VERSION}
    
    if ! command -v starship &> /dev/null; then
        log_error "Failed to install Starship CLI"
        exit 1
    fi
    
    log_info "Starship CLI installed successfully"
}

# Build Docker image
build_docker_image() {
    log_info "Building Sonr Docker image..."
    
    # Build the image
    docker build -t sonr:local -f Dockerfile .
    
    if [ $? -ne 0 ]; then
        log_error "Failed to build Docker image"
        exit 1
    fi
    
    log_info "Docker image built successfully"
}

# Start Starship network
start_network() {
    log_info "Starting Starship network with config: ${STARSHIP_CONFIG}"
    
    # Create namespace if it doesn't exist
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    # Start Starship
    starship start --config ${STARSHIP_CONFIG} --namespace ${NAMESPACE}
    
    if [ $? -ne 0 ]; then
        log_error "Failed to start Starship network"
        exit 1
    fi
    
    log_info "Starship network started"
    
    # Wait for network to be ready
    wait_for_network
}

# Stop Starship network
stop_network() {
    log_info "Stopping Starship network..."
    
    starship stop --namespace ${NAMESPACE}
    
    if [ $? -ne 0 ]; then
        log_warning "Failed to stop Starship network gracefully"
    fi
    
    # Clean up namespace
    kubectl delete namespace ${NAMESPACE} --ignore-not-found=true
    
    log_info "Starship network stopped"
}

# Wait for network to be ready
wait_for_network() {
    log_info "Waiting for network to be ready (timeout: ${TIMEOUT}s)..."
    
    local start_time=$(date +%s)
    local current_time
    local elapsed
    
    while true; do
        current_time=$(date +%s)
        elapsed=$((current_time - start_time))
        
        if [ $elapsed -gt $TIMEOUT ]; then
            log_error "Timeout waiting for network to be ready"
            exit 1
        fi
        
        # Check if all pods are running
        local ready_pods=$(kubectl get pods -n ${NAMESPACE} --no-headers | grep -c "Running\|Completed" || true)
        local total_pods=$(kubectl get pods -n ${NAMESPACE} --no-headers | wc -l || true)
        
        if [ "$ready_pods" -eq "$total_pods" ] && [ "$total_pods" -gt 0 ]; then
            log_info "All pods are ready ($ready_pods/$total_pods)"
            break
        fi
        
        log_info "Waiting for pods to be ready ($ready_pods/$total_pods)..."
        sleep 5
    done
    
    # Additional wait for services to be fully initialized
    log_info "Waiting for services to initialize..."
    sleep 10
    
    # Verify chain connectivity
    verify_chain_connectivity
}

# Verify chain connectivity
verify_chain_connectivity() {
    log_info "Verifying chain connectivity..."
    
    # Port-forward to access the chains
    kubectl port-forward -n ${NAMESPACE} service/sonr-1-validator 1317:1317 &
    local pf_pid=$!
    
    sleep 5
    
    # Check if chain is responding
    if curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/syncing | grep -q "syncing"; then
        log_info "Chain 1 is responding"
    else
        log_error "Chain 1 is not responding"
        kill $pf_pid 2>/dev/null
        exit 1
    fi
    
    kill $pf_pid 2>/dev/null
    
    # Check second chain
    kubectl port-forward -n ${NAMESPACE} service/sonr-2-validator 1318:1318 &
    pf_pid=$!
    
    sleep 5
    
    if curl -s http://localhost:1318/cosmos/base/tendermint/v1beta1/syncing | grep -q "syncing"; then
        log_info "Chain 2 is responding"
    else
        log_error "Chain 2 is not responding"
        kill $pf_pid 2>/dev/null
        exit 1
    fi
    
    kill $pf_pid 2>/dev/null
    
    log_info "Chain connectivity verified"
}

# Get network info
get_network_info() {
    log_info "Network Information:"
    echo "========================"
    
    # Get pod status
    echo "Pods:"
    kubectl get pods -n ${NAMESPACE}
    echo ""
    
    # Get services
    echo "Services:"
    kubectl get services -n ${NAMESPACE}
    echo ""
    
    # Get endpoints
    echo "Endpoints:"
    echo "- Chain 1 REST: http://localhost:1317"
    echo "- Chain 1 RPC: http://localhost:26657"
    echo "- Chain 1 gRPC: localhost:9090"
    echo "- Chain 2 REST: http://localhost:1318"
    echo "- Chain 2 RPC: http://localhost:26658"
    echo "- Chain 2 gRPC: localhost:9091"
    echo "- Faucet: http://localhost:8000"
    echo "- Registry: http://localhost:8081"
    echo "- Explorer: http://localhost:8080"
    echo "========================"
}

# Port forward for local access
setup_port_forward() {
    log_info "Setting up port forwarding..."
    
    # Chain 1
    kubectl port-forward -n ${NAMESPACE} service/sonr-1-validator 1317:1317 26657:26657 9090:9090 &
    
    # Chain 2
    kubectl port-forward -n ${NAMESPACE} service/sonr-2-validator 1318:1318 26658:26658 9091:9091 &
    
    # Faucet
    kubectl port-forward -n ${NAMESPACE} service/faucet 8000:8000 &
    
    # Registry
    kubectl port-forward -n ${NAMESPACE} service/registry 8081:8081 &
    
    # Explorer
    kubectl port-forward -n ${NAMESPACE} service/explorer 8080:8080 &
    
    log_info "Port forwarding established"
    log_info "Press Ctrl+C to stop port forwarding and exit"
    
    # Wait for interrupt
    trap "kill 0" EXIT
    wait
}

# Run E2E tests
run_e2e_tests() {
    log_info "Running E2E tests..."
    
    # Set up port forwarding in background
    setup_port_forward &
    local pf_pid=$!
    
    # Wait for port forwarding to be established
    sleep 5
    
    # Run tests
    cd test/e2e
    
    if [ -n "$1" ]; then
        # Run specific test
        log_info "Running test: $1"
        go test -v -race -run "$1" ./...
    else
        # Run all tests
        log_info "Running all E2E tests"
        go test -v -race ./...
    fi
    
    local test_result=$?
    
    # Clean up port forwarding
    kill $pf_pid 2>/dev/null
    
    if [ $test_result -eq 0 ]; then
        log_info "E2E tests passed successfully"
    else
        log_error "E2E tests failed"
        exit 1
    fi
}

# Main command handler
case "${1:-}" in
    start)
        check_prerequisites
        build_docker_image
        start_network
        get_network_info
        ;;
    stop)
        stop_network
        ;;
    restart)
        stop_network
        sleep 5
        check_prerequisites
        build_docker_image
        start_network
        get_network_info
        ;;
    status)
        get_network_info
        ;;
    port-forward)
        setup_port_forward
        ;;
    test)
        check_prerequisites
        build_docker_image
        start_network
        run_e2e_tests "${2:-}"
        stop_network
        ;;
    test-only)
        run_e2e_tests "${2:-}"
        ;;
    *)
        echo "Starship E2E Testing Manager"
        echo ""
        echo "Usage: $0 {start|stop|restart|status|port-forward|test|test-only} [test-name]"
        echo ""
        echo "Commands:"
        echo "  start         - Start Starship network for E2E testing"
        echo "  stop          - Stop Starship network"
        echo "  restart       - Restart Starship network"
        echo "  status        - Show network status and information"
        echo "  port-forward  - Set up port forwarding for local access"
        echo "  test          - Start network, run tests, then stop network"
        echo "  test-only     - Run tests against existing network"
        echo ""
        echo "Examples:"
        echo "  $0 start                    # Start the network"
        echo "  $0 test                     # Run all E2E tests"
        echo "  $0 test TestBasicChain      # Run specific test"
        echo "  $0 test-only TestIBC        # Run test against existing network"
        echo "  $0 port-forward             # Access network locally"
        echo "  $0 stop                     # Stop the network"
        exit 1
        ;;
esac