#!/bin/sh

# cli-docgen.sh
# A shell script to generate CLI documentation by walking the installed snrd binary.
# This script assumes snrd is already installed and available in PATH.

set -e # Exit immediately if a command exits with a non-zero status.

# --- Default Values ---
OUT_DIR="./docs/cli"

# --- Usage Function ---
usage() {
    echo "Usage: $0 [--out <dir>] [--help]"
    echo
    echo "Options:"
    echo "  --out <dir>      Specify the output directory."
    echo "                   (Default: ./docs/cli)"
    echo "  --help           Display this help message."
    exit 1
}

# --- Argument Parsing ---
while [ "$#" -gt 0 ]; do
    case "$1" in
    --out)
        if [ -n "$2" ]; then
            OUT_DIR="$2"
            shift 2
        else
            echo "Error: --out requires a directory path." >&2
            usage
        fi
        ;;
    --help)
        usage
        ;;
    *)
        echo "Error: Unknown option '$1'" >&2
        usage
        ;;
    esac
done

# --- Main Logic ---

# Check if snrd binary is installed
if ! command -v snrd >/dev/null 2>&1; then
    echo "Error: snrd binary not found in PATH." >&2
    echo "Please run 'make install' first to install the snrd binary." >&2
    exit 1
fi

# Create output directory
echo "Creating output directory: $OUT_DIR"
mkdir -p "$OUT_DIR"

# Clean existing files
echo "Cleaning existing documentation..."
rm -f "$OUT_DIR"/*.md

# Function to sanitize filename
sanitize_filename() {
    echo "$1" | tr ' ' '_' | tr '/' '-' | tr -d '[](){}' | sed 's/^-//;s/-$//'
}

# Generate main command documentation
echo "Generating documentation for snrd..."

# Main snrd help
snrd --help >"$OUT_DIR/snrd.md" 2>&1 || true
echo "Generated: $OUT_DIR/snrd.md"

# Get list of main commands
echo "Walking command tree..."

# Primary commands we want to document
MAIN_COMMANDS="auth genesis help init keys migrate node prune query rollback status tendermint tx version"

# Generate documentation for each main command
for cmd in $MAIN_COMMANDS; do
    echo "Processing: snrd $cmd"

    # Generate main command doc
    filename=$(sanitize_filename "$cmd")
    snrd "$cmd" --help >"$OUT_DIR/snrd_${filename}.md" 2>&1 || true
    echo "  Generated: $OUT_DIR/snrd_${filename}.md"

    # Special handling for commonly used sub-commands
    case "$cmd" in
    "query")
        # Document query sub-modules
        QUERY_MODULES="account accounts auth bank consensus delegation did distribution dwn feegrant gov slashing staking svc tendermint tx"
        for module in $QUERY_MODULES; do
            echo "  Processing: snrd query $module"
            filename=$(sanitize_filename "query_${module}")
            snrd query "$module" --help >"$OUT_DIR/snrd_${filename}.md" 2>&1 || true
        done
        ;;
    "tx")
        # Document tx sub-modules
        TX_MODULES="bank consensus crisis did distribution dwn feegrant gov slashing staking svc"
        for module in $TX_MODULES; do
            echo "  Processing: snrd tx $module"
            filename=$(sanitize_filename "tx_${module}")
            snrd tx "$module" --help >"$OUT_DIR/snrd_${filename}.md" 2>&1 || true
        done
        ;;
    "keys")
        # Document keys sub-commands
        KEYS_COMMANDS="add delete export import list migrate parse show"
        for subcmd in $KEYS_COMMANDS; do
            echo "  Processing: snrd keys $subcmd"
            filename=$(sanitize_filename "keys_${subcmd}")
            snrd keys "$subcmd" --help >"$OUT_DIR/snrd_${filename}.md" 2>&1 || true
        done
        ;;
    "auth")
        # Document auth sub-commands
        AUTH_COMMANDS="register verify sign"
        for subcmd in $AUTH_COMMANDS; do
            echo "  Processing: snrd auth $subcmd"
            filename=$(sanitize_filename "auth_${subcmd}")
            snrd auth "$subcmd" --help >"$OUT_DIR/snrd_${filename}.md" 2>&1 || true
        done
        ;;
    "genesis")
        # Document genesis sub-commands
        GENESIS_COMMANDS="add-genesis-account collect-txs export gentx init migrate validate"
        for subcmd in $GENESIS_COMMANDS; do
            echo "  Processing: snrd genesis $subcmd"
            filename=$(sanitize_filename "genesis_${subcmd}")
            snrd genesis "$subcmd" --help >"$OUT_DIR/snrd_${filename}.md" 2>&1 || true
        done
        ;;
    esac
done

# Generate index file with all commands
echo "Generating index file..."
cat >"$OUT_DIR/index.md" <<EOF
# Sonr CLI Documentation

This documentation is auto-generated from the \`snrd\` binary.

## Main Commands

EOF

# Add links to all generated files
for file in "$OUT_DIR"/snrd*.md; do
    if [ -f "$file" ]; then
        basename=$(basename "$file" .md)
        # Convert filename back to readable format
        readable_name=$(echo "$basename" | sed 's/snrd_//' | tr '_' ' ' | sed 's/-/ /g')
        echo "- [$readable_name](./$basename.md)" >>"$OUT_DIR/index.md"
    fi
done

echo ""
echo "Documentation generation complete!"
echo "Files generated in: $OUT_DIR"
echo "Total files: $(ls -1 "$OUT_DIR"/*.md 2>/dev/null | wc -l)"
