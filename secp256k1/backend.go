package secp256k1

// This file provides a common interface for backend operations.
// The actual implementations are in curve_decred.go and curve_ethereum.go,
// selected at build time via build tags.

// Backend performance characteristics:
//
// Decred backend (default, pure Go):
// - No CGO dependencies
// - Excellent performance (35μs base mul, 120μs scalar mul)
// - Maximum portability
// - 2 memory allocations per operation
//
// Ethereum backend (build tag: ethereum_secp256k1):
// - Requires CGO and libsecp256k1
// - ~50% faster signing operations
// - ~80% faster verification operations
// - Fewer memory allocations
// - Production-ready (Bitcoin Core implementation)

// Common backend operations that both implementations provide:
//
// Key Operations (optimized in Ethereum backend):
// - ScalarBaseMul: Generator point multiplication
// - ScalarMul: Arbitrary point multiplication
// - Sign: ECDSA signature generation
// - Verify: ECDSA signature verification
//
// The Ethereum backend replaces expensive Decred operations:
// - secp256k1.ScalarBaseMultNonConst -> ethsecp256k1.S256().ScalarBaseMult
// - secp256k1.ScalarMultNonConst -> ethsecp256k1.S256().ScalarMult
// - ecdsa.SignASN1 -> ethsecp256k1.Sign (+ DER conversion)
// - ecdsa.VerifyASN1 -> ethsecp256k1.VerifySignature (+ DER conversion)

// Build instructions:
//
// Default (Decred backend):
//   go build
//   CGO_ENABLED=0 go build
//
// Ethereum backend:
//   go build -tags="ethereum_secp256k1"
//   CGO_ENABLED=1 go build -tags="ethereum_secp256k1"
