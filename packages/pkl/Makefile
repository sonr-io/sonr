# Makefile for pkljar - Pkl configuration management
.PHONY: all release dist docker json-schema-to-pkl upload-module upload-release clean clean-out clean-dist clean-npx-cache

# Default target - run release and dist
all: release dist npm-publish clean-npx-cache

# Package Pkl files and upload to R2
release: clean-out
	@pkl project package packages/*/
	@gum log --level debug "Publishing to Cloudflare R2"
	@$(MAKE) upload-release
	@rm -rf .out

# Distribute Pkl files and upload to R2
dist: clean-dist
	@fd -e pkl . -t f -x sh -c 'mkdir -p dist/$$(dirname {}) && cp {} dist/$$(dirname {})/'
	@mv dist/packages/* dist/
	@rm -rf dist/basePklProject.pkl
	@rm -rf dist/packages
	@gum log --level debug "Publishing to Cloudflare R2"
	@$(MAKE) upload-module
	@rm -rf dist

# Interactive JSON Schema to Pkl conversion
json-schema-to-pkl:
	@SCHEMA=$$(gum file --file --header "Select a JSON schema file to convert to PKL") && \
	PACKAGE=package://pkg.pkl-lang.org/pkl-pantry/org.json_schema.contrib@1.1.4#/generate.pkl && \
	gum log --level debug "Converting $$SCHEMA to PKL" && \
	pkl eval $$PACKAGE -m . -p source=$$SCHEMA && \
	(gum confirm "Delete source file?" && rm $$SCHEMA || gum log --level debug "Skipping delete")

# Upload module to R2 (internal)
upload-module:
	@rclone copy dist/ r2:pklmod/

# Upload release to R2 (internal)
upload-release:
	@rclone copy .out/ r2:pkljar/

# Clean targets
clean-out:
	@rm -rf .out

clean-dist:
	@rm -rf dist

clean-npx-cache:
	@gum log --level debug "Clearing npx cache"
	@npx clear-npx-cache -y 2>/dev/null || true
	@gum log --level debug "NPX cache cleared"

clean: clean-out clean-dist
	@rm -rf build/

# Test npx functionality
npx-test:
	@gum log --level debug "Testing pkljar CLI via npm run"
	@npm start

# Publish to npm with version patch bump
npm-publish:
	@gum log --level debug "Bumping version patch"
	@npm version patch --no-git-tag-version
	@gum log --level debug "Publishing pkljar to npm"
	@npm publish

# Help target
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@echo "  all                 : Package and distribute Pkl files (default)"
	@echo "  release             : Package Pkl files and upload to R2"
	@echo "  dist                : Distribute Pkl files and upload to R2"
	@echo "  docker              : Build multi-platform Docker image"
	@echo "  json-schema-to-pkl  : Convert JSON schema to Pkl interactively"
	@echo "  npx-test            : Test pkljar CLI via npm run"
	@echo "  npm-publish         : Publish pkljar package to npm (with version bump)"
	@echo "  clean               : Remove all build artifacts"
	@echo "  clean-npx-cache     : Clear npx cache for testing fresh installs"
	@echo "  help                : Show this help message"

.PHONY: help npx-test npm-publish
