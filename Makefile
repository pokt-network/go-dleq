########################
### Makefile Helpers ###
########################

# Include modular makefiles
include makefiles/benchmark.mk
include makefiles/build.mk
include makefiles/test.mk

.PHONY: prompt_user
# Internal helper target - prompt the user before continuing
prompt_user:
	@echo "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

.PHONY: help
.DEFAULT_GOAL := help
help: ## Prints all the targets in all the Makefiles
	@echo ""
	@echo "\033[1;34m📋 go-dleq Makefile Targets\033[0m"
	@echo ""
	@echo "\033[1;34m=== 🔍 Information & Discovery ===\033[0m"
	@grep -h -E '^(list|help):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== 🧪 Testing ===\033[0m"
	@grep -h -E '^test_.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== ⚡ Benchmarking ===\033[0m"
	@grep -h -E '^benchmark_.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== 🔨 Building ===\033[0m"
	@grep -h -E '^(build_.*|clean_builds):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== 🧹 Code Quality ===\033[0m"
	@grep -h -E '^(go_lint|go_format):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""

.PHONY: list
list: ## List all make targets
	@$(MAKE) -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$$@$$' | sort

###############
### Linting ###
###############

.PHONY: go_lint
go_lint: ## Run golangci-lint on all Go files
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "🧹 Running golangci-lint..."; \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: go_format
go_format: ## Format all Go files using gofmt
	@echo "🧹 Formatting Go files..."
	@gofmt -s -w .
	@echo "✅ Go files formatted"

################
### Shortcuts ###
################

.PHONY: quick
quick: build_auto test_quick benchmark_quick ## Quick development cycle: build, test, benchmark

.PHONY: full
full: go_format go_lint build_all test_all benchmark_all ## Full development cycle: format, lint, build all, test all, benchmark all

.PHONY: ci
ci: go_lint test_all benchmark_ci ## CI pipeline: lint, test, quick benchmark

.PHONY: dev
dev: build_auto test_all ## Development mode: auto-build and test

.PHONY: demo
demo: info backends benchmark_quick ## Demo the dual backend system

.PHONY: perf
perf: benchmark_dleq benchmark_parallel ## Focus on performance testing

.PHONY: compat
compat: test_compatibility build_all ## Test compatibility between backends

###################
### Information ###
###################

.PHONY: info
info: ## Show system and build information
	@echo "\033[1;34m📊 go-dleq Build Information\033[0m"
	@echo ""
	@echo "\033[1;32m=== System Info ===\033[0m"
	@echo "Go version:    $$(go version)"
	@echo "GOOS:          $$(go env GOOS)"
	@echo "GOARCH:        $$(go env GOARCH)"
	@echo "CGO_ENABLED:   $$(go env CGO_ENABLED)"
	@echo ""
	@echo "\033[1;32m=== Available Backends ===\033[0m"
	@echo "✅ Decred (Pure Go):      Always available"
	@if command -v gcc >/dev/null 2>&1; then \
		echo "✅ Ethereum (libsecp256k1): CGO available"; \
	else \
		echo "❌ Ethereum (libsecp256k1): CGO not available"; \
	fi
	@echo ""
	@echo "\033[1;32m=== Build Commands ===\033[0m"
	@echo "Default (Decred):  make build_portable"
	@echo "Fast (Ethereum):   make build_fast"
	@echo "Auto-select:       make build_auto"
	@echo ""
	@echo "\033[1;32m=== Dependencies ===\033[0m"
	@go list -m all | head -10
	@echo ""

.PHONY: backends
backends: ## Show available crypto backends and their status
	@echo "\033[1;34m🔐 Crypto Backend Status\033[0m"
	@echo ""
	@echo "\033[1;32m=== Decred Backend (Pure Go) ===\033[0m"
	@echo "Status:        ✅ Always available"
	@echo "Dependencies:  None (pure Go)"
	@echo "Performance:   Excellent baseline"
	@echo "Portability:   Maximum (works everywhere)"
	@echo "Build:         CGO_ENABLED=0 go build"
	@echo ""
	@echo "\033[1;32m=== Ethereum Backend (libsecp256k1) ===\033[0m"
	@if command -v gcc >/dev/null 2>&1; then \
		echo "Status:        ✅ Available (CGO enabled)"; \
	else \
		echo "Status:        ❌ Not available (no CGO)"; \
	fi
	@echo "Dependencies:  CGO + libsecp256k1"
	@echo "Performance:   🚀 3x faster operations"
	@echo "Portability:   Requires system libraries"
	@echo "Build:         CGO_ENABLED=1 go build -tags=ethereum_secp256k1"
	@echo ""
	@if ! command -v gcc >/dev/null 2>&1; then \
		echo "\033[1;33m💡 To enable Ethereum backend:\033[0m"; \
		echo "  macOS:         brew install libsecp256k1"; \
		echo "  Ubuntu/Debian: sudo apt install libsecp256k1-dev"; \
		echo "  Alpine:        apk add libsecp256k1-dev"; \
		echo ""; \
	fi