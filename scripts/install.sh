#!/usr/bin/env bash

set -e

# Function to detect OS and architecture
detect_platform() {
	OS=$(uname -s)
	ARCH=$(uname -m)

	# Normalize architecture names
	case "${ARCH}" in
	x86_64) ARCH="amd64" ;;
	aarch64 | arm64) ARCH="arm64" ;;
	esac
}

# Function to get latest release version
get_latest_version() {
	LATEST_VERSION=$(curl -s https://api.github.com/repos/sonr-io/snrd/releases/latest | grep "tag_name" | cut -d '"' -f 4)
	LATEST_VERSION=${LATEST_VERSION#v} # Remove 'v' prefix
}

# Function to install binaries to current directory
install_tar() {
	local OS_NAME=$1
	echo "Installing Sonr for ${OS_NAME} (${ARCH})..."
	DOWNLOAD_URL="https://github.com/sonr-io/snrd/releases/download/v${LATEST_VERSION}/sonr_${LATEST_VERSION}_${OS_NAME}_${ARCH}.tar.gz"

	# Download and extract
	echo "Downloading Sonr..."
	curl -L "${DOWNLOAD_URL}" -o sonr.tar.gz
	tar -xzf sonr.tar.gz
	rm sonr.tar.gz

	chmod +x sonrd hway

	echo "Binaries 'sonrd' and 'hway' have been extracted to the current directory"
	echo
	echo "To make them available system-wide, you can move them to /usr/local/bin with:"
	echo "sudo mv sonrd hway /usr/local/bin/"
	echo
	echo "Or move them to your personal bin directory with:"
	echo "mkdir -p ~/.local/bin"
	echo "mv sonrd hway ~/.local/bin/"
	echo "Then add ~/.local/bin to your PATH if it's not already there"
}

# Function to install on Debian/Ubuntu
install_debian() {
	echo "Installing Sonr for Debian/Ubuntu (${ARCH})..."
	SONRD_URL="https://github.com/sonr-io/snrd/releases/download/v${LATEST_VERSION}/sonrd_${LATEST_VERSION}_${ARCH}.deb"
	HWAY_URL="https://github.com/sonr-io/snrd/releases/download/v${LATEST_VERSION}/hway_${LATEST_VERSION}_${ARCH}.deb"

	# Download packages
	TMP_DIR=$(mktemp -d)
	curl -L "${SONRD_URL}" -o "${TMP_DIR}/sonrd.deb"
	curl -L "${HWAY_URL}" -o "${TMP_DIR}/hway.deb"

	# Install packages
	sudo dpkg -i "${TMP_DIR}/sonrd.deb"
	sudo dpkg -i "${TMP_DIR}/hway.deb"

	# Cleanup
	rm -rf "${TMP_DIR}"

	echo "Sonr has been installed system-wide"
}

main() {
	detect_platform
	get_latest_version

	case "${OS}" in
	Darwin)
		install_tar "Darwin"
		;;
	Linux)
		if [[ -f /etc/debian_version ]]; then
			install_debian
		else
			install_tar "Linux"
		fi
		;;
	*)
		echo "Unsupported operating system: ${OS}"
		exit 1
		;;
	esac
}

main
