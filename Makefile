# =============================================================================
# http-executor - Makefile
# =============================================================================

# --- Configuration ---
BINARY_NAME   := http-executor
MAIN_FILE     := main.go
HTTP_FILE     ?= api_requests.http
GO            := go

# --- OAuth2 / Auth env vars (override on command line) ---
# Usage: make run OAUTH_CLIENT_ID=xxx OAUTH_CLIENT_SECRET=yyy
OAUTH_CLIENT_ID     ?= your-client-id
OAUTH_CLIENT_SECRET ?= your-client-secret
OAUTH_TOKEN_URL     ?= https://your-auth-server.com/oauth/token
API_TOKEN           ?=

# =============================================================================
# Default target
# =============================================================================
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show available targets
	@echo "http-executor - available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Setup
# =============================================================================
.PHONY: init
init: ## Initialize Go module and install dependencies
	$(GO) mod init http-executor
	$(GO) get github.com/gookit/greq/ext/httpfile
	$(GO) get golang.org/x/oauth2/clientcredentials
	@echo "✅ Module initialized and dependencies installed"

.PHONY: deps
deps: ## Download/tidy dependencies
	$(GO) mod download
	$(GO) mod tidy
	@echo "✅ Dependencies ready"

# =============================================================================
# Run
# =============================================================================
.PHONY: run
run: ## Run the app with env vars (default creds used unless overridden)
	OAUTH_CLIENT_ID="$(OAUTH_CLIENT_ID)" \
	OAUTH_CLIENT_SECRET="$(OAUTH_CLIENT_SECRET)" \
	OAUTH_TOKEN_URL="$(OAUTH_TOKEN_URL)" \
	API_TOKEN="$(API_TOKEN)" \
	HTTP_FILE="$(HTTP_FILE)" \
	$(GO) run $(MAIN_FILE)

# =============================================================================
# Build
# =============================================================================
.PHONY: build
build: ## Build the binary into ./bin
	@mkdir -p bin
	$(GO) build -o bin/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Built bin/$(BINARY_NAME)"

.PHONY: build-run
build-run: build ## Build then run the compiled binary
	OAUTH_CLIENT_ID="$(OAUTH_CLIENT_ID)" \
	OAUTH_CLIENT_SECRET="$(OAUTH_CLIENT_SECRET)" \
	OAUTH_TOKEN_URL="$(OAUTH_TOKEN_URL)" \
	API_TOKEN="$(API_TOKEN)" \
	HTTP_FILE="$(HTTP_FILE)" \
	./bin/$(BINARY_NAME)

# =============================================================================
# Quality
# =============================================================================
.PHONY: fmt
fmt: ## Format source
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	$(GO) mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/
	$(GO) clean -cache -testcache 2>/dev/null || true
	@echo "✅ Cleaned"