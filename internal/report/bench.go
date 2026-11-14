package report

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mioluigi/go-applied-statistics/stats"
)

// A Benchmark is the median of the repeated measurements of one Go benchmark.
type Benchmark struct {
	Name        string  // without the "Benchmark" prefix and the "-<CPUs>" suffix
	Count       int     // how many times it was measured
	NanosPerOp  float64 // median ns/op
	BytesPerOp  float64 // median B/op
	AllocsPerOp float64 // median allocs/op
}

// benchmarkLine matches one result line of "go test -bench -benchmem".
var benchmarkLine = regexp.MustCompile(`^Benchmark(\S+?)(?:-\d+)?\s+\d+\s+([0-9.]+) ns/op\s+([0-9.]+) B/op\s+([0-9.]+) allocs/op`)

// ReadBenchmarks reads saved "go test -bench -benchmem" output and returns one entry per
// benchmark, in order of first appearance, with the median over its repetitions.
func ReadBenchmarks(path string) ([]Benchmark, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	defer file.Close()
	samples := map[string][][3]float64{}
	var order []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := benchmarkLine.FindStringSubmatch(strings.TrimSpace(scanner.Text()))
		if match == nil {
			continue
		}
		var sample [3]float64
		for i := range sample {
			sample[i], _ = strconv.ParseFloat(match[2+i], 64)
		}
		if _, seen := samples[match[1]]; !seen {
			order = append(order, match[1])
		}
		samples[match[1]] = append(samples[match[1]], sample)
	}
	if len(order) == 0 {
		return nil, fmt.Errorf("%s holds no benchmark results", path)
	}
	benchmarks := make([]Benchmark, 0, len(order))
	for _, name := range order {
		benchmarks = append(benchmarks, medianBenchmark(name, samples[name]))
	}
	return benchmarks, scanner.Err()
}

// medianBenchmark reduces the repetitions of one benchmark to their medians.
func medianBenchmark(name string, samples [][3]float64) Benchmark {
	column := func(position int) float64 {
		values := make([]float64, len(samples))
		for i, sample := range samples {
			values[i] = sample[position]
		}
		return stats.Median(values)
	}
	return Benchmark{Name: name, Count: len(samples), NanosPerOp: column(0), BytesPerOp: column(1), AllocsPerOp: column(2)}
}

// FindBenchmark returns the benchmark with the given name, and whether it exists.
func FindBenchmark(benchmarks []Benchmark, name string) (Benchmark, bool) {
	for _, benchmark := range benchmarks {
		if benchmark.Name == name {
			return benchmark, true
		}
	}
	return Benchmark{}, false
}
