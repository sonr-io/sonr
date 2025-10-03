#!/bin/bash

set -eu

# clear

# Detect git state and determine appropriate comparison base
git_comparison_base() {
  local current_branch=$(git rev-parse --abbrev-ref HEAD)

  # Check for uncommitted changes (staged or unstaged)
  if [[ -n "$(git status --porcelain)" ]]; then
    # Development mode - compare against last commit
    echo "HEAD"
    return
  fi

  # Check if we're on master/main branch
  if [[ "$current_branch" == "master" ]] || [[ "$current_branch" == "main" ]]; then
    # On master - compare against last tag
    local last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [[ -n "$last_tag" ]]; then
      echo "$last_tag"
    else
      # No tags exist, compare against first commit
      echo "$(git rev-list --max-parents=0 HEAD)"
    fi
    return
  fi

  # On feature branch with clean state - compare against master
  echo "master"
}

# Get current git context for display
git_context() {
  local base=$(git_comparison_base)
  local current_branch=$(git rev-parse --abbrev-ref HEAD)
  local has_changes=$(git status --porcelain | wc -l)

  if [[ "$has_changes" -gt 0 ]]; then
    echo "Development mode: comparing against HEAD (last commit)"
  elif [[ "$current_branch" == "master" ]] || [[ "$current_branch" == "main" ]]; then
    echo "Master branch: comparing against $base"
  else
    echo "Feature branch: comparing against master"
  fi
}

changed_files() {
  local base=$(git_comparison_base)

  if [[ "$base" == "HEAD" ]]; then
    # In development mode - show uncommitted changes
    git diff --name-only HEAD
    git diff --cached --name-only
    git ls-files --others --exclude-standard
  else
    # Compare against determined base
    git diff --name-only "$base" "$(git rev-parse HEAD)"
  fi | sort -u
}

changed_scopes() {
  local changed_files=$(changed_files)
  local context=$(git_context)

  echo "🔍 Changed Files by Scope"
  echo "========================="
  echo "📍 $context"
  echo ""

  # Process each scope
  yq eval '.[] | @json' .github/scopes.yml | while IFS= read -r scope_obj; do
    local scope_name=$(echo "$scope_obj" | jq -r '.name')
    local patterns=$(echo "$scope_obj" | jq -r '.include[]' | sed 's|^\./||; s|/\*\*$||')
    local matched_files=""

    # Find matching files for all patterns of this scope
    for pattern in $patterns; do
      while IFS= read -r file; do
        [[ -n "$file" ]] && [[ "$file" == "$pattern"* ]] && matched_files+="$file\n"
      done <<<"$changed_files"
    done

    # Display if there are matches
    if [[ -n "$matched_files" ]]; then
      local unique_count=$(echo -e "$matched_files" | sort -u | grep -c '^' || echo 0)
      echo ""
      echo "📦 $scope_name [$unique_count files]"

      # Group files by directory
      echo -e "$matched_files" | sort -u | while IFS= read -r file; do
        [[ -n "$file" ]] || continue

        # Indent based on directory depth
        local depth=$(echo "$file" | tr -cd '/' | wc -c)
        local indent=""
        for ((i = 0; i < depth; i++)); do
          indent="  $indent"
        done

        # Show just the filename with smart tree characters
        local basename=$(basename "$file")
        local dirname=$(dirname "$file")

        if [[ "$dirname" != "." ]]; then
          echo "  📁 $dirname/"
          echo "    └── $basename"
        else
          echo "  └── $file"
        fi
      done | sort -u # Remove duplicate directory listings
    fi
  done
}

success() {
  gum log --structured "$1" --prefix "✅" --prefix.foreground "#00ff00"
}

error() {
  gum log --structured "$1" --prefix "❌" --prefix.foreground "#ff0000"
}

info() {
  gum log --structured --prefix "📍" --prefix.foreground "#00ffff" "$1"
}

header() {
  local title="$1"
  gum style --border-foreground 240 --border double --align center --padding "0 1" "$title"
  echo "" | gum format
}

separator() {
  echo "---" | gum format
}

list_item() {
  local item="$1"
  echo "- $item" | gum format
}

devbox_run() {
  local command="$1"
  local action=$(echo "$command" | cut -f1 -d: | tr '[:upper:]' '[:lower:]')
  local scope=$(echo "$command" | cut -f2 -d: | tr '[:upper:]' '[:lower:]')
  local title="Executing devbox run $action:$scope..."
  gum spin --show-error --spinner meter --title "$title" -- devbox run "$command"
  success "$scope finished $action"
}

install() {
  header "Install"
  gum spin --show-error --spinner pulse --title "Installing pnpm..." -- pnpm install --frozen-lockfile
  gum spin --show-error --spinner pulse --title "Installing Go..." -- go mod download
  info "pnpm(10.14.0)"
  info "go(1.24.7)"
  separator
}

build_all() {
  header "Build"
  devbox_run "build:auth"
  devbox_run "build:dash"
  devbox_run "build:core"
  devbox_run "build:com"
  devbox_run "build:es"
  devbox_run "build:hway"
  devbox_run "build:motr"
  devbox_run "build:pkl"
  devbox_run "build:sdk"
  devbox_run "build:ui"
  devbox_run "build:vault"
  separator
}

test_all() {
  header "Test"
  devbox_run "test:auth"
  devbox_run "test:dash"
  devbox_run "test:core"
  devbox_run "test:com"
  devbox_run "test:es"
  devbox_run "test:hway"
  devbox_run "test:motr"
  devbox_run "test:pkl"
  devbox_run "test:sdk"
  devbox_run "test:ui"
  devbox_run "test:vault"
  separator
}

release_all() {
  header "Release"
  # devbox_run "release:auth"
  # devbox_run "release:dash"
  devbox_run "release:core"
  devbox_run "release:com"
  devbox_run "release:es"
  devbox_run "release:hway"
  devbox_run "release:motr"
  devbox_run "release:pkl"
  devbox_run "release:sdk"
  devbox_run "release:ui"
  devbox_run "release:vault"
  separator
}

snapshot_all() {
  header "Snapshot"
  devbox_run "snapshot:core"
  devbox_run "snapshot:hway"
  devbox_run "snapshot:motr"
  devbox_run "snapshot:vault"
  separator
}

test_scopes() {
  for scope in $(affected_scopes); do
    # Check if the test script exists in devbox.json
    if grep -q "\"test:$scope\":" devbox.json 2>/dev/null; then
      echo "Testing scope: $scope"
      devbox run "test:$scope"
    else
      echo "Skipping test for scope: $scope (no test script defined)"
    fi
  done
}

build_scopes() {
  for scope in $(affected_scopes); do
    # Check if the build script exists in devbox.json
    if grep -q "\"build:$scope\":" devbox.json 2>/dev/null; then
      echo "Building scope: $scope"
      devbox run "build:$scope"
    else
      echo "Skipping build for scope: $scope (no build script defined)"
    fi
  done
}

release_scopes() {
  for scope in $(affected_scopes); do
    # Check if the release script exists in devbox.json
    if grep -q "\"release:$scope\":" devbox.json 2>/dev/null; then
      echo "Releasing scope: $scope"
      devbox run "release:$scope"
    else
      echo "Skipping release for scope: $scope (no release script defined)"
    fi
  done
}

snapshot_scopes() {
  for scope in $(affected_scopes); do
    # Check if the snapshot script exists in devbox.json
    if grep -q "\"snapshot:$scope\":" devbox.json 2>/dev/null; then
      echo "Creating snapshot for scope: $scope"
      devbox run "snapshot:$scope"
    else
      echo "Skipping snapshot for scope: $scope (no snapshot script defined)"
    fi
  done
}

affected_scopes() {
  local verbose="${1:-false}"
  local changed_files=$(changed_files)

  if [[ "$verbose" == "true" ]] || [[ "$verbose" == "-v" ]]; then
    echo "📍 $(git_context)" >&2
    echo "" >&2
  fi

  yq eval -o=json '.[]' .github/scopes.yml | jq -r --arg files "$changed_files" '
    select(.include[] as $pattern |
      $files | split("\n")[] |
      startswith($pattern | sub("^\\./"; "") | sub("/\\*\\*$"; ""))
    ) |
    .name
  ' | sort -u
}

main() {
  cmd="$1"
  shift || true # Allow shifting even if no more args

  case "$cmd" in
  "build-all")
    build_all
    ;;
  "test-all")
    test_all
    ;;
  "release-all")
    release_all
    ;;
  "snapshot-all")
    snapshot_all
    ;;
  "build-scopes")
    build_scopes
    ;;
  "test-scopes")
    test_scopes
    ;;
  "release-scopes")
    release_scopes
    ;;
  "snapshot-scopes")
    snapshot_scopes
    ;;
  "install-pnpm")
    install_pnpm
    ;;
  "install-go")
    install_go
    ;;
  "install")
    install
    ;;
  *)
    echo "Unknown command: $cmd"
    exit 1
    ;;
  esac
}

main "$@"
