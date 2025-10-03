#!/bin/bash
set -e

# Get script directory and project root
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Variables passed from commitizen
IS_INITIAL=$CZ_POST_IS_INITIAL
CURRENT_TAG=$CZ_POST_CURRENT_TAG_VERSION

# Check if we're not on release branch
if [[ "$BRANCH" != "master" ]] && [[ "$BRANCH" != "main" ]]; then
  echo "❌ Error: Cannot bump versions on feature branch"
  echo "   Please switch to master/main branch and try again."
  exit 1
fi

