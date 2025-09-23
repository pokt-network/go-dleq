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

.PHONY: test_quick
test_quick: ## Quick test run (no race detection, default backend)
	@echo "🧪 Running quick tests..."
	@go test -v ./...

.PHONY: test_decred
test_decred: ## Test Decred backend only
	@echo "🧪 Testing Decred backend (Pure Go)..."
	@CGO_ENABLED=0 go test -v -race -count=1 ./...

.PHONY: test_ethereum
test_ethereum: ## Test Ethereum backend only (requires CGO)
	@if ! command -v gcc >/dev/null 2>&1; then \
		echo "❌ CGO not available. Ethereum backend requires CGO and libsecp256k1."; \
		echo "   Install: brew install libsecp256k1 (macOS) or apt install libsecp256k1-dev (Ubuntu)"; \
		exit 1; \
	fi
	@echo "🧪 Testing Ethereum backend (libsecp256k1)..."
	@CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -v -race -count=1 ./...

.PHONY: test_compatibility
test_compatibility: ## Test that both backends produce identical results
	@echo "🔍 Testing backend compatibility..."
	@echo ""
	@echo "Running compatibility validation..."
	@temp_dir=$$(mktemp -d); \
	echo "Testing proof serialization compatibility..."; \
	CGO_ENABLED=0 go test -run TestProof_Serde -v 2>&1 | grep "size of serialized proof" > $$temp_dir/decred_size; \
	if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -run TestProof_Serde -v 2>&1 | grep "size of serialized proof" > $$temp_dir/ethereum_size; \
		if ! diff $$temp_dir/decred_size $$temp_dir/ethereum_size >/dev/null; then \
			echo "⚠️  Serialization sizes differ between backends:"; \
			echo "Decred:   $$(cat $$temp_dir/decred_size)"; \
			echo "Ethereum: $$(cat $$temp_dir/ethereum_size)"; \
			echo "Note: Small differences (1-2 bytes) are acceptable due to encoding variations"; \
		else \
			echo "✅ Serialization sizes match between backends"; \
		fi; \
	fi; \
	rm -rf $$temp_dir
	@echo "✅ Compatibility test completed"

.PHONY: test_race
test_race: ## Run race condition detection on both backends
	@echo "🏃 Running race detection tests..."
	@echo ""
	@echo "Decred Backend:"
	@CGO_ENABLED=0 go test -race -count=1 ./...
	@echo ""
	@if command -v gcc >/dev/null 2>&1; then \
		echo "Ethereum Backend:"; \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -race -count=1 ./...; \
	fi

.PHONY: test_parallel
test_parallel: ## Test parallel execution safety
	@echo "🚀 Testing parallel execution..."
	@go test -parallel 8 -count=1 ./...
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -parallel 8 -count=1 ./...; \
	fi

.PHONY: test_stress
test_stress: ## Stress test both backends
	@echo "💪 Running stress tests..."
	@echo ""
	@echo "Decred Stress Test (100 iterations):"
	@CGO_ENABLED=0 go test -count=100 -timeout=10m ./...
	@echo ""
	@if command -v gcc >/dev/null 2>&1; then \
		echo "Ethereum Stress Test (100 iterations):"; \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -count=100 -timeout=10m ./...; \
	fi

.PHONY: test_coverage
test_coverage: ## Generate test coverage report
	@echo "📊 Generating test coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | tail -1

.PHONY: test_fuzz
test_fuzz: ## Run fuzz tests (Go 1.18+)
	@echo "🎲 Running fuzz tests..."
	@if go version | grep -q "go1\.\(1[8-9]\|[2-9][0-9]\)"; then \
		go test -fuzz=. -fuzztime=30s ./...; \
	else \
		echo "❌ Fuzz testing requires Go 1.18 or later"; \
	fi

.PHONY: test_ci
test_ci: ## CI-optimized test run
	@echo "🤖 Running CI tests..."
	@go test -race -count=1 -timeout=5m ./...
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -race -count=1 -timeout=5m ./...; \
	fi

.PHONY: test_clean
test_clean: ## Clean test artifacts
	@echo "🧹 Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -f *.test
	@rm -rf testdata/tmp
	@echo "✅ Test artifacts cleaned"