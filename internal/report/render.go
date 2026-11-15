package report

import (
	"fmt"
	"math"
	"strings"
)

// A Block is one generated piece of the README: a name and its Markdown.
type Block struct {
	Name string
	Body string
}

// implementations lists the five timed implementations in display order.
var implementations = []string{"r-loop", "r-boot", "go-r", "go-block-1", "go-block-multi"}

// label returns the display name of an implementation.
func (inputs *Inputs) label(implementation string) string {
	switch implementation {
	case "r-loop":
		return "R, `sample()` loop"
	case "r-boot":
		return "R, `boot` package"
	case "go-r":
		return "Go, R stream (1 core)"
	case "go-block-1":
		return "Go, block stream (1 worker)"
	case "go-block-multi":
		return "Go, block stream (" + inputs.machineValue("parallel workers") + " workers)"
	case "r-empty":
		return "R, empty script"
	case "r-boot-load":
		return "R, loading `boot` only"
	case "go-help":
		return "Go, printing its help"
	case "r-course-program":
		return "Course R program, unchanged"
	}
	return implementation
}

// Formatting helpers. Each returns text ready for a table cell.

func seconds(value float64) string { return fmt.Sprintf("%.3f", value) }

func percent(share float64) string { return fmt.Sprintf("%.1f %%", 100*share) }

func megabytes(bytes float64) string { return fmt.Sprintf("%.1f", bytes/1e6) }

func thousands(n int) string {
	text := fmt.Sprintf("%d", n)
	for position := len(text) - 3; position > 0; position -= 3 {
		text = text[:position] + "," + text[position:]
	}
	return text
}

// times formats a ratio such as "12.3×"; large ratios lose their decimal.
func times(ratio float64) string {
	if ratio >= 100 {
		return fmt.Sprintf("%.0f×", ratio)
	}
	return fmt.Sprintf("%.1f×", ratio)
}

// milliseconds formats a duration in ms, or a dash when the program logs none.
func milliseconds(value float64) string {
	if math.IsNaN(value) {
		return "–"
	}
	return fmt.Sprintf("%.1f", value)
}

// dollars formats a cost; amounts under one dollar keep four decimals.
func dollars(value float64) string {
	if value < 1 {
		return fmt.Sprintf("$%.4f", value)
	}
	return fmt.Sprintf("$%.2f", value)
}

// table renders a Markdown table from a header row and data rows.
func table(header []string, rows [][]string) string {
	var out strings.Builder
	out.WriteString("| " + strings.Join(header, " | ") + " |\n")
	out.WriteString("|" + strings.Repeat("---|", len(header)) + "\n")
	for _, row := range rows {
		out.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}
	return out.String()
}

// precisionNote flags a mean whose standard error exceeds 2 % of it (spec NFR-2).
func precisionNote(summary Summary) string {
	if summary.RelativeError > 0.02 {
		return percent(summary.RelativeError) + " ⚠"
	}
	return percent(summary.RelativeError)
}

// Blocks generates every block of the report, in the order the README uses them.
func (inputs *Inputs) Blocks() ([]Block, error) {
	generators := []struct {
		name     string
		generate func() (string, error)
	}{
		{"headline", inputs.headline},
		{"mersenne", inputs.generatorComparison},
		{"example", inputs.exampleAnalysis},
		{"correctness", inputs.correctness},
		{"startup", inputs.startupTable},
		{"scaling", inputs.scalingTable},
		{"ratios", inputs.ratiosTable},
		{"shapes", inputs.shapesTable},
		{"course", inputs.courseComparison},
		{"memory", inputs.memoryTable},
		{"benchmarks", inputs.benchmarkTables},
		{"step4", inputs.stepFourOfficial},
		{"cost", inputs.costTables},
		{"effort", inputs.effortTable},
		{"predictions", inputs.predictions},
		{"machine", inputs.machineBlock},
	}
	blocks := make([]Block, 0, len(generators))
	for _, generator := range generators {
		body, err := generator.generate()
		if err != nil {
			return nil, fmt.Errorf("block %s: %w", generator.name, err)
		}
		blocks = append(blocks, Block{Name: generator.name, Body: body})
	}
	return blocks, nil
}

// correctness summarises results/comparison.tsv: one row per comparison.
func (inputs *Inputs) correctness() (string, error) {
	type tally struct {
		checks, passed int
		largest        float64
		limits         map[string]bool
	}
	tallies := map[string]*tally{}
	var order []string
	for _, check := range inputs.Checks {
		current, seen := tallies[check.Comparison]
		if !seen {
			current = &tally{limits: map[string]bool{}}
			tallies[check.Comparison] = current
			order = append(order, check.Comparison)
		}
		current.checks++
		if check.Pass {
			current.passed++
		}
		// The medians check reports a count, not a relative difference; it is shown apart.
		if !strings.HasPrefix(check.Quantity, "resample medians") {
			current.largest = math.Max(current.largest, check.Difference)
		}
	}
	var rows [][]string
	for _, name := range order {
		current := tallies[name]
		rows = append(rows, []string{name, fmt.Sprintf("%d of %d", current.passed, current.checks),
			fmt.Sprintf("%.2g", current.largest)})
	}
	body := table([]string{"Comparison (12 data sets each)", "Checks passed", "Largest relative difference"}, rows)
	identical, compared := inputs.medianIdentity()
	body += fmt.Sprintf("\nResample medians compared between R and Go: **%s, of which %s identical** (same 64-bit value).\n",
		thousands(compared), thousands(identical))
	return body, nil
}

// medianIdentity counts the resample medians compared between R and Go and how many were
// identical, from the "resample medians differing (of B)" checks.
func (inputs *Inputs) medianIdentity() (identical, compared int) {
	for _, check := range inputs.Checks {
		var replicates int
		if _, err := fmt.Sscanf(check.Quantity, "resample medians differing (of %d)", &replicates); err == nil {
			compared += replicates
			identical += replicates - int(check.Difference)
		}
	}
	return identical, compared
}

// checksPassed counts all comparison checks and those that passed.
func (inputs *Inputs) checksPassed() (passed, total int) {
	for _, check := range inputs.Checks {
		total++
		if check.Pass {
			passed++
		}
	}
	return passed, total
}

// startupTable shows what starting each program costs, in time and memory.
func (inputs *Inputs) startupTable() (string, error) {
	var rows [][]string
	for _, implementation := range []string{"r-empty", "r-boot-load", "go-help"} {
		summary, err := inputs.startup(implementation)
		if err != nil {
			return "", err
		}
		rows = append(rows, []string{inputs.label(implementation), fmt.Sprint(summary.Count),
			seconds(summary.Mean), seconds(summary.Median), seconds(summary.Slowest), seconds(summary.Deviation),
			precisionNote(summary), megabytes(summary.MeanFootprint), megabytes(summary.MeanResident)})
	}
	return table([]string{"Start-up condition", "Runs", "Mean (s)", "Median (s)", "Slowest run (s)", "SD (s)",
		"Std. error of mean", "Peak footprint (MB)", "Max resident (MB)"}, rows), nil
}

// scalingTable shows whole-process time of every implementation at every size.
func (inputs *Inputs) scalingTable() (string, error) {
	var rows [][]string
	for _, n := range Sizes {
		for _, implementation := range implementations {
			summary, err := inputs.scaling(implementation, n)
			if err != nil {
				return "", err
			}
			rows = append(rows, []string{thousands(n), inputs.label(implementation), fmt.Sprint(summary.Count),
				seconds(summary.Mean), seconds(summary.Median), seconds(summary.Deviation), precisionNote(summary),
				milliseconds(summary.MeanReadMs), milliseconds(summary.MeanComputeMs)})
		}
	}
	return table([]string{"n", "Implementation", "Runs", "Whole process: mean (s)", "Median (s)", "SD (s)",
		"Std. error of mean", "Read (ms)", "Compute (ms)"}, rows), nil
}

// ratiosTable puts R and Go side by side at each size: whole process, computation only,
// start-up share, and what parallel workers add.
func (inputs *Inputs) ratiosTable() (string, error) {
	empty, err := inputs.startup("r-empty")
	if err != nil {
		return "", err
	}
	var rows [][]string
	for _, n := range Sizes {
		summaries := map[string]Summary{}
		for _, implementation := range implementations {
			if summaries[implementation], err = inputs.scaling(implementation, n); err != nil {
				return "", err
			}
		}
		loop, goR := summaries["r-loop"], summaries["go-r"]
		rows = append(rows, []string{thousands(n),
			times(loop.Mean / goR.Mean), times(loop.Median / goR.Median),
			times(loop.MeanComputeMs / goR.MeanComputeMs),
			times(loop.MeanComputeMs / summaries["go-block-1"].MeanComputeMs),
			percent(empty.Mean / loop.Mean),
			times(summaries["go-block-1"].MeanComputeMs / summaries["go-block-multi"].MeanComputeMs),
			times(loop.Mean / summaries["go-block-multi"].Mean),
			times(summaries["r-boot"].Mean / loop.Mean)})
	}
	workers := inputs.machineValue("parallel workers")
	return table([]string{"n", "R ÷ Go (R stream), whole process, means", "same, medians",
		"R ÷ Go (R stream), computation only", "R ÷ Go (block stream, 1 worker), computation only",
		"R start-up as share of R's time", "Go: " + workers + " workers ÷ 1 worker, computation",
		"R ÷ Go with " + workers + " workers, whole process", "`boot` ÷ `sample()` loop"}, rows), nil
}
