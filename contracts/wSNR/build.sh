#!/bin/bash
set -e

echo "🔧 Setting up Foundry environment..."

# Source foundry if installed via foundryup
if [ -f "$HOME/.foundry/bin/forge" ]; then
	export PATH="$HOME/.foundry/bin:$PATH"
fi

# Check if forge is available
if ! command -v forge &>/dev/null; then
	echo "❌ Forge not found in PATH. Please ensure Foundry is installed:"
	echo "   curl -L https://foundry.paradigm.xyz | bash"
	echo "   foundryup"
	echo ""
	echo "Then run: source ~/.bashrc or source ~/.zshrc"
	exit 1
fi

echo "✅ Forge found at: $(which forge)"
echo "📦 Installing OpenZeppelin contracts..."

# Install OpenZeppelin if not already installed
if [ ! -d "lib/openzeppelin-contracts" ]; then
	forge install OpenZeppelin/openzeppelin-contracts@v5.0.0 --no-commit
else
	echo "✅ OpenZeppelin already installed"
fi

echo "🏗️  Building contracts..."
forge build

echo "✅ Build complete!"
