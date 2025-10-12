#!/bin/bash

# scripts/lib/jq_patch.sh - JSON patching utilities for genesis and config files

set -euo pipefail

# Source environment helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/env.sh"

# Patch JSON file with jq expression and save to temp file then move back
# Usage: patch_json <file> <jq_expression>
patch_json() {
    local file="$1"
    local jq_expr="$2"

    ensure_file "$file"

    log_info "Patching $file with: $jq_expr"

    # Create temp file in same directory to avoid cross-filesystem issues
    local temp_file
    temp_file="$(dirname "$file")/.tmp.$(basename "$file").$$"

    if ! jq -r "$jq_expr" "$file" > "$temp_file" 2>/dev/null; then
        rm -f "$temp_file"
        log_error "Failed to patch JSON file: $file"
        return 1
    fi

    mv "$temp_file" "$file"
    log_success "Successfully patched $file"
}

# Ensure array contains specific value
# Usage: ensure_array_contains <file> <path> <value>
ensure_array_contains() {
    local file="$1"
    local path="$2"
    local value="$3"

    ensure_file "$file"

    local current_value
    current_value=$(jq -r "$path // []" "$file")

    # Check if value already exists
    if [[ "$current_value" == *"\"$value\""* ]]; then
        log_info "Value '$value' already exists in $path"
        return 0
    fi

    # Add value to array
    local jq_expr
    jq_expr=".${path} += [\"$value\"]"

    patch_json "$file" "$jq_expr"
}

# Set string value at path
# Usage: set_json_string <file> <path> <value>
set_json_string() {
    local file="$1"
    local path="$2"
    local value="$3"

    local jq_expr
    jq_expr=".${path} = \"$value\""

    patch_json "$file" "$jq_expr"
}

# Set numeric value at path
# Usage: set_json_number <file> <path> <value>
set_json_number() {
    local file="$1"
    local path="$2"
    local value="$3"

    local jq_expr
    jq_expr=".${path} = $value"

    patch_json "$file" "$jq_expr"
}

# Set boolean value at path
# Usage: set_json_bool <file> <path> <value>
set_json_bool() {
    local file="$1"
    local path="$2"
    local value="$3"

    local jq_expr
    jq_expr=".${path} = $value"

    patch_json "$file" "$jq_expr"
}

# Set object value at path
# Usage: set_json_object <file> <path> <json_object>
set_json_object() {
    local file="$1"
    local path="$2"
    local json_object="$3"

    local jq_expr
    jq_expr=".${path} = $json_object"

    patch_json "$file" "$jq_expr"
}

# Check if path exists and is not null
# Usage: json_path_exists <file> <path>
json_path_exists() {
    local file="$1"
    local path="$2"

    ensure_file "$file"

    local result
    result=$(jq -r "$path // empty" "$file" 2>/dev/null)

    [[ -n "$result" && "$result" != "null" ]]
}

# Get value at path
# Usage: get_json_value <file> <path>
get_json_value() {
    local file="$1"
    local path="$2"

    ensure_file "$file"

    jq -r "$path" "$file"
}

# Validate JSON file
# Usage: validate_json <file>
validate_json() {
    local file="$1"

    ensure_file "$file"

    if ! jq empty "$file" >/dev/null 2>&1; then
        log_error "Invalid JSON in file: $file"
        return 1
    fi

    log_success "JSON validation passed for $file"
}

# Export functions
export -f patch_json ensure_array_contains set_json_string set_json_number
export -f set_json_bool set_json_object json_path_exists get_json_value validate_json