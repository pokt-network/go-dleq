//go:build cgo && ethereum_secp256k1
// +build cgo,ethereum_secp256k1

package secp256k1

import (
	"math/big"
	"sync"
)

// Object pools for memory optimization in Ethereum backend
// These pools reduce allocations in performance-critical paths by reusing objects

var (
	// Pool for big.Int objects to reduce allocations in scalar operations
	bigIntPool = sync.Pool{
		New: func() interface{} {
			return new(big.Int)
		},
	}

	// Pool for 32-byte slices (common for scalar operations)
	bytes32Pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32)
		},
	}

	// Pool for 33-byte slices (compressed points)
	bytes33Pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 33)
		},
	}

	// Pool for 64-byte slices (common for signature operations)
	bytes64Pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 64)
		},
	}

	// Pool for 65-byte slices (uncompressed public keys)
	bytes65Pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 65)
		},
	}
)

// getBigInt retrieves a big.Int from the pool
func getBigInt() *big.Int {
	return bigIntPool.Get().(*big.Int)
}

// putBigInt returns a big.Int to the pool after clearing it for security
func putBigInt(b *big.Int) {
	b.SetInt64(0) // Clear the value for security
	bigIntPool.Put(b)
}

// getBytes32 retrieves a 32-byte slice from the pool
func getBytes32() []byte {
	return bytes32Pool.Get().([]byte)
}

// putBytes32 returns a 32-byte slice to the pool after clearing it for security
func putBytes32(b []byte) {
	// Clear the slice for security
	for i := range b {
		b[i] = 0
	}
	bytes32Pool.Put(b)
}

// getBytes33 retrieves a 33-byte slice from the pool
func getBytes33() []byte {
	return bytes33Pool.Get().([]byte)
}

// putBytes33 returns a 33-byte slice to the pool after clearing it for security
func putBytes33(b []byte) {
	// Clear the slice for security
	for i := range b {
		b[i] = 0
	}
	bytes33Pool.Put(b)
}

// getBytes64 retrieves a 64-byte slice from the pool
func getBytes64() []byte {
	return bytes64Pool.Get().([]byte)
}

// putBytes64 returns a 64-byte slice to the pool after clearing it for security
func putBytes64(b []byte) {
	// Clear the slice for security
	for i := range b {
		b[i] = 0
	}
	bytes64Pool.Put(b)
}

// getBytes65 retrieves a 65-byte slice from the pool
func getBytes65() []byte {
	return bytes65Pool.Get().([]byte)
}

// putBytes65 returns a 65-byte slice to the pool after clearing it for security
func putBytes65(b []byte) {
	// Clear the slice for security
	for i := range b {
		b[i] = 0
	}
	bytes65Pool.Put(b)
}