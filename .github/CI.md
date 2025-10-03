# CI/CD Workflows

## Overview

Our CI/CD pipeline is optimized for speed and automation using:
- Self-hosted `builder` runner for better performance
- Smart scope detection via `github-env.sh`
- Automated versioning and releases via `devbox` commands

## Workflows

### PR CI (`pr.yml`)
**Trigger**: Pull requests
**Purpose**: Validate, test, and build changed components
**Features**:
- Only tests/builds affected scopes
- Runs linting first for fast feedback
- Cleans up artifacts after run

### Post-Merge Release (`merge.yml`)
**Trigger**: Push to master/main
**Purpose**: Automated version bumping and releases
**Features**:
- Auto-detects version increment from milestones
- Bumps versions, creates tags, and releases
- Publishes to package registries

### Nightly Snapshot (`nightly.yml`)
**Trigger**: Daily at 2 AM UTC or manual
**Purpose**: Build development snapshots
**Features**:
- Creates snapshot builds without version bumps
- Useful for testing bleeding-edge changes

### Manual Release (`manual-release.yml`)
**Trigger**: Manual workflow dispatch
**Purpose**: Override for emergency releases
**Features**:
- Select specific scope to release
- Choose version increment (patch/minor/major)
- Bypasses automatic detection

### CI Status (`ci-status.yml`)
**Trigger**: Other workflow completions
**Purpose**: Monitor CI health and report failures

## Key Commands

All workflows use these devbox commands that automatically detect changes:

```bash
devbox run test       # Tests only affected scopes
devbox run build      # Builds only affected scopes
devbox run bump       # Bumps versions for affected scopes
devbox run release    # Releases affected scopes
devbox run publish    # Publishes affected scopes
devbox run snapshot   # Creates snapshots for affected scopes
```

## Performance Optimizations

1. **Self-hosted Runner**: All workflows run on `builder` for:
   - Local dependency caching
   - No cold starts
   - Unlimited build time

2. **Smart Caching**: 
   - Go modules cached at `~/go/pkg/mod`
   - pnpm store cached at `~/.local/share/pnpm/store`
   - Devbox packages cached at `~/.devbox`

3. **Scope Detection**: Only test/build what changed using `github-env.sh`

4. **Cleanup Steps**: Prevent disk fill on self-hosted runner

## Adding New Components

1. Add scope to `.github/scopes.json`:
```json
{
  "name": "new-component",
  "include": ["path/to/component/**"]
}
```

2. Add scripts to `devbox.json`:
```json
"test:new-component": "make test-new-component",
"build:new-component": "make build-new-component",
"bump:new-component": "make -C path/to/component bump",
"release:new-component": "make -C path/to/component release"
```

3. That's it! CI automatically picks up the new scope.

## Secrets Required

- `GH_PAT_TOKEN`: GitHub Personal Access Token for pushing tags
- `GORELEASER_KEY`: GoReleaser Pro license key
- `NPM_TOKEN`: NPM registry authentication
- `CLOUDFLARE_API_TOKEN`: For deploying web apps
- `AWS_ACCESS_KEY_ID` & `AWS_SECRET_ACCESS_KEY`: For S3 uploads

## Troubleshooting

### Workflow not detecting changes?
- Check `.github/scopes.json` includes the correct paths
- Verify `github-env.sh` is executable
- Ensure full git history with `fetch-depth: 0`

### Self-hosted runner issues?
- Check runner has required dependencies
- Verify disk space for builds
- Check runner is online in Settings > Actions > Runners

### Release not triggering?
- Verify component has `bump:*` and `release:*` scripts
- Check git permissions with PAT token
- Ensure commitizen config exists for component