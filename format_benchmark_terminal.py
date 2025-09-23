#!/usr/bin/env python3
"""
Terminal-optimized benchmark formatter for go-dleq performance comparison.

This script creates clean, aligned tables optimized for terminal display
with proper spacing, colors, and readability.
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


def calculate_improvement(decred_val: float, ethereum_val: float) -> Tuple[str, float]:
    """Calculate improvement percentage/multiplier and return with ratio."""
    if decred_val == 0 or ethereum_val == 0:
        return "N/A", 1.0

    ratio = decred_val / ethereum_val
    if ethereum_val < decred_val:
        # Ethereum is faster (lower is better for time)
        return f"{ratio:.1f}x faster", ratio
    else:
        # Decred is faster
        return f"{1/ratio:.1f}x slower", ratio


def group_benchmarks_by_operation(results: List[BenchmarkResult]) -> Dict[str, Dict[Backend, BenchmarkResult]]:
    """Group benchmark results by operation name and backend."""
    grouped = {}

    for result in results:
        if result.name not in grouped:
            grouped[result.name] = {}
        grouped[result.name][result.backend] = result

    return grouped


def create_terminal_table(grouped_results: Dict[str, Dict[Backend, BenchmarkResult]]) -> str:
    """Create a terminal-optimized comparison table."""
    if not grouped_results:
        return "❌ No benchmark results found to compare."

    # Collect data and sort by improvement
    data_rows = []
    for operation in grouped_results.keys():
        backends = grouped_results[operation]
        decred_result = backends.get(Backend.DECRED)
        ethereum_result = backends.get(Backend.ETHEREUM)

        if decred_result and ethereum_result:
            improvement_text, ratio = calculate_improvement(decred_result.ns_per_op, ethereum_result.ns_per_op)
            data_rows.append({
                'operation': operation,
                'decred_time': format_time(decred_result.ns_per_op),
                'ethereum_time': format_time(ethereum_result.ns_per_op),
                'improvement': improvement_text,
                'ratio': ratio,
                'decred_mem': format_memory(decred_result.bytes_per_op),
                'ethereum_mem': format_memory(ethereum_result.bytes_per_op),
                'decred_allocs': int(decred_result.allocs_per_op),
                'ethereum_allocs': int(ethereum_result.allocs_per_op)
            })

    # Sort by improvement ratio (best improvements first)
    data_rows.sort(key=lambda x: x['ratio'], reverse=True)

    # Calculate column widths
    max_op_width = max(len(row['operation']) for row in data_rows) + 2
    max_op_width = max(max_op_width, 22)  # Minimum width

    # Create output
    lines = []
    lines.append("")
    lines.append("\033[1;36m🔬 Performance Comparison: Decred vs Ethereum Backends\033[0m")
    lines.append("\033[90m" + "=" * 65 + "\033[0m")
    lines.append("")

    # Performance table header
    header = f"{'Operation':<{max_op_width}} │ {'Decred':<9} │ {'Ethereum':<9} │ {'Improvement':<14}"
    lines.append(header)
    lines.append("─" * len(header))

    # Performance table rows
    for row in data_rows:
        # Add improvement indicators
        if row['ratio'] >= 3.0:
            indicator = "🚀"
        elif row['ratio'] >= 2.0:
            indicator = "⚡"
        elif row['ratio'] >= 1.3:
            indicator = "✨"
        elif row['ratio'] < 0.9:
            indicator = "⚠️ "
        else:
            indicator = "  "

        improvement_with_indicator = f"{indicator} {row['improvement']}"

        line = f"{row['operation']:<{max_op_width}} │ {row['decred_time']:>8} │ {row['ethereum_time']:>8} │ {improvement_with_indicator:<14}"
        lines.append(line)

    lines.append("")

    # Memory usage section
    lines.append("Memory Usage Comparison")
    lines.append("─" * 35)
    lines.append("")

    mem_header = f"{'Operation':<{max_op_width}} │ {'Decred':<10} │ {'Ethereum':<10} │ {'Allocs':<8}"
    lines.append(mem_header)
    lines.append("─" * len(mem_header))

    for row in data_rows:
        alloc_info = f"{row['decred_allocs']}/{row['ethereum_allocs']}"
        line = f"{row['operation']:<{max_op_width}} │ {row['decred_mem']:>9} │ {row['ethereum_mem']:>9} │ {alloc_info:<8}"
        lines.append(line)

    lines.append("")

    # Summary section
    lines.append("🚀 Performance Summary")
    lines.append("─" * 25)
    lines.append("")

    # Top improvements
    top_improvements = [row for row in data_rows if row['ratio'] >= 1.2][:5]
    if top_improvements:
        lines.append("Top Improvements (Ethereum vs Decred):")
        for row in top_improvements:
            if row['ratio'] >= 2.0:
                lines.append(f"  🚀 {row['operation']}: {row['improvement']}")
            else:
                lines.append(f"  ⚡ {row['operation']}: {row['improvement']}")
        lines.append("")

    # Average improvement
    if data_rows:
        avg_ratio = sum(row['ratio'] for row in data_rows) / len(data_rows)
        lines.append(f"Average Performance Improvement: {avg_ratio:.1f}x")

    lines.append("")

    return "\n".join(lines)


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
        print("Example: make benchmark_report")
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
    print(create_terminal_table(grouped))


if __name__ == "__main__":
    main()