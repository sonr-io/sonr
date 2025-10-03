#!/usr/bin/env bash

set -e

# Function to detect OS and architecture
detect_platform() {
	OS=$(uname -s)
	ARCH=$(uname -m)

	# Normalize OS names to match GitHub release naming
	case "${OS}" in
	Darwin) OS_NAME="darwin" ;;
	Linux) OS_NAME="linux" ;;
	MINGW* | MSYS* | CYGWIN*) OS_NAME="windows" ;;
	*)
		echo "Unsupported OS: ${OS}"
		exit 1
		;;
	esac

	# Normalize architecture names
	case "${ARCH}" in
	x86_64) ARCH="amd64" ;;
	aarch64 | arm64) ARCH="arm64" ;;
	*)
		echo "Unsupported architecture: ${ARCH}"
		exit 1
		;;
	esac
}

# Function to check if we can build locally
can_build_locally() {
	# Check if we're in the sonr project directory
	if [[ -f "Makefile" ]] && [[ -f "go.mod" ]] && grep -q "module github.com/sonr-io/sonr" go.mod 2>/dev/null; then
		return 0
	fi
	return 1
}

# Function to build locally
build_locally() {
	local INSTALL_DIR="$1"
	echo "Building Sonr locally..."

	# Check if make and go are available
	if ! command -v make >/dev/null 2>&1; then
		echo "Error: make is required to build locally"
		exit 1
	fi

	if ! command -v go >/dev/null 2>&1; then
		echo "Error: Go is required to build locally"
		exit 1
	fi

	# Build binaries
	echo "Building snrd..."
	make build || {
		echo "Error: Failed to build snrd"
		exit 1
	}

	echo "Building motr..."
	make motr || {
		echo "Error: Failed to build motr"
		exit 1
	}

	# Copy binaries to install directory
	cp build/snrd "${INSTALL_DIR}/" || {
		echo "Error: Failed to copy snrd"
		exit 1
	}
	cp build/motr.wasm "${INSTALL_DIR}/" || {
		echo "Error: Failed to copy motr.wasm"
		exit 1
	}

	# Make binaries executable
	chmod +x "${INSTALL_DIR}/snrd"

	echo "Binaries built and installed successfully to ${INSTALL_DIR}"
}

# Function to get latest release version
get_latest_version() {
	RELEASE_DATA=$(curl -s https://api.github.com/repos/sonr-io/sonr/releases/latest)

	# Check if API returned an error (no releases available or private repo)
	if echo "${RELEASE_DATA}" | grep -q '"message": "Not Found"'; then
		if can_build_locally; then
			echo "No public releases found, building locally..."
			return 1
		else
			echo "Error: No releases found for sonr-io/sonr"
			echo "Please build and install from source:"
			echo "  git clone https://github.com/sonr-io/sonr.git"
			echo "  cd sonr"
			echo "  ./scripts/install.sh"
			exit 1
		fi
	fi

	LATEST_VERSION=$(echo "${RELEASE_DATA}" | grep "tag_name" | cut -d '"' -f 4)
	if [[ -z "${LATEST_VERSION}" ]]; then
		echo "Error: Could not determine latest version"
		if can_build_locally; then
			echo "Falling back to local build..."
			return 1
		fi
		exit 1
	fi

	LATEST_VERSION=${LATEST_VERSION#v} # Remove 'v' prefix
	return 0
}

# Function to install binaries
install_binaries() {
	local INSTALL_DIR="${1:-$(pwd)}"

	# Check if we have a version from releases
	if [[ -n "${LATEST_VERSION}" ]]; then
		echo "Installing Sonr v${LATEST_VERSION} for ${OS_NAME} (${ARCH}) to ${INSTALL_DIR}..."

		# Use dl.sonr.io CDN for faster downloads
		BASE_URL="https://dl.sonr.io/v${LATEST_VERSION}"

		# Download snrd binary
		echo "Downloading snrd..."
		if [[ ${OS_NAME} == "windows" ]]; then
			echo "Error: snrd does not support Windows"
			exit 1
		else
			if ! curl -L "${BASE_URL}/snrd_${OS_NAME}_${ARCH}" -o "${INSTALL_DIR}/snrd" -f; then
				echo "Warning: Failed to download from CDN, trying GitHub releases..."
				# Fallback to GitHub releases
				GITHUB_URL="https://github.com/sonr-io/sonr/releases/download/v${LATEST_VERSION}"
				if ! curl -L "${GITHUB_URL}/snrd_${OS_NAME}_${ARCH}" -o "${INSTALL_DIR}/snrd" -f; then
					echo "Error: Failed to download snrd binary from both CDN and GitHub"
					exit 1
				fi
			fi
		fi

		# Download motr.wasm (platform independent)
		echo "Downloading motr.wasm..."
		if ! curl -L "${BASE_URL}/motr.wasm" -o "${INSTALL_DIR}/motr.wasm" -f; then
			echo "Warning: Failed to download from CDN, trying GitHub releases..."
			# Fallback to GitHub releases
			GITHUB_URL="https://github.com/sonr-io/sonr/releases/download/v${LATEST_VERSION}"
			if ! curl -L "${GITHUB_URL}/motr.wasm" -o "${INSTALL_DIR}/motr.wasm" -f; then
				echo "Error: Failed to download motr.wasm from both CDN and GitHub"
				exit 1
			fi
		fi

		# Make binaries executable
		chmod +x "${INSTALL_DIR}/snrd"

		echo "Binaries installed successfully to ${INSTALL_DIR}"
	else
		# Fall back to local build
		build_locally "${INSTALL_DIR}"
	fi

	echo
	echo "Available commands:"
	echo "  snrd       - Blockchain daemon"
	echo "  motr.wasm  - WebAssembly enclave"
	echo
	echo "Quick start:"
	echo "  snrd --help   # Show available commands"
	echo "  snrd start    # Start the blockchain node"
}

# Function to install system-wide
install_system() {
	if [[ ${EUID} -eq 0 ]]; then
		INSTALL_DIR="/usr/local/bin"
	else
		echo "Installing to /usr/local/bin requires sudo privileges..."
		sudo -v || {
			echo "Sudo access required for system-wide installation"
			exit 1
		}
		INSTALL_DIR="/usr/local/bin"

		# Create temporary directory and install there first
		TMP_DIR=$(mktemp -d)
		install_binaries "${TMP_DIR}"

		# Move to system directory with sudo
		sudo mv "${TMP_DIR}/"* "${INSTALL_DIR}/"
		rm -rf "${TMP_DIR}"

		echo "Binaries installed system-wide to ${INSTALL_DIR}"
		return
	fi

	install_binaries "${INSTALL_DIR}"
}

# Function to install to user directory
install_user() {
	USER_BIN_DIR="${HOME}/.local/bin"
	mkdir -p "${USER_BIN_DIR}"

	install_binaries "${USER_BIN_DIR}"

	# Check if user bin is in PATH
	if [[ ":${PATH}:" != *":${USER_BIN_DIR}:"* ]]; then
		echo
		echo "WARNING: ${USER_BIN_DIR} is not in your PATH"
		echo "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
		echo 'export PATH="$HOME/.local/bin:$PATH"'
		echo
		echo "Then restart your terminal or run: source ~/.bashrc"
	fi
}

main() {
	detect_platform

	# Try to get latest version from releases, but don't fail if not available
	if ! get_latest_version; then
		# get_latest_version returned 1, meaning no releases but we can build locally
		LATEST_VERSION=""
	fi

	# Parse command line arguments
	case "${1-}" in
	--system | -s)
		install_system
		;;
	--user | -u)
		install_user
		;;
	--current | -c)
		install_binaries "$(pwd)"
		;;
	--help | -h)
		echo "Sonr Installation Script"
		echo
		echo "Usage: $0 [OPTIONS]"
		echo
		echo "Options:"
		echo "  --system, -s    Install system-wide to /usr/local/bin (requires sudo)"
		echo "  --user, -u      Install to ~/.local/bin (user directory)"
		echo "  --current, -c   Install to current directory"
		echo "  --help, -h      Show this help message"
		echo
		echo "If no option is specified, defaults to --user installation"
		echo
		if can_build_locally; then
			echo "Note: No public releases found, will build from source"
		fi
		exit 0
		;;
	"")
		# Default to user installation
		install_user
		;;
	*)
		echo "Unknown option: $1"
		echo "Use --help for usage information"
		exit 1
		;;
	esac
}

main "$@"
