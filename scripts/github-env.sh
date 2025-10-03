#!/bin/bash

set -eu

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

# Get development mode status
is_development_mode() {
  if [[ -n "$(git status --porcelain)" ]]; then
    echo "true"
  else
    echo "false"
  fi
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

default_branch() {
  # This function is kept for backward compatibility
  echo "master"
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

changed_scopes_diff() {
  local changed_files=$(changed_files)
  local base_branch=$(default_branch)
  local max_lines="${1:-10}" # Max diff lines to show per file

  echo "🔍 Changed Files with Diffs"
  echo "==========================="

  # Process each scope
  yq eval '.[] | @json' .github/scopes.yml | while IFS= read -r scope_obj; do
    local scope_name=$(echo "$scope_obj" | jq -r '.name')
    local patterns=$(echo "$scope_obj" | jq -r '.include[]' | sed 's|^\./||; s|/\*\*$||')
    local matched_files=""

    # Find matching files
    for pattern in $patterns; do
      while IFS= read -r file; do
        [[ -n "$file" ]] && [[ "$file" == "$pattern"* ]] && matched_files+="$file\n"
      done <<<"$changed_files"
    done

    if [[ -n "$matched_files" ]]; then
      echo ""
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      echo "📦 $scope_name"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

      echo -e "$matched_files" | sort -u | while IFS= read -r file; do
        [[ -n "$file" ]] || continue

        echo ""
        echo "┌─ 📄 $file"
        echo "├─────────────────────────────────"

        # Show the diff with context
        git diff --unified=2 "$base_branch...HEAD" -- "$file" |
          tail -n +5 |
          head -n "$max_lines" |
          while IFS= read -r line; do
            case "$line" in
            @@*)
              echo "│ $(tput setaf 6)$line$(tput sgr0)"
              ;;
            +*)
              echo "│ $(tput setaf 2)$line$(tput sgr0)"
              ;;
            -*)
              echo "│ $(tput setaf 1)$line$(tput sgr0)"
              ;;
            *)
              echo "│ $line"
              ;;
            esac
          done

        # Check if diff was truncated
        local full_diff_lines=$(git diff "$base_branch...HEAD" -- "$file" | wc -l)
        if [[ $full_diff_lines -gt $max_lines ]]; then
          echo "│ ... ($(($full_diff_lines - $max_lines)) more lines)"
        fi

        echo "└─────────────────────────────────"
      done
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

current_issue() {
  local branch=$(git rev-parse --abbrev-ref HEAD)
  local feature_part=$(echo "$branch" | cut -f2 -d'/')
  local result=$(gh issue ls | rg -i "$feature_part" | cut -f1 | head -n1)
  if [[ -z "$result" ]]; then
    local middle_part=$(echo "$feature_part" | cut -f2 -d'-')
    if [[ -n "$middle_part" ]]; then
      result=$(gh issue ls | rg -i "$middle_part" | cut -f1 | head -n1)
    fi
  fi
  echo "$result"
}

current_branch() {
  git rev-parse --abbrev-ref HEAD
}

current_pr() {
  gh pr ls --json headRefName,number --jq '.[] | {number: .number, branch: .headRefName}' | rg "$(git rev-parse --abbrev-ref HEAD)" | jq '.number'
}

current_milestone() {
  gh issue view "$(current_issue)" --json milestone --jq '.milestone'
}

final_issue() {
  local issue_num=$(current_issue)

  if [[ -z "$issue_num" ]]; then
    echo "false"
    return
  fi

  # Get milestone details
  local milestone_data=$(gh issue view "$issue_num" --json milestone)
  local milestone_title=$(echo "$milestone_data" | jq -r '.milestone.title // empty')
  local milestone_number=$(echo "$milestone_data" | jq -r '.milestone.number // empty')

  if [[ -z "$milestone_title" ]]; then
    echo "false"
    return
  fi

  # Get all open issues in the milestone
  local open_issues=$(gh issue list --milestone "$milestone_title" --state open --json number,title)
  local open_count=$(echo "$open_issues" | jq 'length')

  # Debug output (optional - remove if not needed)
  if [[ "${DEBUG:-}" == "true" ]]; then
    echo "Current issue: #$issue_num" >&2
    echo "Milestone: $milestone_title" >&2
    echo "Open issues in milestone: $open_count" >&2
    echo "$open_issues" | jq -r '.[] | "#\(.number): \(.title)"' >&2
  fi

  # Return true only if exactly 1 issue remains
  [[ "$open_count" -eq 1 ]] && echo "true" || echo "false"
}

# Validate that we're on the default branch (master/main)
# Returns exit 1 if not on default branch - useful for pre-bump hooks
validate_default_branch() {
  local current_branch=$(git rev-parse --abbrev-ref HEAD)

  if [[ "$current_branch" == "master" ]] || [[ "$current_branch" == "main" ]]; then
    echo "✅ On default branch: $current_branch"
    return 0
  else
    echo "❌ Error: Not on default branch (master/main)"
    echo "   Current branch: $current_branch"
    echo "   Please switch to master/main before bumping versions"
    return 1
  fi
}

main() {
  cmd="$1"
  shift || true # Allow shifting even if no more args

  case "$cmd" in
  "git-context")
    echo "$(git_context)"
    ;;
  "validate-default-branch")
    validate_default_branch
    exit $?
    ;;
  "affected-scopes")
    echo "$(affected_scopes "$@")"
    ;;
  "changed-files")
    echo "$(changed_files)"
    ;;
  "changed-scopes")
    echo "$(changed_scopes)"
    ;;
  "changed-scopes-diff")
    echo "$(changed_scopes_diff)"
    ;;
  "current-branch")
    echo "$(current_branch)"
    ;;
  "current-issue")
    echo "$(current_issue)"
    ;;
  "current-milestone")
    echo "$(current_milestone)"
    ;;
  "current-pr")
    echo "$(current_pr)"
    ;;
  "final-issue")
    echo "$(final_issue)"
    ;;
  *)
    echo "Unknown command: $cmd"
    exit 1
    ;;
  esac
}

main "$@"
