#!/usr/bin/env python3
"""
Benchmark formatter for go-dleq performance comparison.

This script processes Go benchmark output and creates formatted tables
comparing Decred (pure Go) vs Ethereum (libsecp256k1) backends.

Usage:
    ./format_benchmark.py < benchmark_output.txt
    make benchmark_all | python3 format_benchmark.py
    go test -bench=. -benchmem | python3 format_benchmark.py
"""

import sys
import re
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass
from enum import Enum


class Backend(Enum):
    DECRED = "Decred"
    ETHEREUM = "Ethereum"
    UNKNOWN = "Unknown"


@dataclass
class BenchmarkResult:
    name: str
    backend: Backend
    iterations: int
    ns_per_op: float
    bytes_per_op: float
    allocs_per_op: float


def parse_benchmark_line(line: str) -> Optional[BenchmarkResult]:
    """Parse a single Go benchmark output line."""
    # Example: BenchmarkComparison_ScalarMul-10    	    9624	    120260 ns/op	     136 B/op	       2 allocs/op
    # Example: BenchmarkScalarBaseMul-10           	  101220	     35105 ns/op	     136 B/op	       2 allocs/op

    # Handle both spaces and tabs in Go benchmark output
    benchmark_pattern = r'Benchmark(\w+)-\d+\s+(\d+)\s+(\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op'

    # Normalize whitespace (replace tabs and multiple spaces with single space)
    normalized_line = re.sub(r'\s+', ' ', line.strip())
    match = re.match(benchmark_pattern, normalized_line)

    if not match:
        return None

    bench_name = match.group(1)
    iterations = int(match.group(2))
    ns_per_op = float(match.group(3))
    bytes_per_op = float(match.group(4))
    allocs_per_op = float(match.group(5))

    # Clean up benchmark name
    if bench_name.startswith('Comparison_'):
        bench_name = bench_name[11:]  # Remove 'Comparison_' prefix

    return BenchmarkResult(
        name=bench_name,
        backend=Backend.UNKNOWN,  # Will be determined by context
        iterations=iterations,
        ns_per_op=ns_per_op,
        bytes_per_op=bytes_per_op,
        allocs_per_op=allocs_per_op
    )


def detect_backend_from_context(lines: List[str], line_idx: int) -> Backend:
    """Detect which backend based on surrounding context."""
    # Look backwards for backend indicators
    for i in range(max(0, line_idx - 10), line_idx):
        line = lines[i].lower()
        if 'decred' in line or 'pure go' in line:
            return Backend.DECRED
        elif 'ethereum' in line or 'libsecp256k1' in line:
            return Backend.ETHEREUM
        elif 'cgo_enabled=0' in line:
            return Backend.DECRED
        elif 'cgo_enabled=1' in line or 'tags=ethereum_secp256k1' in line:
            return Backend.ETHEREUM

    return Backend.UNKNOWN


def format_time(ns: float) -> str:
    """Format nanoseconds into human-readable time units."""
    if ns >= 1_000_000_000:
        return f"{ns/1_000_000_000:.1f}s"
    elif ns >= 1_000_000:
        return f"{ns/1_000_000:.1f}ms"
    elif ns >= 1_000:
        return f"{ns/1_000:.0f}μs"
    else:
        return f"{ns:.0f}ns"


def format_memory(bytes_val: float) -> str:
    """Format bytes into human-readable memory units."""
    if bytes_val >= 1_048_576:
        return f"{bytes_val/1_048_576:.1f}MB"
    elif bytes_val >= 1_024:
        return f"{bytes_val/1_024:.1f}KB"
    else:
        return f"{bytes_val:.0f}B"


def format_number(num: float) -> str:
    """Format large numbers with K/M suffixes."""
    if num >= 1_000_000:
        return f"{num/1_000_000:.1f}M"
    elif num >= 1_000:
        return f"{num/1_000:.1f}K"
    else:
        return f"{num:.0f}"


def calculate_improvement(decred_val: float, ethereum_val: float) -> str:
    """Calculate improvement percentage/multiplier."""
    if decred_val == 0 or ethereum_val == 0:
        return "N/A"

    if ethereum_val < decred_val:
        # Ethereum is faster (lower is better for time)
        improvement = decred_val / ethereum_val
        return f"{improvement:.1f}x faster"
    else:
        # Decred is faster
        degradation = ethereum_val / decred_val
        return f"{degradation:.1f}x slower"


def group_benchmarks_by_operation(results: List[BenchmarkResult]) -> Dict[str, Dict[Backend, BenchmarkResult]]:
    """Group benchmark results by operation name and backend."""
    grouped = {}

    for result in results:
        if result.name not in grouped:
            grouped[result.name] = {}
        grouped[result.name][result.backend] = result

    return grouped


def create_comparison_table(grouped_results: Dict[str, Dict[Backend, BenchmarkResult]]) -> str:
    """Create a formatted comparison table."""
    if not grouped_results:
        return "❌ No benchmark results found to compare."

    # Table header
    table = []
    table.append("🔬 **Performance Comparison: Decred vs Ethereum Backends**")
    table.append("")
    table.append("| Operation | Decred (Pure Go) | Ethereum (libsecp256k1) | Improvement |")
    table.append("|-----------|------------------|--------------------------|-------------|")

    # Sort operations by name for consistent output
    for operation in sorted(grouped_results.keys()):
        backends = grouped_results[operation]

        decred_result = backends.get(Backend.DECRED)
        ethereum_result = backends.get(Backend.ETHEREUM)

        if not decred_result and not ethereum_result:
            continue

        # Format the results
        decred_time = format_time(decred_result.ns_per_op) if decred_result else "N/A"
        ethereum_time = format_time(ethereum_result.ns_per_op) if ethereum_result else "N/A"

        # Calculate improvement
        improvement = "N/A"
        if decred_result and ethereum_result:
            improvement = calculate_improvement(decred_result.ns_per_op, ethereum_result.ns_per_op)

        table.append(f"| **{operation}** | {decred_time} | {ethereum_time} | **{improvement}** |")

    table.append("")

    # Add memory comparison if available
    if any(Backend.DECRED in backends and Backend.ETHEREUM in backends
           for backends in grouped_results.values()):
        table.append("### Memory Usage Comparison")
        table.append("")
        table.append("| Operation | Decred Memory | Ethereum Memory | Decred Allocs | Ethereum Allocs |")
        table.append("|-----------|---------------|-----------------|---------------|-----------------|")

        for operation in sorted(grouped_results.keys()):
            backends = grouped_results[operation]
            decred_result = backends.get(Backend.DECRED)
            ethereum_result = backends.get(Backend.ETHEREUM)

            if decred_result and ethereum_result:
                decred_mem = format_memory(decred_result.bytes_per_op)
                ethereum_mem = format_memory(ethereum_result.bytes_per_op)
                decred_allocs = format_number(decred_result.allocs_per_op)
                ethereum_allocs = format_number(ethereum_result.allocs_per_op)

                table.append(f"| {operation} | {decred_mem} | {ethereum_mem} | {decred_allocs} | {ethereum_allocs} |")

    return "\n".join(table)


def create_summary_stats(grouped_results: Dict[str, Dict[Backend, BenchmarkResult]]) -> str:
    """Create summary statistics."""
    improvements = []

    for operation, backends in grouped_results.items():
        decred = backends.get(Backend.DECRED)
        ethereum = backends.get(Backend.ETHEREUM)

        if decred and ethereum:
            ratio = decred.ns_per_op / ethereum.ns_per_op
            improvements.append((operation, ratio))

    if not improvements:
        return ""

    # Sort by improvement ratio (highest first)
    improvements.sort(key=lambda x: x[1], reverse=True)

    summary = []
    summary.append("### 🚀 **Performance Summary**")
    summary.append("")
    summary.append("**Top Improvements (Ethereum vs Decred):**")
    summary.append("")

    for operation, ratio in improvements[:5]:  # Top 5
        if ratio > 1.1:  # Only show significant improvements
            summary.append(f"- **{operation}**: {ratio:.1f}x faster")
        elif ratio < 0.9:  # Show regressions too
            summary.append(f"- **{operation}**: {1/ratio:.1f}x slower")

    summary.append("")

    # Calculate average improvement
    if improvements:
        avg_improvement = sum(ratio for _, ratio in improvements) / len(improvements)
        summary.append(f"**Average Performance Improvement**: {avg_improvement:.1f}x")

    return "\n".join(summary)


def main():
    """Main function to process benchmark output."""
    if len(sys.argv) > 1 and sys.argv[1] in ['-h', '--help']:
        print(__doc__)
        return

    # Read all input
    lines = []
    try:
        for line in sys.stdin:
            lines.append(line.rstrip())
    except KeyboardInterrupt:
        return

    if not lines:
        print("❌ No input provided. Please pipe benchmark output to this script.")
        print("Example: make benchmark_all | python3 format_benchmark.py")
        return

    # Parse benchmark results
    results = []
    current_backend = Backend.UNKNOWN

    for i, line in enumerate(lines):
        # Detect backend context
        line_lower = line.lower()
        if 'decred' in line_lower or 'pure go' in line_lower:
            current_backend = Backend.DECRED
        elif 'ethereum' in line_lower or 'libsecp256k1' in line_lower:
            current_backend = Backend.ETHEREUM

        # Parse benchmark line
        result = parse_benchmark_line(line)
        if result:
            # Use current context or try to detect from surrounding lines
            if current_backend != Backend.UNKNOWN:
                result.backend = current_backend
            else:
                result.backend = detect_backend_from_context(lines, i)

            results.append(result)

    if not results:
        print("❌ No benchmark results found in input.")
        print("Make sure you're piping Go benchmark output (go test -bench=. -benchmem)")
        return

    # Group results and create table
    grouped = group_benchmarks_by_operation(results)

    print(create_comparison_table(grouped))
    print()
    print(create_summary_stats(grouped))

    # Add usage note
    print()
    print("---")
    print("💡 **Usage Tips:**")
    print("- Run `make benchmark_all | python3 format_benchmark.py` for full comparison")
    print("- Use `make benchmark_report` for quick results")
    print("- Both backends must be available for meaningful comparison")


if __name__ == "__main__":
    main()