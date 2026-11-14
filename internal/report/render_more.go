package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// shapesTable compares R and Go on the two other shapes at n = 10,000.
func (inputs *Inputs) shapesTable() (string, error) {
	var rows [][]string
	for _, shape := range []string{"normal", ScalingShape, "mirrored-lognormal"} {
		group := "shapes"
		if shape == ScalingShape {
			group = "scaling"
		}
		dataset := shape + "-10000.csv"
		loop, foundLoop := Find(inputs.Summaries, group, "r-loop", dataset)
		goR, foundGo := Find(inputs.Summaries, group, "go-r", dataset)
		if !foundLoop || !foundGo {
			return "", fmt.Errorf("results/runs.tsv has no runs of r-loop and go-r on %s", dataset)
		}
		rows = append(rows, []string{shape, fmt.Sprint(loop.Count), seconds(loop.Mean), seconds(goR.Mean),
			times(loop.Mean / goR.Mean), milliseconds(loop.MeanComputeMs), milliseconds(goR.MeanComputeMs),
			times(loop.MeanComputeMs / goR.MeanComputeMs)})
	}
	return table([]string{"Shape (n = 10,000)", "Runs", "R whole process (s)", "Go whole process (s)", "R ÷ Go",
		"R compute (ms)", "Go compute (ms)", "R ÷ Go, computation"}, rows), nil
}

// courseComparison sets the course's unchanged R program beside Go doing the same volume of
// resampling. The two do not run the same program, and the text says so.
func (inputs *Inputs) courseComparison() (string, error) {
	course, found := Find(inputs.Summaries, "course", "r-course-program", "-")
	if !found {
		return "", fmt.Errorf("results/runs.tsv has no runs of the course program")
	}
	volume, err := inputs.benchmark("CourseStudyVolume")
	if err != nil {
		return "", err
	}
	goSeconds := volume.NanosPerOp / 1e9
	rows := [][]string{
		{"Course R program, unchanged (whole process)", fmt.Sprint(course.Count), seconds(course.Mean),
			megabytes(course.MeanFootprint)},
		{"Go, same resampling volume (benchmark, computation only)", fmt.Sprint(volume.Count), seconds(goSeconds),
			megabytes(volume.BytesPerOp)},
	}
	return table([]string{"Program", "Runs", "Mean time (s)", "Memory (MB: peak footprint; Go: allocated per run)"}, rows) +
		fmt.Sprintf("\nRatio: **%s**. Same volume of resampling (40,000 resamples, mean and median), not the same program.\n",
			times(course.Mean/goSeconds)), nil
}

// memoryTable shows peak memory per implementation and size, and how much of it is above
// the memory the language uses just to start.
func (inputs *Inputs) memoryTable() (string, error) {
	baselines := map[string]string{"r-loop": "r-empty", "r-boot": "r-boot-load",
		"go-r": "go-help", "go-block-1": "go-help", "go-block-multi": "go-help"}
	var rows [][]string
	for _, n := range Sizes {
		for _, implementation := range implementations {
			summary, err := inputs.scaling(implementation, n)
			if err != nil {
				return "", err
			}
			baseline, err := inputs.startup(baselines[implementation])
			if err != nil {
				return "", err
			}
			rows = append(rows, []string{thousands(n), inputs.label(implementation),
				megabytes(summary.MeanFootprint), megabytes(summary.Spread), megabytes(summary.MeanResident),
				megabytes(summary.MeanFootprint - baseline.MeanFootprint)})
		}
	}
	return table([]string{"n", "Implementation", "Peak footprint: mean (MB)", "SD (MB)", "Max resident: mean (MB)",
		"Footprint above the language's start-up (MB)"}, rows), nil
}

// benchmarkTables shows component benchmarks and the before/after pair of each
// performance step.
func (inputs *Inputs) benchmarkTables() (string, error) {
	workerNames := []string{}
	for _, benchmark := range inputs.Benchmarks {
		if strings.HasPrefix(benchmark.Name, "Workers/workers=") {
			workerNames = append(workerNames, benchmark.Name)
		}
	}
	if len(workerNames) < 2 {
		return "", fmt.Errorf("results/benchmarks.txt needs the Workers benchmark at two worker counts")
	}
	steps := []struct{ title, before, after string }{
		{"1. Median by selection instead of sorting (n = 10,000)", "Step1Median/before=sort", "Step1Median/after=select"},
		{"2. One resample buffer per worker (n = 10,000, B = 100)", "Step2Resample/before=allocate", "Step2Resample/after=reuse"},
		{"3. Index bit count without a logarithm (n = 100)", "Step3Index/before=logarithm/n=100", "Step3Index/after=bitlength/n=100"},
		{"   same, n = 100,000", "Step3Index/before=logarithm/n=100000", "Step3Index/after=bitlength/n=100000"},
		{"Whole engine, steps 1 and 2 (n = 1,000, B = 100)", "Engine/before/n=1000", "Engine/after/n=1000"},
		{"Whole engine, steps 1 and 2 (n = 100,000, B = 100)", "Engine/before/n=100000", "Engine/after/n=100000"},
		{"Parallel workers: " + strings.TrimPrefix(workerNames[len(workerNames)-1], "Workers/workers=") +
			" instead of 1 (block stream, n = 10,000, B = 1,000)", workerNames[0], workerNames[len(workerNames)-1]},
	}
	var stepRows [][]string
	for _, step := range steps {
		before, err := inputs.benchmark(step.before)
		if err != nil {
			return "", err
		}
		after, err := inputs.benchmark(step.after)
		if err != nil {
			return "", err
		}
		stepRows = append(stepRows, []string{step.title, duration(before.NanosPerOp), duration(after.NanosPerOp),
			times(before.NanosPerOp / after.NanosPerOp),
			fmt.Sprintf("%s → %s", bytesText(before.BytesPerOp), bytesText(after.BytesPerOp)),
			fmt.Sprintf("%.0f → %.0f", before.AllocsPerOp, after.AllocsPerOp)})
	}
	var componentRows [][]string
	for _, benchmark := range inputs.Benchmarks {
		componentRows = append(componentRows, []string{"`" + benchmark.Name + "`", fmt.Sprint(benchmark.Count),
			duration(benchmark.NanosPerOp), bytesText(benchmark.BytesPerOp), fmt.Sprintf("%.0f", benchmark.AllocsPerOp)})
	}
	return "**Performance steps, before and after** (median of repeated benchmark runs)\n\n" +
		table([]string{"Step", "Before", "After", "Speed-up", "Memory allocated per run", "Allocations per run"}, stepRows) +
		"\n**All component benchmarks** (`go test -bench . -benchmem`)\n\n" +
		table([]string{"Benchmark", "Repetitions", "Time per operation", "Bytes per operation", "Allocations per operation"}, componentRows), nil
}

// duration formats nanoseconds with a fitting unit.
func duration(nanos float64) string {
	switch {
	case nanos >= 1e9:
		return fmt.Sprintf("%.2f s", nanos/1e9)
	case nanos >= 1e6:
		return fmt.Sprintf("%.2f ms", nanos/1e6)
	case nanos >= 1e3:
		return fmt.Sprintf("%.1f µs", nanos/1e3)
	}
	return fmt.Sprintf("%.1f ns", nanos)
}

// bytesText formats a byte count with a fitting unit.
func bytesText(bytes float64) string {
	switch {
	case bytes >= 1e6:
		return fmt.Sprintf("%.1f MB", bytes/1e6)
	case bytes >= 1e3:
		return fmt.Sprintf("%.1f kB", bytes/1e3)
	}
	return fmt.Sprintf("%.0f B", bytes)
}

// machineBlock lists the machine record as it was saved by measure.sh.
func (inputs *Inputs) machineBlock() (string, error) {
	var out strings.Builder
	for _, line := range inputs.Machine {
		out.WriteString("- " + line + "\n")
	}
	return out.String(), nil
}

// countCodeLines counts the lines of a source file that are neither blank nor only a
// comment (comment marker "//" for Go, "#" for R).
func countCodeLines(path, commentMarker string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("cannot count lines: %w", err)
	}
	count := 0
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, commentMarker) {
			count++
		}
	}
	return count, nil
}

// goLines counts code lines in the Go files of one folder, product and tests apart.
func goLines(folder string) (product, tests int, err error) {
	paths, err := filepath.Glob(filepath.Join(folder, "*.go"))
	if err != nil || len(paths) == 0 {
		return 0, 0, fmt.Errorf("no Go files in %s", folder)
	}
	for _, path := range paths {
		lines, err := countCodeLines(path, "//")
		if err != nil {
			return 0, 0, err
		}
		if strings.HasSuffix(path, "_test.go") {
			tests += lines
		} else {
			product += lines
		}
	}
	return product, tests, nil
}

// effortTable answers "how easy is it to use Go in place of R" with a count: the lines each
// language needs for the same analysis, and which parts Go had to supply itself.
func (inputs *Inputs) effortTable() (string, error) {
	rLoop, err := countCodeLines(filepath.Join(inputs.Root, "r", "bootstrap-se.R"), "#")
	if err != nil {
		return "", err
	}
	rBoot, err := countCodeLines(filepath.Join(inputs.Root, "r", "bootstrap-se-boot.R"), "#")
	if err != nil {
		return "", err
	}
	rows := [][]string{
		{"R: `r/bootstrap-se.R` (base R)", fmt.Sprint(rLoop), "–", "Everything else is built into R"},
		{"R: `r/bootstrap-se-boot.R` (`boot` package)", fmt.Sprint(rBoot), "–", "Resampling comes from the package"},
	}
	parts := []struct{ folder, name, supplies string }{
		{"rrand", "Go: `rrand`", "R's random generator and `sample()` (built into R)"},
		{"stats", "Go: `stats`", "`mean`, `sd`, `median` (built into R)"},
		{"bootstrap", "Go: `bootstrap`", "The resampling loop, the streams, parallel workers"},
		{"dataset", "Go: `dataset`", "Reading and checking the data file (`scan` in R)"},
		{"record", "Go: `record`", "The shared output line"},
		{".", "Go: `main.go` (command line)", "Flags, validation, logging, profiles"},
	}
	totalProduct, totalTests := 0, 0
	for _, part := range parts {
		product, tests, err := goLines(filepath.Join(inputs.Root, part.folder))
		if err != nil {
			return "", err
		}
		totalProduct, totalTests = totalProduct+product, totalTests+tests
		rows = append(rows, []string{part.name, fmt.Sprint(product), fmt.Sprint(tests), part.supplies})
	}
	rows = append(rows, []string{"**Go, total**", fmt.Sprintf("**%d**", totalProduct), fmt.Sprintf("**%d**", totalTests),
		fmt.Sprintf("**%s the base-R script**", times(float64(totalProduct)/float64(rLoop)))})
	return table([]string{"Code", "Lines of code", "Lines of tests", "What it supplies"}, rows) +
		"\nLines counted: not blank and not only a comment.\n", nil
}
