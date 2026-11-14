// Command compare checks one set of results against another by the comparability rule of
// spec §6 and prints one line per compared quantity. It exits 0 when every check passes,
// 1 when any fails, and 2 when it is used wrongly.
//
// Usage:
//
//	compare -label TEXT -rule same-draws|different-draws|same-program -reference FILE -candidate FILE
//	        [-reference-medians DIR -candidate-medians DIR]
//	compare -label TEXT -formula -candidate FILE -data DIR
//
// A results file holds one record per line (package record); lines starting with "#" and the
// header line are skipped. A medians folder holds, per data set, a file NAME.medians with
// one resample median per line.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/mioluigi/go-applied-statistics/dataset"
	"github.com/mioluigi/go-applied-statistics/internal/compare"
	"github.com/mioluigi/go-applied-statistics/record"
)

// request holds the command-line settings.
type request struct {
	label, rule                        string
	reference, candidate               string
	referenceMedians, candidateMedians string
	formula                            bool
	dataFolder                         string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses the request, performs the checks and prints them.
func run(arguments []string, stdout, stderr io.Writer) int {
	var settings request
	flags := flag.NewFlagSet("compare", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&settings.label, "label", "", "name of this comparison, printed on every line")
	flags.StringVar(&settings.rule, "rule", "", "same-draws, different-draws or same-program")
	flags.StringVar(&settings.reference, "reference", "", "results file to compare against")
	flags.StringVar(&settings.candidate, "candidate", "", "results file being checked")
	flags.StringVar(&settings.referenceMedians, "reference-medians", "", "folder of the reference's resample medians")
	flags.StringVar(&settings.candidateMedians, "candidate-medians", "", "folder of the candidate's resample medians")
	flags.BoolVar(&settings.formula, "formula", false, "check the candidate's SE of the mean against theory (needs -data)")
	flags.StringVar(&settings.dataFolder, "data", "", "folder of the data sets (for -formula)")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	checks, err := perform(settings)
	if err != nil {
		fmt.Fprintln(stderr, "compare:", err)
		return 2
	}
	failures := 0
	for _, check := range checks {
		fmt.Fprintf(stdout, "%s\t%s\n", settings.label, check)
		if !check.Pass {
			failures++
		}
	}
	if failures > 0 {
		fmt.Fprintf(stderr, "compare: %s: %d of %d checks FAILED\n", settings.label, failures, len(checks))
		return 1
	}
	return 0
}

// perform runs the comparison the request asks for.
func perform(settings request) ([]compare.Check, error) {
	candidates, err := readRecords(settings.candidate)
	if err != nil {
		return nil, err
	}
	if settings.formula {
		return formulaChecks(candidates, settings.dataFolder)
	}
	rules := map[string]compare.Rule{
		"same-draws": compare.SameDraws, "different-draws": compare.DifferentDraws, "same-program": compare.SameProgram,
	}
	rule, known := rules[settings.rule]
	if !known {
		return nil, fmt.Errorf("-rule is %q; choose same-draws, different-draws or same-program", settings.rule)
	}
	references, err := readRecords(settings.reference)
	if err != nil {
		return nil, err
	}
	var checks []compare.Check
	for _, name := range sortedNames(references) {
		reference := references[name]
		candidate, found := candidates[reference.Dataset]
		if !found {
			return nil, fmt.Errorf("%s has no result for %s", settings.candidate, reference.Dataset)
		}
		recordChecks, err := compare.Records(reference, candidate, rule)
		if err != nil {
			return nil, err
		}
		checks = append(checks, recordChecks...)
		if settings.referenceMedians != "" {
			check, err := medianCheck(reference.Dataset, settings.referenceMedians, settings.candidateMedians)
			if err != nil {
				return nil, err
			}
			checks = append(checks, check)
		}
	}
	return checks, nil
}

// formulaChecks compares each candidate's SE of the mean with the theoretical value.
func formulaChecks(candidates map[string]record.Record, dataFolder string) ([]compare.Check, error) {
	var checks []compare.Check
	for _, name := range sortedNames(candidates) {
		data, err := dataset.ReadFile(filepath.Join(dataFolder, name))
		if err != nil {
			return nil, err
		}
		checks = append(checks, compare.MeanFormula(candidates[name], data))
	}
	return checks, nil
}

// medianCheck compares the saved resample medians of one data set from two folders.
func medianCheck(name, referenceFolder, candidateFolder string) (compare.Check, error) {
	reference, err := readNumbers(filepath.Join(referenceFolder, name+".medians"))
	if err != nil {
		return compare.Check{}, err
	}
	candidate, err := readNumbers(filepath.Join(candidateFolder, name+".medians"))
	if err != nil {
		return compare.Check{}, err
	}
	return compare.ResampleMedians(name, reference, candidate), nil
}

// sortedNames returns the data set names of a result set in alphabetical order, so output
// lines always appear in the same order.
func sortedNames(records map[string]record.Record) []string {
	names := make([]string, 0, len(records))
	for name := range records {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// readRecords reads a results file into a map keyed by data set.
func readRecords(path string) (map[string]record.Record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read results: %w", err)
	}
	defer file.Close()
	records := map[string]record.Record{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") || line == record.Header {
			continue
		}
		parsed, err := record.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		records[parsed.Dataset] = parsed
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s holds no results", path)
	}
	return records, scanner.Err()
}

// readNumbers reads one number per line.
func readNumbers(path string) ([]float64, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read medians: %w", err)
	}
	var numbers []float64
	for _, field := range strings.Fields(string(content)) {
		number, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		numbers = append(numbers, number)
	}
	return numbers, nil
}
