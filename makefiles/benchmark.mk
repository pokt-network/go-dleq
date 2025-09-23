####################
### Benchmarking ###
####################

.PHONY: benchmark_all
benchmark_all: ## Run comprehensive benchmarks comparing both backends
	@echo "🔬 Running comprehensive go-dleq backend comparison..."
	@./benchmark_runner.sh

.PHONY: benchmark_quick
benchmark_quick: ## Quick benchmark comparison (1s per test)
	@echo "🔬 Quick benchmark comparison..."
	@echo ""
	@echo "\033[1;34m📊 Decred Backend (Pure Go)\033[0m"
	@go test -bench=BenchmarkComparison -benchmem -run=^$$ -benchtime=1s 2>/dev/null | \
		grep "BenchmarkComparison" | \
		sed 's/BenchmarkComparison_/  /' | \
		awk '{printf "%-20s %8.0f μs  %6s %s  %6s %s\n", $$1, $$2/1000, $$3, $$4, $$5, $$6}'
	@echo ""
	@echo "\033[1;34m📊 Ethereum Backend (libsecp256k1)\033[0m"
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -bench=BenchmarkComparison -benchmem -run=^$$ -benchtime=1s 2>/dev/null | \
			grep "BenchmarkComparison" | \
			sed 's/BenchmarkComparison_/  /' | \
			awk '{printf "%-20s %8.0f μs  %6s %s  %6s %s\n", $$1, $$2/1000, $$3, $$4, $$5, $$6}'; \
	else \
		echo "  ❌ Not available (CGO required)"; \
	fi
	@echo ""

.PHONY: benchmark_decred
benchmark_decred: ## Benchmark Decred backend only
	@echo "🔬 Benchmarking Decred backend (Pure Go)..."
	@CGO_ENABLED=0 go test -bench=. -benchmem -run=^$$ -benchtime=3s

.PHONY: benchmark_ethereum
benchmark_ethereum: ## Benchmark Ethereum backend only (requires CGO)
	@if ! command -v gcc >/dev/null 2>&1; then \
		echo "❌ CGO not available. Ethereum backend requires CGO and libsecp256k1."; \
		echo "   Install: brew install libsecp256k1 (macOS) or apt install libsecp256k1-dev (Ubuntu)"; \
		exit 1; \
	fi
	@echo "🔬 Benchmarking Ethereum backend (libsecp256k1)..."
	@CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -bench=. -benchmem -run=^$$ -benchtime=3s

.PHONY: benchmark_ci
benchmark_ci: ## CI-friendly benchmark (shorter runtime)
	@echo "🔬 Running CI benchmarks..."
	@go test -bench=BenchmarkComparison_ScalarMul -benchmem -run=^$$ -benchtime=500ms
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -bench=BenchmarkComparison_ScalarMul -benchmem -run=^$$ -benchtime=500ms; \
	fi

.PHONY: benchmark_memory
benchmark_memory: ## Focus on memory allocation patterns
	@echo "🧠 Memory allocation analysis..."
	@echo ""
	@echo "Decred Backend:"
	@go test -bench=BenchmarkComparison_Memory -benchmem -run=^$$ -benchtime=1s | grep -E "(Memory|allocs)"
	@echo ""
	@if command -v gcc >/dev/null 2>&1; then \
		echo "Ethereum Backend:"; \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -bench=BenchmarkComparison_Memory -benchmem -run=^$$ -benchtime=1s | grep -E "(Memory|allocs)"; \
	fi

.PHONY: benchmark_parallel
benchmark_parallel: ## Test parallel performance scaling
	@echo "🚀 Parallel performance comparison..."
	@echo ""
	@echo "Decred Parallel ScalarMul:"
	@go test -bench=BenchmarkComparison_ParallelScalarMul -benchmem -run=^$$ -benchtime=2s
	@echo ""
	@if command -v gcc >/dev/null 2>&1; then \
		echo "Ethereum Parallel ScalarMul:"; \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -bench=BenchmarkComparison_ParallelScalarMul -benchmem -run=^$$ -benchtime=2s; \
	fi

.PHONY: benchmark_dleq
benchmark_dleq: ## Benchmark DLEQ operations specifically
	@echo "🔐 DLEQ proof performance comparison..."
	@echo ""
	@echo "Decred DLEQ Operations:"
	@go test -bench="BenchmarkComparison_DLEQ" -benchmem -run=^$$ -benchtime=1s
	@echo ""
	@if command -v gcc >/dev/null 2>&1; then \
		echo "Ethereum DLEQ Operations:"; \
		CGO_ENABLED=1 go test -tags=ethereum_secp256k1 -bench="BenchmarkComparison_DLEQ" -benchmem -run=^$$ -benchtime=1s; \
	fi