#!/bin/bash
set -e

# Cleanup function to discard CHANGELOG.md changes on failure
cleanup_on_error() {
  local exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    echo "❌ Error occurred during version bump (exit code: $exit_code)"

    # Find all CHANGELOG.md files with changes
    local changed_changelogs=$(git diff --name-only | grep "CHANGELOG.md" || true)
    local staged_changelogs=$(git diff --cached --name-only | grep "CHANGELOG.md" || true)

    # Discard unstaged changes
    if [[ -n "$changed_changelogs" ]]; then
      echo "🧹 Discarding local changes to CHANGELOG.md files..."
      while IFS= read -r file; do
        if [[ -n "$file" ]]; then
          echo "   Reverting: $file"
          git checkout -- "$file" 2>/dev/null || true
        fi
      done <<< "$changed_changelogs"
      echo "✅ Unstaged CHANGELOG.md changes discarded"
    fi

    # Discard staged changes
    if [[ -n "$staged_changelogs" ]]; then
      echo "🧹 Unstaging and discarding staged CHANGELOG.md files..."
      while IFS= read -r file; do
        if [[ -n "$file" ]]; then
          echo "   Reverting: $file"
          git reset HEAD "$file" 2>/dev/null || true
          git checkout -- "$file" 2>/dev/null || true
        fi
      done <<< "$staged_changelogs"
      echo "✅ Staged CHANGELOG.md changes discarded"
    fi

    if [[ -z "$changed_changelogs" ]] && [[ -z "$staged_changelogs" ]]; then
      echo "ℹ️  No CHANGELOG.md changes to discard"
    fi
  fi
  exit $exit_code
}

# Set trap to cleanup on any error
trap cleanup_on_error EXIT

# Get script directory and project root
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
NEW_TAG=$CZ_PRE_NEW_TAG_VERSION
IS_NEW=$CZ_PRE_IS_INITIAL
CHANGELOG=$CZ_PRE_CHANGELOG_FILE_NAME
ORIGIN_TAGS=$(git tag -l)

# Check if we're not on release branch
if [[ "$BRANCH" != "master" ]] && [[ "$BRANCH" != "main" ]]; then
  echo "❌ Error: Cannot bump versions on feature branch"
  echo "   Please switch to master/main branch and try again."
  exit 1
fi

# Verify next tag is not already present
if echo "$ORIGIN_TAGS" | grep -q "$NEW_TAG"; then
  echo "❌ Error: Tag $NEW_TAG already exists"
  echo "   Please choose a different tag name."
  exit 1
fi

# Create CHANGELOG.md if it doesn't exist
if [ ! -f "$CHANGELOG" ]; then
  echo "⚠️ Warning: CHANGELOG.md not found, creating new file."
  touch "$CHANGELOG"
fi
