package report

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Sizes are the data-set sizes of the scaling study, smallest first.
var Sizes = []int{100, 1000, 10000, 100000}

// ScalingShape is the shape whose four sizes make the scaling study.
const ScalingShape = "lognormal"

// A ComparisonRow is one check from results/comparison.tsv.
type ComparisonRow struct {
	Comparison string
	Dataset    string
	Quantity   string
	Difference float64
	Limit      float64
	Pass       bool
}

// Inputs is everything the report is generated from.
type Inputs struct {
	Summaries  []Summary
	Benchmarks []Benchmark
	Prices     []Price
	Checks     []ComparisonRow
	Machine    []string // lines of results/machine.txt
	Root       string   // repository root, for counting lines of code
}

// Load reads the saved results and prices. root is the repository folder.
func Load(root string) (*Inputs, error) {
	results := filepath.Join(root, "results")
	runs, err := ReadRuns(filepath.Join(results, "runs.tsv"))
	if err != nil {
		return nil, err
	}
	inputs := &Inputs{Summaries: Summarize(runs), Root: root}
	if inputs.Benchmarks, err = ReadBenchmarks(filepath.Join(results, "benchmarks.txt")); err != nil {
		return nil, err
	}
	if inputs.Prices, err = ReadPrices(filepath.Join(root, "prices", "aws-ec2.tsv")); err != nil {
		return nil, err
	}
	if inputs.Checks, err = readComparison(filepath.Join(results, "comparison.tsv")); err != nil {
		return nil, err
	}
	if inputs.Machine, err = readLines(filepath.Join(results, "machine.txt")); err != nil {
		return nil, err
	}
	return inputs, nil
}

// readComparison reads results/comparison.tsv.
func readComparison(path string) ([]ComparisonRow, error) {
	rows, column, err := readTable(path)
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"comparison", "dataset", "quantity", "difference", "limit", "result"} {
		if _, found := column[name]; !found {
			return nil, fmt.Errorf("%s lacks the column %q", path, name)
		}
	}
	checks := make([]ComparisonRow, 0, len(rows))
	for _, row := range rows {
		checks = append(checks, ComparisonRow{
			Comparison: row[column["comparison"]], Dataset: row[column["dataset"]], Quantity: row[column["quantity"]],
			Difference: number(row[column["difference"]]), Limit: number(row[column["limit"]]),
			Pass: row[column["result"]] == "PASS",
		})
	}
	if len(checks) == 0 {
		return nil, fmt.Errorf("%s holds no checks", path)
	}
	return checks, nil
}

// readLines returns the non-empty lines of a text file.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

// machineValue returns the text after "key: " in the machine record, or "" if absent.
func (inputs *Inputs) machineValue(key string) string {
	for _, line := range inputs.Machine {
		if value, found := strings.CutPrefix(line, key+": "); found {
			return value
		}
	}
	return ""
}

// scaling returns the summary of one implementation at one size of the scaling study.
func (inputs *Inputs) scaling(implementation string, n int) (Summary, error) {
	dataset := fmt.Sprintf("%s-%d.csv", ScalingShape, n)
	summary, found := Find(inputs.Summaries, "scaling", implementation, dataset)
	if !found {
		return Summary{}, fmt.Errorf("results/runs.tsv has no scaling runs of %s on %s", implementation, dataset)
	}
	return summary, nil
}

// startup returns the summary of one start-up condition.
func (inputs *Inputs) startup(implementation string) (Summary, error) {
	summary, found := Find(inputs.Summaries, "startup", implementation, "-")
	if !found {
		return Summary{}, fmt.Errorf("results/runs.tsv has no start-up runs of %s", implementation)
	}
	return summary, nil
}

// price returns the saved price of the instance with the given vCPU count.
func (inputs *Inputs) price(vcpus int) (Price, error) {
	for _, price := range inputs.Prices {
		if price.VCPUs == vcpus {
			return price, nil
		}
	}
	return Price{}, fmt.Errorf("prices/aws-ec2.tsv has no instance with %d vCPU(s)", vcpus)
}

// benchmark returns one benchmark by name.
func (inputs *Inputs) benchmark(name string) (Benchmark, error) {
	benchmark, found := FindBenchmark(inputs.Benchmarks, name)
	if !found {
		return Benchmark{}, fmt.Errorf("results/benchmarks.txt has no benchmark %s", name)
	}
	return benchmark, nil
}
