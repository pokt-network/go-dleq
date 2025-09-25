// Usage:
//   go run cmd/benchmark/main.go -compare         # Full comparison
//   go run cmd/benchmark/main.go -report          # Quick report
//   go run cmd/benchmark/main.go -compare -duration=10s  # Custom duration

package main

import (
	"bufio"
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
	Name     string
	NsOp     float64
	BytesOp  int
	AllocsOp int
	Backend  string
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

	// Run compatibility verification first
	if checkCGO() {
		fmt.Printf("%s🔍 Verifying Backend Compatibility%s\n", colorBlue, colorReset)
		if !runCompatibilityTest() {
			fmt.Printf("%s❌ Backend compatibility test failed!%s\n", colorRed, colorReset)
			os.Exit(1)
		}
		fmt.Printf("%s✅ Backend compatibility verified%s\n\n", colorGreen, colorReset)
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

func runCompatibilityTest() bool {
	fmt.Println("  • Testing Decred backend produces consistent results...")

	// Run Decred backend compatibility test
	decredCmd := exec.Command("go", "test", "-v", "-run", "TestBackendCompatibility")
	decredCmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	decredOut, err := decredCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s    ❌ Decred backend test failed: %v%s\n", colorRed, err, colorReset)
		fmt.Printf("%s%s%s\n", colorRed, string(decredOut), colorReset)
		return false
	}

	fmt.Println("  • Testing Ethereum backend produces consistent results...")

	// Run Ethereum backend compatibility test
	ethCmd := exec.Command("go", "test", "-tags=ethereum_secp256k1", "-v", "-run", "TestBackendCompatibility")
	ethCmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	ethOut, err := ethCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s    ❌ Ethereum backend test failed: %v%s\n", colorRed, err, colorReset)
		fmt.Printf("%s%s%s\n", colorRed, string(ethOut), colorReset)
		return false
	}

	// Parse outputs to verify both backends produce same deterministic results
	fmt.Println("  • Comparing deterministic outputs...")

	decredValues := extractDeterministicValues(string(decredOut))
	ethValues := extractDeterministicValues(string(ethOut))

	if len(decredValues) == 0 || len(ethValues) == 0 {
		fmt.Printf("%s    ❌ Could not extract deterministic values from test outputs%s\n", colorRed, colorReset)
		return false
	}

	// Compare extracted values
	for key, decredValue := range decredValues {
		if ethValue, exists := ethValues[key]; !exists || ethValue != decredValue {
			fmt.Printf("%s    ❌ Backends produce different outputs for %s%s\n", colorRed, key, colorReset)
			fmt.Printf("      Decred:   %s\n", decredValue)
			fmt.Printf("      Ethereum: %s\n", ethValue)
			return false
		}
	}

	fmt.Printf("    ✅ Both backends produce identical deterministic outputs\n")
	return true
}

func extractDeterministicValues(output string) map[string]string {
	values := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "DETERMINISTIC_") {
			parts := strings.SplitN(line, "DETERMINISTIC_", 2)
			if len(parts) == 2 {
				keyValue := strings.SplitN(parts[1], "=", 2)
				if len(keyValue) == 2 {
					values[keyValue[0]] = keyValue[1]
				}
			}
		}
	}

	return values
}
