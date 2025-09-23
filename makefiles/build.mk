#####################
### Build Targets ###
#####################

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

.PHONY: build_auto
build_auto: ## Auto-select optimal backend (Ethereum if CGO available, otherwise Decred)
	@echo "🎯 Auto-selecting optimal crypto backend..."
	@if command -v gcc >/dev/null 2>&1 && [ "$${CGO_ENABLED:-1}" != "0" ]; then \
		echo "   • CGO available, building fast version..."; \
		$(MAKE) build_fast; \
	else \
		echo "   • No CGO or CGO disabled, building portable version..."; \
		$(MAKE) build_portable; \
	fi

.PHONY: build_all
build_all: ## Build both Ethereum (fast) and Decred (portable) versions
	@echo "🏗️  Building all go-dleq variants..."
	@$(MAKE) build_portable
	@$(MAKE) build_fast
	@echo "=================================================================="
	@echo "✅ Built all variants:"
	@echo "   • go-dleq-portable  (Decred backend - pure Go)"
	@echo "   • go-dleq-fast      (Ethereum backend - CGO)"
	@ls -la go-dleq-* 2>/dev/null || true

.PHONY: build_library
build_library: ## Build as library (no main binary)
	@echo "📚 Building go-dleq as library..."
	@go build -buildmode=archive -o libgo-dleq.a
	@echo "✅ Built: libgo-dleq.a"

.PHONY: build_shared
build_shared: ## Build as shared library (requires CGO)
	@if ! command -v gcc >/dev/null 2>&1; then \
		echo "❌ CGO required for shared library build"; \
		exit 1; \
	fi
	@echo "🔗 Building go-dleq as shared library..."
	@CGO_ENABLED=1 go build -buildmode=c-shared -o libgo-dleq.so
	@echo "✅ Built: libgo-dleq.so and libgo-dleq.h"

.PHONY: build_wasm
build_wasm: ## Build for WebAssembly (Decred backend only)
	@echo "🌐 Building go-dleq for WebAssembly..."
	@echo "   • Using Decred backend (pure Go)"
	@echo "   • Ethernet backend not supported in WASM"
	@GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -o go-dleq.wasm
	@echo "✅ Built: go-dleq.wasm"

.PHONY: clean_builds
clean_builds: ## Remove all built binaries and libraries
	@echo "🧹 Cleaning built binaries and libraries..."
	@rm -f go-dleq-fast go-dleq-portable go-dleq.wasm
	@rm -f libgo-dleq.a libgo-dleq.so libgo-dleq.h
	@echo "✅ Cleaned all builds"

.PHONY: install_deps
install_deps: ## Install build dependencies for all backends
	@echo "📦 Installing build dependencies..."
	@echo ""
	@if command -v brew >/dev/null 2>&1; then \
		echo "🍺 Installing via Homebrew (macOS):"; \
		brew install libsecp256k1 || echo "   libsecp256k1 already installed"; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "📦 Installing via apt (Ubuntu/Debian):"; \
		sudo apt-get update && sudo apt-get install -y libsecp256k1-dev; \
	elif command -v apk >/dev/null 2>&1; then \
		echo "🏔️  Installing via apk (Alpine):"; \
		apk add --no-cache libsecp256k1-dev; \
	else \
		echo "❌ Unsupported package manager. Please install libsecp256k1 manually."; \
		echo "   See: https://github.com/bitcoin-core/secp256k1"; \
	fi
	@echo ""
	@echo "✅ Dependencies installed (if supported)"

.PHONY: cross_compile
cross_compile: ## Cross-compile for multiple platforms (Decred backend only)
	@echo "🌍 Cross-compiling go-dleq for multiple platforms..."
	@echo "   • Using Decred backend (CGO not supported in cross-compilation)"
	@echo ""
	@mkdir -p dist
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ] && [ "$$arch" = "arm64" ]; then continue; fi; \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o dist/go-dleq-$$os-$$arch; \
			if [ "$$os" = "windows" ]; then \
				mv dist/go-dleq-$$os-$$arch dist/go-dleq-$$os-$$arch.exe; \
			fi; \
		done; \
	done
	@echo ""
	@echo "✅ Cross-compilation complete:"
	@ls -la dist/