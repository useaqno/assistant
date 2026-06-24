# Aqno — developer workflow
# Run `make` (or `make help`) to see all commands.

SHELL := /bin/bash
PNPM  := pnpm
DAEMON_DIR := daemon

# App data directory (SQLite DB + downloaded voice models), per OS.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  DATA_DIR := $(HOME)/Library/Application Support/io.aqno
  APP_BUNDLE := src-tauri/target/release/bundle/macos/Aqno.app
else
  DATA_DIR := $(HOME)/.config/io.aqno
  APP_BUNDLE := src-tauri/target/release/aqno
endif

.DEFAULT_GOAL := help

# ----------------------------------------------------------------------------
# Help
# ----------------------------------------------------------------------------
.PHONY: help
help: ## Show this help
	@echo "Aqno — make targets:"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Dev hot reload: run 'make dev' (frontend+Rust HMR). For Go hot reload too,"
	@echo "use two terminals: 'make dev-daemon' + 'make dev-web'."

# ----------------------------------------------------------------------------
# Dependencies
# ----------------------------------------------------------------------------
.PHONY: install deps refresh
install: refresh ## Alias for 'refresh'
deps: refresh    ## Alias for 'refresh'
refresh: ## Install/refresh all dependencies (JS + Go)
	@echo "› installing JS dependencies (pnpm)…"
	$(PNPM) install
	@echo "› downloading Go modules…"
	cd $(DAEMON_DIR) && go mod download
	@echo "✓ dependencies ready"

# ----------------------------------------------------------------------------
# Development (hot reload)
# ----------------------------------------------------------------------------
.PHONY: dev dev-web dev-daemon sidecar
dev: sidecar ## Run the full desktop app with hot reload (Tauri + Vite HMR)
	$(PNPM) tauri dev

dev-web: ## Run only the SvelteKit frontend with HMR (browser, talks to :8787)
	$(PNPM) dev

dev-daemon: ## Run the Go daemon standalone with hot reload (uses 'air' if present)
	@if command -v air >/dev/null 2>&1; then \
	  echo "› air hot reload on :8787"; cd $(DAEMON_DIR) && air; \
	else \
	  echo "air not found — running once without hot reload."; \
	  echo "  install hot reload with: go install github.com/air-verse/air@latest"; \
	  cd $(DAEMON_DIR) && go run . ; \
	fi

sidecar: ## Build the Go daemon sidecar for the host platform
	$(PNPM) daemon:sidecar

# ----------------------------------------------------------------------------
# Production
# ----------------------------------------------------------------------------
.PHONY: build prod run start
build: prod      ## Alias for 'prod' (build the production bundle for this host)
prod: ## Build the production app bundle for the current platform
	$(PNPM) daemon:sidecar
	$(PNPM) tauri build

run: prod start  ## Build and launch the final (production) app
start: ## Launch the already-built production app
	@if [ "$(UNAME_S)" = "Darwin" ]; then \
	  open "$(APP_BUNDLE)"; \
	else \
	  "$(APP_BUNDLE)"; \
	fi

# ----------------------------------------------------------------------------
# Multi-platform builds
# Note: Tauri/Rust cannot cross-build between operating systems. macOS targets
# cross-build on macOS; Windows/Linux targets must run on that OS (or in CI).
# The Go sidecar (pure Go) cross-compiles for each target automatically.
# ----------------------------------------------------------------------------
.PHONY: build-all build-mac build-windows build-linux
build-all: ## Build every platform feasible on this machine (+ guidance)
	@if [ "$(UNAME_S)" = "Darwin" ]; then \
	  $(MAKE) build-mac; \
	  echo ""; \
	  echo "ℹ Windows/Linux bundles must be built on those OSes (or CI):"; \
	  echo "    make build-windows   # on Windows"; \
	  echo "    make build-linux     # on Linux"; \
	else \
	  echo "On this OS, build the native target:"; \
	  echo "    make build-linux   |   make build-windows"; \
	fi

build-mac: ## Build a universal macOS bundle (arm64 + x86_64)
	rustup target add aarch64-apple-darwin x86_64-apple-darwin
	AQNO_TARGET=aarch64-apple-darwin $(PNPM) daemon:sidecar
	AQNO_TARGET=x86_64-apple-darwin  $(PNPM) daemon:sidecar
	$(PNPM) tauri build --target universal-apple-darwin

build-windows: ## Build the Windows bundle (run on Windows)
	rustup target add x86_64-pc-windows-msvc
	AQNO_TARGET=x86_64-pc-windows-msvc $(PNPM) daemon:sidecar
	$(PNPM) tauri build --target x86_64-pc-windows-msvc

build-linux: ## Build the Linux bundle (run on Linux)
	rustup target add x86_64-unknown-linux-gnu
	AQNO_TARGET=x86_64-unknown-linux-gnu $(PNPM) daemon:sidecar
	$(PNPM) tauri build --target x86_64-unknown-linux-gnu

# ----------------------------------------------------------------------------
# Quality
# ----------------------------------------------------------------------------
.PHONY: check lint format test
check: ## Type-check the frontend + vet the daemon
	$(PNPM) check
	cd $(DAEMON_DIR) && go vet ./...

lint: ## Lint the frontend
	$(PNPM) lint

format: ## Format JS/Svelte (prettier) and Go (gofmt)
	$(PNPM) format
	cd $(DAEMON_DIR) && gofmt -w .

test: ## Run the Go test suite
	cd $(DAEMON_DIR) && go test ./...

# ----------------------------------------------------------------------------
# Cleanup
# ----------------------------------------------------------------------------
.PHONY: clean reset-db reset
clean: ## Remove build artifacts (keeps deps and the database)
	rm -rf build .svelte-kit
	rm -rf src-tauri/target
	rm -f src-tauri/binaries/aqnod*
	cd $(DAEMON_DIR) && go clean -cache
	@echo "✓ build artifacts removed"

reset-db: ## Delete only the local database + downloaded voice models
	rm -rf "$(DATA_DIR)"
	@echo "✓ removed $(DATA_DIR)"

reset: clean reset-db ## Full reset: artifacts, node_modules, Go module cache, and the database
	rm -rf node_modules
	@echo "› clearing the Go module cache (global)…"
	-go clean -modcache
	@echo "✓ full reset done — run 'make refresh' to reinstall"
