#!/bin/bash
# Setup script for running Sonr localnet on various systems
# Supports: Arch Linux, Ubuntu/Debian, RedHat/Fedora, macOS

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

print_color() {
    color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Detect OS and distribution
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/arch-release ]; then
            echo "arch"
        elif [ -f /etc/debian_version ]; then
            echo "debian"
        elif [ -f /etc/redhat-release ]; then
            echo "redhat"
        elif [ -f /etc/alpine-release ]; then
            echo "alpine"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

# Install Docker on different systems
install_docker() {
    local os=$1
    
    print_color $BLUE "Installing Docker for ${os}..."
    
    case $os in
        arch)
            print_color $YELLOW "Installing Docker on Arch Linux..."
            sudo pacman -Sy --noconfirm docker docker-compose
            sudo systemctl enable docker
            sudo systemctl start docker
            ;;
        debian)
            print_color $YELLOW "Installing Docker on Debian/Ubuntu..."
            sudo apt-get update
            sudo apt-get install -y docker.io docker-compose
            sudo systemctl enable docker
            sudo systemctl start docker
            ;;
        redhat)
            print_color $YELLOW "Installing Docker on RedHat/Fedora..."
            sudo dnf install -y docker docker-compose
            sudo systemctl enable docker
            sudo systemctl start docker
            ;;
        macos)
            print_color $YELLOW "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
            return 1
            ;;
        *)
            print_color $RED "Unsupported OS for automatic Docker installation"
            return 1
            ;;
    esac
    
    # Add user to docker group
    sudo usermod -aG docker $USER
    print_color $GREEN "Docker installed. You may need to log out and back in for group changes to take effect."
}

# Install Go if not present
install_go() {
    if command -v go >/dev/null 2>&1; then
        print_color $GREEN "Go is already installed: $(go version)"
        return 0
    fi
    
    print_color $BLUE "Installing Go..."
    
    local os=$(detect_os)
    local GO_VERSION="1.24.1"
    local ARCH=$(uname -m)
    
    # Map architecture
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_color $RED "Unsupported architecture: $ARCH"
            return 1
            ;;
    esac
    
    # Determine OS for Go download
    local GO_OS=""
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        GO_OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        GO_OS="darwin"
    else
        print_color $RED "Unsupported OS for Go installation"
        return 1
    fi
    
    # Download and install Go
    local GO_TAR="go${GO_VERSION}.${GO_OS}-${ARCH}.tar.gz"
    wget "https://go.dev/dl/${GO_TAR}" -O /tmp/${GO_TAR}
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/${GO_TAR}
    rm /tmp/${GO_TAR}
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
    
    # For zsh users
    if [ -f ~/.zshrc ]; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
        echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc
    fi
    
    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
    
    print_color $GREEN "Go ${GO_VERSION} installed successfully"
}

# Install dependencies
install_dependencies() {
    local os=$1
    
    print_color $BLUE "Installing dependencies for ${os}..."
    
    case $os in
        arch)
            sudo pacman -Sy --noconfirm base-devel git jq make gcc
            ;;
        debian)
            sudo apt-get update
            sudo apt-get install -y build-essential git jq make gcc
            ;;
        redhat)
            sudo dnf groupinstall -y "Development Tools"
            sudo dnf install -y git jq make gcc
            ;;
        macos)
            # Check for Homebrew
            if ! command -v brew >/dev/null 2>&1; then
                print_color $YELLOW "Installing Homebrew..."
                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            fi
            brew install jq make gcc
            ;;
        *)
            print_color $YELLOW "Please manually install: git, jq, make, gcc"
            ;;
    esac
}

# Setup function
setup_environment() {
    local os=$(detect_os)
    
    print_color $MAGENTA "=== Sonr Localnet Setup for ${os} ==="
    
    # Install dependencies
    install_dependencies $os
    
    # Install Go
    install_go
    
    # Docker setup (optional)
    print_color $BLUE "Would you like to install/configure Docker? (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        if command -v docker >/dev/null 2>&1; then
            print_color $GREEN "Docker is already installed"
            
            # Check if user is in docker group
            if ! groups | grep -q docker; then
                print_color $YELLOW "Adding user to docker group..."
                sudo usermod -aG docker $USER
                print_color $YELLOW "You'll need to log out and back in for this to take effect"
            fi
        else
            install_docker $os
        fi
    fi
    
    # Build the binary
    print_color $BLUE "Building Sonr binary..."
    make install
    
    # Create helpful aliases
    print_color $BLUE "Creating helpful aliases..."
    
    cat >> ~/.bashrc << 'EOF'

# Sonr aliases
alias sonr-localnet='cd $(pwd) && make localnet-x'
alias sonr-status='curl -s http://localhost:26657/status | jq'
alias sonr-logs='docker logs -f sonr-testnode 2>/dev/null || echo "No Docker container running"'
alias sonr-stop='docker stop sonr-testnode 2>/dev/null || pkill -f snrd'
EOF
    
    if [ -f ~/.zshrc ]; then
        cat >> ~/.zshrc << 'EOF'

# Sonr aliases
alias sonr-localnet='cd $(pwd) && make localnet-x'
alias sonr-status='curl -s http://localhost:26657/status | jq'
alias sonr-logs='docker logs -f sonr-testnode 2>/dev/null || echo "No Docker container running"'
alias sonr-stop='docker stop sonr-testnode 2>/dev/null || pkill -f snrd'
EOF
    fi
    
    print_color $GREEN "=== Setup Complete ==="
    print_color $YELLOW "Next steps:"
    print_color $YELLOW "1. If Docker was installed, log out and back in for group changes"
    print_color $YELLOW "2. Run 'make localnet-x' to start the cross-platform localnet"
    print_color $YELLOW "3. Or use the aliases: sonr-localnet, sonr-status, sonr-logs, sonr-stop"
    
    # OS-specific instructions
    case $os in
        arch)
            print_color $BLUE "\nArch Linux specific:"
            print_color $YELLOW "- If you prefer systemd service:"
            print_color $YELLOW "  sudo cp etc/systemd/sonr.service /etc/systemd/system/sonr@${USER}.service"
            print_color $YELLOW "  sudo systemctl daemon-reload"
            print_color $YELLOW "  sudo systemctl enable sonr@${USER}"
            print_color $YELLOW "  sudo systemctl start sonr@${USER}"
            ;;
    esac
}

# Main execution
main() {
    # Check if we're in the right directory
    if [ ! -f "Makefile" ] || [ ! -d "cmd/snrd" ]; then
        print_color $RED "Error: This script must be run from the Sonr repository root"
        exit 1
    fi
    
    setup_environment
}

# Run main
main "$@"