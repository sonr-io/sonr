#!/bin/bash
# Cross-platform localnet script that works on all systems including Arch Linux
# This script handles Docker permission issues and provides fallback to local binary

set -eu

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/arch-release ]; then
            echo "arch"
        elif [ -f /etc/debian_version ]; then
            echo "debian"
        elif [ -f /etc/redhat-release ]; then
            echo "redhat"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

# Check if user is in docker group
check_docker_access() {
    if command -v docker >/dev/null 2>&1; then
        # Try to run a simple docker command
        if docker info >/dev/null 2>&1; then
            return 0
        else
            # Check if it's a permission issue
            if groups | grep -q docker; then
                print_color "$YELLOW" "You're in the docker group but Docker daemon might not be running"
                return 1
            else
                print_color "$YELLOW" "You need to be in the docker group to use Docker without sudo"
                return 1
            fi
        fi
    else
        return 1
    fi
}

# Check if binary exists and is accessible
check_binary() {
    local binary_name="${1:-snrd}"

    # Check in build directory first
    if [ -x "./build/${binary_name}" ]; then
        echo "./build/${binary_name}"
        return 0
    fi

    # Check in PATH
    if command -v "${binary_name}" >/dev/null 2>&1; then
        echo "$(which "${binary_name}")"
        return 0
    fi

    return 1
}

# Setup proper permissions for directories
setup_permissions() {
    local home_dir="${1}"

    # Ensure directory exists
    mkdir -p "${home_dir}"

    # Set proper ownership if running as non-root
    if [ "$EUID" -ne 0 ]; then
        # Make sure current user owns the directory
        if [ -w "${home_dir}" ]; then
            return 0
        else
            print_color "$YELLOW" "Setting up permissions for ${home_dir}..."
            # Try to take ownership (will fail if not owner)
            if ! chmod -R u+rwX "${home_dir}" 2>/dev/null; then
                print_color "$RED" "Cannot set permissions on ${home_dir}. You may need to remove it manually: sudo rm -rf ${home_dir}"
                return 1
            fi
        fi
    fi

    return 0
}

# Main execution
main() {
    print_color "$BLUE" "=== Sonr Cross-Platform Localnet Setup ==="

    # Detect OS
    OS=$(detect_os)
    print_color "$GREEN" "Detected OS: ${OS}"

    # Set default values
    export CHAIN_ID=${CHAIN_ID:-"sonrtest_1-1"}
    export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.sonr"}")
    export BINARY=${BINARY:-"snrd"}
    export BLOCK_TIME=${BLOCK_TIME:-"1000ms"}
    export CLEAN=${CLEAN:-"true"}

    # Decision logic for execution method
    USE_METHOD=""
    BINARY_PATH=""

    # 1. Check if FORCE_DOCKER is set
    if [[ "${FORCE_DOCKER:-false}" == "true" ]]; then
        if check_docker_access; then
            USE_METHOD="docker"
            print_color "$GREEN" "Using Docker (forced)"
        else
            print_color "$RED" "FORCE_DOCKER=true but Docker is not accessible"
            exit 1
        fi
    # 2. Check for local binary
    elif BINARY_PATH=$(check_binary "${BINARY}"); then
        USE_METHOD="local"
        print_color "$GREEN" "Using local binary: ${BINARY_PATH}"
    # 3. Try Docker as fallback
    elif check_docker_access; then
        USE_METHOD="docker"
        print_color "$YELLOW" "No local binary found, using Docker"
    else
        # No method available
        print_color "$RED" "Error: No execution method available!"
        print_color "$YELLOW" "Please either:"
        print_color "$YELLOW" "  1. Run 'make install' to build the binary"
        print_color "$YELLOW" "  2. Install Docker and ensure it's running"
        if [[ "${OS}" == "arch" ]]; then
            print_color "$YELLOW" "  On Arch Linux: sudo pacman -S docker && sudo systemctl start docker"
            print_color "$YELLOW" "  Add yourself to docker group: sudo usermod -aG docker $USER"
        fi
        exit 1
    fi

    # Setup permissions for home directory
    if ! setup_permissions "${HOME_DIR}"; then
        print_color "$RED" "Failed to setup permissions"
        exit 1
    fi

    # Export the method for test_node.sh to use
    if [[ "${USE_METHOD}" == "docker" ]]; then
        export FORCE_DOCKER=true
    else
        export FORCE_DOCKER=false
        # Add build directory to PATH if using local binary from build/
        if [[ "${BINARY_PATH}" == "./build/"* ]]; then
            export PATH="$(pwd)/build:${PATH}"
        fi
    fi

    # Run the actual test node script
    print_color "$BLUE" "Starting localnet with ${USE_METHOD} method..."
    bash scripts/test_node.sh
}

# Run main function
main "$@"
