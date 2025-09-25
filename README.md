# go-dleq <!-- omit in toc -->

Cross-group discrete logarithm equality implementation per [MRL-0010](https://www.getmonero.org/resources/research-lab/pubs/MRL-0010.pdf) with **dual secp256k1 backends** for 3x performance gains.

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Performance](#performance)
  - [Apple M1 Max Results](#apple-m1-max-results)
  - [Memory Usage](#memory-usage)
  - [Run Your Own Benchmarks](#run-your-own-benchmarks)
- [Backend Selection](#backend-selection)
  - [Build Commands](#build-commands)
  - [Installation](#installation)
  - [When to Use Each Backend](#when-to-use-each-backend)
- [API Usage](#api-usage)
- [Technical Details](#technical-details)
  - [Architecture](#architecture)
  - [Optimizations](#optimizations)
  - [Compatibility Testing](#compatibility-testing)
  - [Performance TODOs](#performance-todos)
  - [Shannon SDK Impact](#shannon-sdk-impact)

## Overview

Implementation of cross-group discrete logarithm equality with proof of knowledge signatures on both curves. Supports secp256k1 and ed25519.

**Key Feature:** Pluggable secp256k1 backends - choose between portability (pure Go) or **3x performance** (libsecp256k1).

## Quick Start

```bash
# Auto-detect and build optimal backend
make build_auto

# Run benchmarks
make benchmark_all

# Run tests (both backends)
make test_all

# Test cross-backend compatibility
make test_compatibility
```

## Performance

| Operation        | Decred (Pure Go) | Ethereum (libsecp256k1) | **Improvement** |
| ---------------- | ---------------- | ----------------------- | --------------- |
| **ScalarMul**    | 125 μs           | 43 μs                   | **3x faster**   |
| **ECDSA Sign**   | 93 μs            | 36 μs                   | **2.6x faster** |
| **ECDSA Verify** | 212 μs           | 42 μs                   | **5x faster**   |
| **DLEQ Proof**   | 485 ms           | 157 ms                  | **3x faster**   |

<details>
<summary><b>📊 Full Benchmark Results</b></summary>

### Apple M1 Max Results

| Operation               | Decred | Ethereum | Improvement |
| ----------------------- | ------ | -------- | ----------- |
| ScalarBaseMul           | 36 μs  | 43 μs    | Similar     |
| ScalarMul               | 125 μs | 43 μs    | 3.0x faster |
| Sign                    | 93 μs  | 36 μs    | 2.6x faster |
| Verify                  | 212 μs | 42 μs    | 5.0x faster |
| DLEQ Proof Generation   | 485 ms | 157 ms   | 3.1x faster |
| DLEQ Proof Verification | 413 ms | 131 ms   | 3.2x faster |
| Parallel ScalarMul      | 18 μs  | 6 μs     | 3.0x faster |

### Memory Usage

- **Decred:** 136 B/op, 2 allocations
- **Ethereum:** 328 B/op, 8 allocations (optimized with sync.Pool)

### Run Your Own Benchmarks

```bash
make benchmark_all              # Full comparison
go run cmd/benchmark/main.go -compare -duration=10s
```

</details>

## Backend Selection

### Build Commands

```bash
# Pure Go (Default) - Maximum portability
CGO_ENABLED=0 go build

# High Performance - Requires libsecp256k1
CGO_ENABLED=1 go build -tags="ethereum_secp256k1"

# Auto-select optimal backend
make build_auto
```

### Installation

<details>
<summary><b>📦 Installing libsecp256k1</b></summary>

**macOS:**

```bash
brew install libsecp256k1
```

**Ubuntu/Debian:**

```bash
sudo apt install libsecp256k1-dev
```

**Alpine:**

```bash
apk add libsecp256k1-dev
```

</details>

### When to Use Each Backend

**Decred (Pure Go):**

- ✅ Cross-compilation
- ✅ WebAssembly targets
- ✅ No system dependencies
- ✅ Excellent baseline performance

**Ethereum (libsecp256k1):**

- ✅ Production systems
- ✅ Ring signature workloads
- ✅ High-throughput operations
- ✅ 3x performance critical paths

## API Usage

```go
import (
    "github.com/athanorlabs/go-dleq"
    "github.com/athanorlabs/go-dleq/ed25519"
    "github.com/athanorlabs/go-dleq/secp256k1"
)

// Create curves
curveA := secp256k1.NewCurve()
curveB := ed25519.NewCurve()

// Generate secret
secret, _ := dleq.GenerateSecretForCurves(curveA, curveB)

// Create and verify proof
proof, _ := dleq.NewProof(curveA, curveB, secret)
err := proof.Verify(curveA, curveB)
```

API is identical between backends - just change build tags.

## Technical Details

<details>
<summary><b>🔧 Implementation Details</b></summary>

### Architecture

- **Backend selection:** Build-time via tags (not runtime)
- **Files:**
  - `secp256k1/curve_decred.go` - Pure Go implementation
  - `secp256k1/curve_ethereum.go` - libsecp256k1 wrapper (build tag: `ethereum_secp256k1`)
  - `secp256k1/curve_ethereum_pooling.go` - Memory optimization pools for Ethereum backend

### Optimizations

The Ethereum backend replaces critical operations:

| Decred (Pure Go)                   | Ethereum (libsecp256k1)              |
| ---------------------------------- | ------------------------------------ |
| `secp256k1.ScalarMultNonConst`     | `ethsecp256k1.S256().ScalarMult`     |
| `secp256k1.ScalarBaseMultNonConst` | `ethsecp256k1.S256().ScalarBaseMult` |
| `ecdsa.SignASN1`                   | `ethsecp256k1.Sign`                  |
| `ecdsa.VerifyASN1`                 | `ethsecp256k1.VerifySignature`       |

### Compatibility Testing

Backend consistency is verified through deterministic tests:

- **Mathematical equivalence:** Same inputs produce identical outputs
- **Signature correctness:** All backends generate valid signatures
- **DLEQ proof integrity:** Proofs verify correctly across implementations
- **Deterministic behavior:** Known test vectors ensure repeatability

```bash
make test_compatibility           # Run compatibility tests
go test -run TestBackendCompatibility  # Direct test execution
```

### Performance TODOs

- ✅ **COMPLETED**: Reduced Ethereum backend allocations (336 B/op → 328 B/op)
- `TODO_OPTIMIZE`: ScalarBaseMul slower than Decred (43μs vs 36μs)
- `TODO_IMPROVE`: Replace panics with error returns

### Shannon SDK Impact

The PATH → Shannon SDK → Ring-go → go-dleq pipeline achieves:

- **Ring signatures:** ~3x faster
- **DLEQ proofs:** ~3x faster
- **Parallel ops:** ~3x better scaling

</details>
