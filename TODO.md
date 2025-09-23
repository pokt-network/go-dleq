# TODO.md - go-dleq Optimization Roadmap <!-- omit in toc -->

- [Completed ✅](#completed-)
- [Priority Optimizations 🚀](#priority-optimizations-)
- [Future Enhancements 🔮](#future-enhancements-)
- [Technical Debt 🧹](#technical-debt-)
- [Integration Tasks 🔗](#integration-tasks-)

## Completed ✅

### Core Backend Implementation
- [x] **Dual backend architecture** - Decred (pure Go) + Ethereum (libsecp256k1)
- [x] **Build tag system** - Clean separation via `ethereum_secp256k1` tag
- [x] **Go 1.24.3 compatibility** - Updated module and dependencies
- [x] **Performance benchmarking** - Comprehensive comparison infrastructure
- [x] **API compatibility** - 100% drop-in replacement guarantee
- [x] **Documentation** - Updated README with performance data

### Performance Achievements
- [x] **3x faster ScalarMul** - Critical for ring signatures (125μs → 43μs)
- [x] **2.6x faster signing** - ECDSA operations (93μs → 36μs)
- [x] **5x faster verification** - ECDSA verification (212μs → 42μs)
- [x] **3x faster DLEQ proofs** - End-to-end proof generation/verification
- [x] **3x better parallel scaling** - Multi-core workloads

## Priority Optimizations 🚀

### Memory Optimization
- [ ] **Reduce Ethereum backend allocations** - Currently 8 allocs vs 2 for Decred
  - Target: Bring down to 4-6 allocations per operation
  - Approach: Pool big.Int objects, reduce intermediate allocations
  - Impact: Lower GC pressure in high-throughput scenarios

### Algorithm Optimization
- [ ] **Optimize ScalarBaseMul in Ethereum backend** - Currently slower than Decred
  - Issue: 43μs vs 36μs for base point multiplication
  - Approach: Direct libsecp256k1 calls instead of Go curve interface
  - Target: Match or beat Decred performance (≤36μs)

### Build System
- [ ] **Add Makefile targets** for easy backend switching
  ```bash
  make build-fast      # Ethereum backend
  make build-portable  # Decred backend
  make benchmark-all   # Compare both backends
  ```

### Error Handling
- [ ] **Improve error messages** with backend-specific context
- [ ] **Add validation** for build-time vs runtime backend mismatches
- [ ] **CGO availability detection** with fallback guidance

## Future Enhancements 🔮

### Additional Backends
- [ ] **BLST backend** - Even faster BLS12-381 operations for future curves
  - Research: Evaluate compatibility with existing DLEQ interface
  - Performance target: 2x faster than Ethereum backend

- [ ] **Hardware acceleration** - Intel ADX/BMI2, ARM crypto extensions
  - Platform-specific optimizations
  - Auto-detection and fallback

### Advanced Features
- [ ] **Batch operations** - Multi-scalar operations in single call
  - ScalarMul batch: Process N operations together
  - Memory efficiency: Reduce per-operation overhead

- [ ] **Precomputed tables** - For fixed base points
  - 8x faster ScalarBaseMul with precomputation
  - Optional feature with memory trade-off

### Performance Monitoring
- [ ] **Runtime performance metrics** - Optional instrumentation
- [ ] **Benchmark CI integration** - Performance regression detection
- [ ] **Memory profiling tools** - Automated allocation analysis

## Technical Debt 🧹

### Code Quality
- [ ] **Reduce code duplication** between backends
  - Extract common validation logic
  - Shared test utilities
  - Common error handling patterns

- [ ] **Improve test coverage** for Ethereum backend edge cases
  - Nil pointer handling (partially done)
  - Invalid input validation
  - CGO failure scenarios

### Documentation
- [ ] **Add architecture diagrams** - Visual backend selection guide
- [ ] **Performance tuning guide** - Real-world optimization tips
- [ ] **Migration guide** - From original go-dleq to this fork

### Maintenance
- [ ] **Automated dependency updates** - Keep go-ethereum current
- [ ] **Cross-platform testing** - Ensure CGO works on all targets
- [ ] **Version compatibility matrix** - Go version support policy

## Integration Tasks 🔗

### Shannon SDK Integration
- [ ] **Update ring-go dependency** to use this fork
  ```go
  replace github.com/athanorlabs/go-dleq => github.com/yourusername/go-dleq v0.2.0
  ```

- [ ] **Shannon SDK build configuration** - Default to Ethereum backend
  - Dockerfile updates with libsecp256k1
  - CI/CD pipeline modifications
  - Performance testing integration

### Upstream Contribution
- [ ] **Prepare upstreaming proposal** to original go-dleq
  - Clean commit history
  - Comprehensive test suite
  - Performance documentation
  - Backward compatibility guarantee

### Distribution
- [ ] **Release automation** - Tagged releases with prebuilt binaries
- [ ] **Docker images** - Both backend variants
- [ ] **Performance dashboard** - Public benchmark results

---

## Implementation Notes

### Performance Targets
- **ScalarBaseMul**: Target ≤30μs (currently 43μs Ethereum, 36μs Decred)
- **Memory usage**: Target ≤200 B/op (currently 336 B/op Ethereum)
- **DLEQ operations**: Target ≤100ms (currently 157ms generation, 131ms verification)

### Success Metrics
- **Ring signature operations**: Maintain 3x speedup over Decred
- **API compatibility**: Zero breaking changes
- **Build reliability**: 100% success rate across platforms
- **Documentation quality**: Zero external questions about backend selection

### Risk Mitigation
- **CGO dependency**: Comprehensive fallback documentation
- **Performance regression**: Automated benchmark CI
- **API changes**: Strict semantic versioning
- **Security**: Regular dependency audits