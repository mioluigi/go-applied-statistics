// Package report turns the saved measurements (results/) and saved prices (prices/) into
// the tables, headline figures and cost estimates the README shows. Nothing in the README
// is typed by hand: every number comes from here (constitution 1.3).
package report

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/mioluigi/go-applied-statistics/stats"
)

// A Run is one timed execution of one program, from results/runs.tsv.
type Run struct {
	Group          string  // startup, scaling, shapes or course
	Implementation string  // r-loop, r-boot, go-r, go-block-1, go-block-multi, …
	Dataset        string  // data file, or "-" for start-up and course runs
	Observations   int     // n, or 0
	Seconds        float64 // whole-process wall time
	MaxResident    float64 // bytes: maximum resident set size
	PeakFootprint  float64 // bytes: peak memory footprint
	ReadMs         float64 // time reading the data, logged by the program (NaN if absent)
	ComputeMs      float64 // time evaluating the B replicates (NaN if absent)
}

// A Condition identifies runs that measured the same thing.
type Condition struct {
	Group          string
	Implementation string
	Dataset        string
}

// A Summary describes the repeated runs of one condition.
type Summary struct {
	Condition
	Observations int
	Count        int     // N, the number of runs
	Mean         float64 // seconds
	Median       float64 // seconds
	Deviation    float64 // standard deviation, seconds
	// RelativeError is the standard error of the mean as a share of the mean: how far the
	// reported mean may be from the true one.
	RelativeError float64
	MeanComputeMs float64 // NaN when the program logs none
	MeanReadMs    float64
	MeanFootprint float64 // bytes
	MeanResident  float64 // bytes
	Spread        float64 // standard deviation of the peak footprint, bytes
}

// readTable reads a tab-separated file, skipping comment lines, and returns its data rows
// (without the header) and the column positions by header name.
func readTable(path string) (rows [][]string, column map[string]int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if column == nil {
			column = map[string]int{}
			for position, name := range fields {
				column[name] = position
			}
			continue
		}
		if len(fields) != len(column) {
			return nil, nil, fmt.Errorf("%s: a row has %d fields, the header has %d", path, len(fields), len(column))
		}
		rows = append(rows, fields)
	}
	if column == nil {
		return nil, nil, fmt.Errorf("%s is empty", path)
	}
	return rows, column, scanner.Err()
}

// number parses a field; an empty field is NaN (the program logged no such value).
func number(text string) float64 {
	if text == "" {
		return math.NaN()
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return math.NaN()
	}
	return value
}

// ReadRuns reads results/runs.tsv.
func ReadRuns(path string) ([]Run, error) {
	rows, column, err := readTable(path)
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"group", "implementation", "dataset", "n", "real_s",
		"max_rss_bytes", "peak_footprint_bytes", "read_ms", "compute_ms"} {
		if _, found := column[name]; !found {
			return nil, fmt.Errorf("%s lacks the column %q", path, name)
		}
	}
	runs := make([]Run, 0, len(rows))
	for _, row := range rows {
		observations, _ := strconv.Atoi(row[column["n"]])
		run := Run{
			Group: row[column["group"]], Implementation: row[column["implementation"]],
			Dataset: row[column["dataset"]], Observations: observations,
			Seconds:     number(row[column["real_s"]]),
			MaxResident: number(row[column["max_rss_bytes"]]), PeakFootprint: number(row[column["peak_footprint_bytes"]]),
			ReadMs: number(row[column["read_ms"]]), ComputeMs: number(row[column["compute_ms"]]),
		}
		if math.IsNaN(run.Seconds) {
			return nil, fmt.Errorf("%s: a run of %s has no time", path, run.Implementation)
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// meanOf averages one field over runs, ignoring runs where it is absent; NaN if all are.
func meanOf(runs []Run, field func(Run) float64) float64 {
	var present []float64
	for _, run := range runs {
		if value := field(run); !math.IsNaN(value) {
			present = append(present, value)
		}
	}
	return stats.Mean(present)
}

// Summarize groups runs by condition and describes each group. The order of the result
// follows the first appearance of each condition in runs.
func Summarize(runs []Run) []Summary {
	grouped := map[Condition][]Run{}
	var order []Condition
	for _, run := range runs {
		condition := Condition{run.Group, run.Implementation, run.Dataset}
		if _, seen := grouped[condition]; !seen {
			order = append(order, condition)
		}
		grouped[condition] = append(grouped[condition], run)
	}
	summaries := make([]Summary, 0, len(order))
	for _, condition := range order {
		summaries = append(summaries, summarize(condition, grouped[condition]))
	}
	return summaries
}

// summarize describes the runs of one condition.
func summarize(condition Condition, runs []Run) Summary {
	seconds := make([]float64, len(runs))
	footprints := make([]float64, len(runs))
	for i, run := range runs {
		seconds[i], footprints[i] = run.Seconds, run.PeakFootprint
	}
	summary := Summary{
		Condition: condition, Observations: runs[0].Observations, Count: len(runs),
		Mean: stats.Mean(seconds), Median: stats.Median(seconds),
		MeanComputeMs: meanOf(runs, func(run Run) float64 { return run.ComputeMs }),
		MeanReadMs:    meanOf(runs, func(run Run) float64 { return run.ReadMs }),
		MeanFootprint: stats.Mean(footprints),
		MeanResident:  meanOf(runs, func(run Run) float64 { return run.MaxResident }),
	}
	if len(runs) >= 2 {
		summary.Deviation = stats.StandardDeviation(seconds)
		summary.Spread = stats.StandardDeviation(footprints)
		summary.RelativeError = summary.Deviation / math.Sqrt(float64(len(runs))) / summary.Mean
	}
	return summary
}

// Find returns the summary of one condition, and whether it exists.
func Find(summaries []Summary, group, implementation, dataset string) (Summary, bool) {
	position := slices.IndexFunc(summaries, func(summary Summary) bool {
		return summary.Condition == Condition{group, implementation, dataset}
	})
	if position < 0 {
		return Summary{}, false
	}
	return summaries[position], true
}
