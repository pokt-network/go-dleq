// Consolidated benchmark tool for go-dleq backend comparison
//
// This single Go program replaces three separate files:
//   - benchmark_runner.sh (Bash script)
//   - format_benchmark.py (Python formatter)
//   - format_benchmark_terminal.py (Python terminal formatter)
//
// Benefits of consolidation:
//   - No Python dependency required
//   - Single portable Go binary
//   - Easier to maintain and test
//   - Better cross-platform support
//   - Consistent formatting across environments
//
// Usage:
//   go run cmd/benchmark/main.go -compare         # Full comparison
//   go run cmd/benchmark/main.go -report          # Quick report
//   go run cmd/benchmark/main.go -compare -duration=10s  # Custom duration

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorBlue   = "\033[0;34m"
	colorYellow = "\033[1;33m"
	colorCyan   = "\033[0;36m"
)

type BenchmarkResult struct {
	Name      string
	NsOp      float64
	BytesOp   int
	AllocsOp  int
	Backend   string
}

func main() {
	var (
		report   = flag.Bool("report", false, "Generate a formatted report")
		compare  = flag.Bool("compare", false, "Run comparison between backends")
		duration = flag.String("duration", "3s", "Benchmark duration")
	)
	flag.Parse()

	if *report || *compare {
		runComparison(*duration)
	} else {
		fmt.Println("Usage: go run cmd/benchmark/main.go [options]")
		fmt.Println("  -compare   Run full comparison between backends")
		fmt.Println("  -report    Generate a quick performance report")
		fmt.Println("  -duration  Benchmark duration (default: 3s)")
	}
}

func runComparison(duration string) {
	fmt.Printf("%s🔬 Go-DLEQ Backend Performance Comparison%s\n", colorBlue, colorReset)
	fmt.Println("==========================================")
	fmt.Println()

	// Check CGO availability
	if !checkCGO() {
		fmt.Printf("%s⚠️  CGO not available. Only Decred backend will be tested.%s\n\n", colorYellow, colorReset)
	}

	// Run Decred backend benchmarks
	fmt.Printf("%s📊 Testing Decred Backend (Pure Go)%s\n", colorBlue, colorReset)
	decredResults := runBenchmarks("", "0", duration)

	// Run Ethereum backend benchmarks if available
	var ethResults []BenchmarkResult
	if checkCGO() {
		fmt.Printf("\n%s📊 Testing Ethereum Backend (libsecp256k1)%s\n", colorBlue, colorReset)
		ethResults = runBenchmarks("-tags=ethereum_secp256k1", "1", duration)
	}

	// Display comparison
	if len(ethResults) > 0 {
		displayComparison(decredResults, ethResults)
	}
}

func checkCGO() bool {
	cmd := exec.Command("go", "env", "CGO_ENABLED")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	// Check if gcc is available
	if exec.Command("gcc", "--version").Run() != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != "0"
}

func runBenchmarks(tags, cgoEnabled, duration string) []BenchmarkResult {
	args := []string{"test"}
	if tags != "" {
		args = append(args, tags)
	}
	args = append(args,
		"-bench=BenchmarkComparison",
		"-benchmem",
		"-run=^$",
		"-benchtime="+duration,
	)

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "CGO_ENABLED="+cgoEnabled)

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s❌ Error running benchmarks: %v%s\n", colorRed, err, colorReset)
		return nil
	}

	return parseBenchmarkOutput(string(output))
}

func parseBenchmarkOutput(output string) []BenchmarkResult {
	var results []BenchmarkResult

	// Parse benchmark output lines
	// Format: BenchmarkComparison_Name-10    1000    1234567 ns/op    336 B/op    8 allocs/op
	re := regexp.MustCompile(`BenchmarkComparison_(\w+)(?:-\d+)?\s+(\d+)\s+([\d.]+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op`)

	matches := re.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 6 {
			nsOp, _ := strconv.ParseFloat(match[3], 64)
			bytesOp, _ := strconv.Atoi(match[4])
			allocsOp, _ := strconv.Atoi(match[5])

			results = append(results, BenchmarkResult{
				Name:     match[1],
				NsOp:     nsOp,
				BytesOp:  bytesOp,
				AllocsOp: allocsOp,
			})
		}
	}

	return results
}

func displayComparison(decred, ethereum []BenchmarkResult) {
	fmt.Printf("\n%s=== Performance Comparison ===%s\n\n", colorGreen, colorReset)

	// Create a map for easy lookup
	ethMap := make(map[string]BenchmarkResult)
	for _, r := range ethereum {
		ethMap[r.Name] = r
	}

	// Table header
	fmt.Printf("%-30s %15s %15s %12s\n", "Operation", "Decred (Pure Go)", "Ethereum (CGO)", "Improvement")
	fmt.Println(strings.Repeat("-", 75))

	for _, d := range decred {
		if e, ok := ethMap[d.Name]; ok {
			// Format times
			dTime := formatTime(d.NsOp)
			eTime := formatTime(e.NsOp)

			// Calculate improvement
			improvement := d.NsOp / e.NsOp
			impStr := fmt.Sprintf("%.1fx", improvement)

			// Color code based on improvement
			color := colorReset
			if improvement > 2 {
				color = colorGreen
				impStr = "🚀 " + impStr + " faster"
			} else if improvement > 1.5 {
				color = colorYellow
				impStr = "⚡ " + impStr + " faster"
			} else if improvement < 0.9 {
				color = colorRed
				impStr = "🐌 " + fmt.Sprintf("%.1fx slower", 1/improvement)
			}

			fmt.Printf("%-30s %15s %15s %s%12s%s\n",
				d.Name, dTime, eTime, color, impStr, colorReset)
		}
	}

	fmt.Println("\n" + strings.Repeat("-", 75))

	// Memory comparison
	fmt.Printf("\n%s=== Memory Usage Comparison ===%s\n\n", colorCyan, colorReset)
	fmt.Printf("%-30s %20s %20s\n", "Operation", "Decred", "Ethereum")
	fmt.Println(strings.Repeat("-", 75))

	for _, d := range decred {
		if e, ok := ethMap[d.Name]; ok {
			dMem := fmt.Sprintf("%d B, %d allocs", d.BytesOp, d.AllocsOp)
			eMem := fmt.Sprintf("%d B, %d allocs", e.BytesOp, e.AllocsOp)

			fmt.Printf("%-30s %20s %20s\n", d.Name, dMem, eMem)
		}
	}
}

func formatTime(ns float64) string {
	switch {
	case ns >= 1_000_000_000:
		return fmt.Sprintf("%.1f s", ns/1_000_000_000)
	case ns >= 1_000_000:
		return fmt.Sprintf("%.1f ms", ns/1_000_000)
	case ns >= 1_000:
		return fmt.Sprintf("%.0f μs", ns/1_000)
	default:
		return fmt.Sprintf("%.0f ns", ns)
	}
}