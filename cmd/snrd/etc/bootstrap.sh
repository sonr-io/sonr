#!/bin/bash
set -e

# ================================
# SONR TESTNET BOOTSTRAP SCRIPT
# ================================

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_header() { echo -e "${BLUE}==== $1 ====${NC}"; }

# Script directory and repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Load environment variables if .env exists
if [ -f "$REPO_ROOT/.env" ]; then
    export $(grep -v '^#' "$REPO_ROOT/.env" | xargs)
fi

# Default values
export CHAIN_ID=${CHAIN_ID:-"sonrtest_1-1"}
export DENOM=${DENOM:-"usnr"}
export DOCKER_IMAGE=${DOCKER_IMAGE:-"onsonr/snrd:latest"}

# Function to check prerequisites
check_prerequisites() {
    log_header "Checking prerequisites"
    
    local has_error=false
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        has_error=true
    else
        log_info "Docker: $(docker --version)"
    fi
    
    # Check Docker Compose
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed"
        has_error=true
    else
        log_info "Docker Compose: $(docker compose version)"
    fi
    
    # Check jq
    if ! command -v jq &> /dev/null; then
        log_warn "jq is not installed - some commands may not work"
        log_warn "Install with: apt-get install jq (Ubuntu) or brew install jq (macOS)"
    else
        log_info "jq: $(jq --version)"
    fi
    
    if [ "$has_error" = true ]; then
        log_error "Prerequisites check failed"
        exit 1
    fi
    
    log_info "All prerequisites met"
    echo ""
}

# Function to initialize validators
init_validators() {
    log_header "Initializing validators and sentries"

    # Check for local snrd binary
    if ! command -v snrd &> /dev/null; then
        log_error "snrd binary not found in PATH"
        log_error "Please install snrd locally or add it to your PATH"
        log_error "The initialization requires a local snrd binary to avoid permission issues"
        exit 1
    fi

    if [ -f "$SCRIPT_DIR/init-testnet.sh" ]; then
        log_info "Using init-testnet.sh (local snrd binary initialization)"
        log_info "snrd location: $(which snrd)"
        "$SCRIPT_DIR/init-testnet.sh"
    else
        log_error "init-testnet.sh not found in $SCRIPT_DIR!"
        exit 1
    fi
}

# Function to start testnet
start_testnet() {
    log_header "Starting testnet"

    # Check if validators are initialized
    if [ ! -d "val-alice" ] || [ ! -d "sentry-alice" ]; then
        log_warn "Validators not initialized, running init first..."
        init_validators
    fi

    # Pull latest image
    log_info "Pulling Docker image: $DOCKER_IMAGE"
    docker pull "$DOCKER_IMAGE"

    # Start services
    log_info "Starting services with docker compose..."
    docker compose up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to start..."
    sleep 5
    
    # Check status
    show_status
    
    log_info "Testnet started successfully!"
    echo ""
    log_info "View logs: docker compose logs -f"
    log_info "Stop testnet: ./bootstrap.sh stop"
}

# Function to stop testnet
stop_testnet() {
    log_header "Stopping testnet"
    
    docker compose down
    log_info "Testnet stopped"
}

# Function to restart testnet
restart_testnet() {
    log_header "Restarting testnet"
    
    stop_testnet
    sleep 2
    start_testnet
}

# Function to clean all data
clean_all() {
    log_header "Cleaning testnet data"

    # Stop containers and remove volumes
    log_info "Stopping containers and removing volumes..."
    docker compose down -v 2>/dev/null || true

    # Remove data directories using Docker to handle permission issues
    log_info "Removing validator and sentry directories..."
    if [ -d "val-alice" ] || [ -d "val-bob" ] || [ -d "val-carol" ] || [ -d "sentry-alice" ] || [ -d "sentry-bob" ] || [ -d "sentry-carol" ]; then
        docker run --rm -v "$(pwd):/workspace" alpine:latest sh -c \
            "rm -rf /workspace/val-alice /workspace/val-bob /workspace/val-carol /workspace/sentry-alice /workspace/sentry-bob /workspace/sentry-carol" \
            2>/dev/null || true
    fi

    log_info "Clean complete"
}

# Function to show status
show_status() {
    log_header "Testnet Status"
    
    echo ""
    echo "Container Status:"
    docker ps --filter "name=val-" --filter "name=sentry-" --filter "name=ipfs" --format "table {{.Names}}\t{{.Status}}\t{{.State}}"
    
    echo ""
    echo "Network Endpoints (via Cloudflare Tunnel):"
    echo "  Alice RPC:  https://alice-rpc.sonr.land"
    echo "  Alice REST: https://alice-rest.sonr.land"
    echo "  Alice gRPC: https://alice-grpc.sonr.land"
    echo "  Alice EVM:  https://alice-evm.sonr.land"
    echo "  Bob RPC:    https://bob-rpc.sonr.land"
    echo "  Carol RPC:  https://carol-rpc.sonr.land"
    echo "  IPFS API:   https://ipfs-api.sonr.land"
    echo "  IPFS Gateway: https://ipfs-gateway.sonr.land"
    
    # Check sync status if services are running
    if docker ps | grep -q "sentry-alice"; then
        echo ""
        echo "Sync Status:"
        local block_height=$(docker exec sentry-alice sh -c 'curl -s http://localhost:26657/status 2>/dev/null | jq -r ".result.sync_info.latest_block_height" 2>/dev/null' || echo "0")
        echo "  Block Height: $block_height"
        for container in sentry-alice sentry-bob sentry-carol; do
            status=$(docker exec $container sh -c 'curl -s http://localhost:26657/status 2>/dev/null | jq -r ".result.sync_info.catching_up" 2>/dev/null' || echo "unavailable")
            echo "  $container: catching_up=$status"
        done
    fi

    echo ""
}

# Function to view logs
view_logs() {
    local service=$1
    
    if [ -z "$service" ]; then
        docker compose logs -f --tail=100
    else
        docker compose logs -f --tail=100 "$service"
    fi
}

# Function to execute command in container
exec_in_container() {
    local container=$1
    shift
    
    if [ -z "$container" ]; then
        log_error "Container name required"
        echo "Usage: ./bootstrap.sh exec <container> <command>"
        echo "Example: ./bootstrap.sh exec val-alice status"
        return 1
    fi
    
    docker exec -it "$container" snrd --home /root/.sonr "$@"
}

# Function to run tests
run_tests() {
    log_header "Running testnet tests"
    
    # Check if all services are running
    log_info "Checking service health..."
    local all_healthy=true
    
    for service in val-alice val-bob val-carol sentry-alice sentry-bob sentry-carol; do
        if ! docker ps | grep -q "$service"; then
            log_error "$service is not running"
            all_healthy=false
        fi
    done
    
    if [ "$all_healthy" = false ]; then
        log_error "Not all services are running"
        return 1
    fi
    
    # Test RPC endpoints
    log_info "Testing RPC endpoints..."
    for container in sentry-alice sentry-bob sentry-carol; do
        if docker exec $container sh -c 'curl -s http://localhost:26657/status' > /dev/null 2>&1; then
            log_info "$container: OK"
        else
            log_error "$container: FAILED"
        fi
    done
    
    # Test validator connectivity
    log_info "Testing validator connectivity..."
    local block_height=$(docker exec sentry-alice sh -c 'curl -s http://localhost:26657/status 2>/dev/null | jq -r ".result.sync_info.latest_block_height" 2>/dev/null' || echo "unavailable")
    if [ "$block_height" != "unavailable" ] && [ "$block_height" != "null" ]; then
        log_info "Current block height: $block_height"
    else
        log_error "Failed to get validator status"
    fi

    log_info "Tests complete"
}

# Function to show usage
show_usage() {
    echo "Sonr Testnet Bootstrap Script"
    echo ""
    echo "Usage: ./bootstrap.sh [command]"
    echo ""
    echo "Commands:"
    echo "  init       Initialize validators and sentries"
    echo "  start      Start the testnet"
    echo "  stop       Stop the testnet"
    echo "  restart    Restart the testnet"
    echo "  status     Show testnet status"
    echo "  logs       View logs (optional: service name)"
    echo "  clean      Clean all data and volumes"
    echo "  exec       Execute command in container"
    echo "  test       Run basic tests"
    echo "  help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./bootstrap.sh init"
    echo "  ./bootstrap.sh start"
    echo "  ./bootstrap.sh logs sentry-alice"
    echo "  ./bootstrap.sh exec val-alice status"
    echo ""
}

# Main command handler
main() {
    case "${1:-help}" in
        init)
            check_prerequisites
            init_validators
            ;;
        start)
            check_prerequisites
            start_testnet
            ;;
        stop)
            stop_testnet
            ;;
        restart)
            restart_testnet
            ;;
        status)
            show_status
            ;;
        logs)
            view_logs "${2:-}"
            ;;
        clean)
            clean_all
            ;;
        exec)
            shift
            exec_in_container "$@"
            ;;
        test)
            run_tests
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            log_error "Unknown command: $1"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"