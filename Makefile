########################
### Makefile Helpers ###
########################

.PHONY: help
.DEFAULT_GOAL := help
help: ## Prints all the targets in all the Makefiles
	@echo ""
	@echo "\033[1;34m📋 go-dleq Makefile Targets\033[0m"
	@echo ""
	@echo "\033[1;34m=== 🔍 Information & Discovery ===\033[0m"
	@grep -h -E '^help:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== 🧪 Testing ===\033[0m"
	@grep -h -E '^test_(all|compatibility):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== ⚡ Benchmarking ===\033[0m"
	@grep -h -E '^benchmark_(all|report):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== 🔨 Building ===\033[0m"
	@grep -h -E '^(build_auto|build_fast|build_portable|clean_builds):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;34m=== 🧹 Code Quality ===\033[0m"
	@grep -h -E '^go_lint_and_format:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-58s\033[0m %s\n", $$1, $$2}'
	@echo ""

###############
### Testing ###
###############

.PHONY: test_all
test_all: ## Run all tests on both backends
	@echo "🧪 Running tests on all backends..."
	@echo ""
	@echo "\033[1;34m📊 Testing Decred Backend (Pure Go)\033[0m"
	@CGO_ENABLED=0 go test -v -race -count=1 ./...
	@echo ""
	@if command -v gcc >/dev/null 2>&1; then \
		echo "\033[1;34m📊 Testing Ethereum Backend (libsecp256k1)\033[0m"; \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -v -race -count=1 ./...; \
	else \
		echo "\033[1;33m⚠️  Skipping Ethereum backend tests (CGO not available)\033[0m"; \
	fi

.PHONY: test_compatibility
test_compatibility: ## Test backend compatibility and deterministic behavior
	@echo "🔬 Testing backend compatibility..."
	@go test -v -run TestBackendCompatibility -run TestCrossBackendResults

####################
### Benchmarking ###
####################

.PHONY: benchmark_all
benchmark_all: ## Run comprehensive benchmarks comparing both backends
	@echo "🔬 Running comprehensive go-dleq backend comparison..."
	@go run cmd/benchmark/main.go -compare -duration=3s

.PHONY: benchmark_report
benchmark_report: ## Generate a quick performance report
	@echo "🔬 Generating benchmark performance report..."
	@go run cmd/benchmark/main.go -report -duration=1s

#####################
### Build Targets ###
#####################

.PHONY: build_auto
build_auto: ## Auto-detect and build optimal backend based on environment
	@echo "🤖 Auto-detecting optimal backend..."
	@if command -v gcc >/dev/null 2>&1; then \
		echo "✨ CGO available - selecting Ethereum backend for maximum performance"; \
		$(MAKE) build_fast; \
	else \
		echo "🌍 CGO not available - selecting Decred backend for portability"; \
		$(MAKE) build_portable; \
	fi

.PHONY: build_fast
build_fast: ## Build with Ethereum backend (3x faster operations, requires CGO)
	@if ! command -v gcc >/dev/null 2>&1; then \
		echo "❌ CGO not available. Ethereum backend requires CGO and libsecp256k1."; \
		echo "   macOS:         brew install libsecp256k1"; \
		echo "   Ubuntu/Debian: sudo apt install libsecp256k1-dev"; \
		echo "   Alpine:        apk add libsecp256k1-dev"; \
		exit 1; \
	fi
	@echo "🚀 Building with Ethereum secp256k1 backend..."
	@echo "   • Requires CGO and libsecp256k1"
	@echo "   • ~3x faster scalar operations"
	@echo "   • ~3x faster DLEQ proofs"
	@echo "=================================================================="
	@CGO_ENABLED=1 go build -tags="ethereum_secp256k1" -o go-dleq-fast
	@echo "✅ Built: go-dleq-fast (Ethereum backend)"

.PHONY: build_portable
build_portable: ## Build with Decred backend (pure Go, maximum portability)
	@echo "🌍 Building with Decred secp256k1 backend..."
	@echo "   • Pure Go, no CGO dependencies"
	@echo "   • Excellent performance, maximum portability"
	@echo "   • Runs anywhere Go runs"
	@echo "=================================================================="
	@CGO_ENABLED=0 go build -o go-dleq-portable
	@echo "✅ Built: go-dleq-portable (Decred backend)"

.PHONY: clean_builds
clean_builds: ## Remove all built binaries and libraries
	@echo "🧹 Cleaning built binaries and libraries..."
	@rm -f go-dleq-fast go-dleq-portable
	@echo "✅ Cleaned all builds"

###############
### Linting ###
###############

.PHONY: go_lint_and_format
go_lint_and_format: ## Run golangci-lint on all Go files
	@echo "🧹 Formatting Go files..."
	@gofmt -s -w .
	@echo "✅ Go files formatted"
	@echo ""
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "🧹 Running golangci-lint..."; \
		golangci-lint run --timeout=5m; \
		echo "✅ Linting completed"; \
	else \
		echo "⚠️  golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi
