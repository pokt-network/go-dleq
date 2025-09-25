package secp256k1

// Dual backend secp256k1 implementation:
// - curve_decred.go: Pure Go (default)
// - curve_ethereum.go: libsecp256k1 wrapper (build tag: ethereum_secp256k1)
// - curve_ethereum_pooling.go: Memory optimization pools for Ethereum backend
//
// Build commands:
//   CGO_ENABLED=0 go build                                    # Decred backend
//   CGO_ENABLED=1 go build -tags="ethereum_secp256k1"        # Ethereum backend
