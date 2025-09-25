package secp256k1

// TODO_OPTIMIZE: Reduce Ethereum backend allocations from 8 to 4-6 by pooling big.Int objects

// This file provides a common interface for backend operations.
// The actual implementations are in curve_decred.go and curve_ethereum.go.
// The exact implementation is selected at build time via build tags.

// Backend performance characteristics:
//
// 1) Decred backend (default, pure Go):
// - No CGO dependencies
// - Excellent performance (35μs base mul, 120μs scalar mul)
// - Maximum portability
// - 2 memory allocations per operation
//
// 2) Ethereum backend (build tag: ethereum_secp256k1):
// - Requires CGO and libsecp256k1
// - ~3x faster ScalarMul operations (127μs → 42μs)
// - ~2.6x faster signing operations (93μs → 36μs)
// - ~5x faster verification operations (212μs → 42μs)
// - Higher memory usage (8 allocs vs 2, needs optimization)
// - Production-ready (Bitcoin Core implementation)

// Build instructions:
//
// Default (Decred backend):
//   go build
//   CGO_ENABLED=0 go build
//
// Ethereum backend:
//   go build -tags="ethereum_secp256k1"
//   CGO_ENABLED=1 go build -tags="ethereum_secp256k1"
