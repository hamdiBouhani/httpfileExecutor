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
run: ## Run the app with env vars
	OAUTH_CLIENT_ID="$(OAUTH_CLIENT_ID)" \
	OAUTH_CLIENT_SECRET="$(OAUTH_CLIENT_SECRET)" \
	OAUTH_TOKEN_URL="$(OAUTH_TOKEN_URL)" \
	API_TOKEN="$(API_TOKEN)" \
	HTTP_FILE="$(HTTP_FILE)" \
	$(GO) run $(MAIN_FILE)

# =============================================================================
# Test
# =============================================================================
.PHONY: test
test: ## Run all tests (verbose)
	$(GO) test -v ./...

.PHONY: test-short
test-short: ## Run tests in short mode (skips long-running ones)
	$(GO) test -short ./...

.PHONY: test-race
test-race: ## Run tests with the race detector
	$(GO) test -v -race ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report (HTML)
	@mkdir -p coverage
	$(GO) test -v -coverprofile=coverage/coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "✅ Coverage report: coverage/coverage.html"

.PHONY: test-clean
test-clean: ## Clear Go test cache
	$(GO) clean -testcache
	@echo "✅ Test cache cleared"

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

.PHONY: check
check: fmt vet test ## Run fmt, vet, and tests



# =============================================================================
# Cleanup
# =============================================================================
.PHONY: clean
clean: ## Remove build artifacts and coverage
	rm -rf bin/ coverage/
	$(GO) clean -cache -testcache 2>/dev/null || true
	@echo "✅ Cleaned"