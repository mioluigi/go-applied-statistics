// Command bootse estimates the bootstrap standard errors of the median and the mean of a
// data set, and can reproduce R's result for the same data and seed number for number.
//
// It is a thin layer: it reads flags, checks them, calls the reusable packages (dataset,
// bootstrap, stats) and prints. Results go to standard output, log lines to standard error.
package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/mioluigi/go-applied-statistics/bootstrap"
	"github.com/mioluigi/go-applied-statistics/dataset"
	"github.com/mioluigi/go-applied-statistics/record"
	"github.com/mioluigi/go-applied-statistics/stats"
)

// Exit codes (spec FR-9).
const (
	exitSuccess = 0
	exitFailure = 1 // the run could not be completed (for example an output file cannot be written)
	exitInvalid = 2 // the request was not usable: bad flag, bad value, bad data file
)

// builtInName is the committed data set a run without -data analyses.
const builtInName = "lognormal-1000.csv"

//go:embed data/lognormal-1000.csv
var builtInData []byte

// errInvalid marks a request the user must correct; it maps to exitInvalid.
var errInvalid = errors.New("invalid request")

// options holds the checked command-line settings.
type options struct {
	dataPath    string
	replicates  int
	seed        uint32
	streamName  string
	workers     int
	format      string
	mediansPath string
	cpuProfile  string
	memProfile  string
	quiet       bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the whole program with its inputs and outputs passed in, so it can be tested.
func run(arguments []string, stdout, stderr io.Writer) int {
	settings, err := parseOptions(arguments, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	if err == nil {
		err = analyse(settings, stdout, stderr)
	}
	if err == nil {
		return exitSuccess
	}
	fmt.Fprintln(stderr, "bootse:", err)
	if errors.Is(err, errInvalid) || errors.Is(err, dataset.ErrInvalid) || errors.Is(err, bootstrap.ErrInvalid) {
		fmt.Fprintln(stderr, "Run \"bootse -h\" for the list of options.")
		return exitInvalid
	}
	return exitFailure
}

// parseOptions reads and checks the flags. Every problem is reported in plain words.
func parseOptions(arguments []string, stderr io.Writer) (options, error) {
	var settings options
	var seed uint64
	flags := flag.NewFlagSet("bootse", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&settings.dataPath, "data", "", "data file: a header line, then one number per line (default: built-in "+builtInName+")")
	flags.IntVar(&settings.replicates, "replicates", 1000, "number of bootstrap resamples B (at least 2)")
	flags.Uint64Var(&seed, "seed", 431, "seed of the random stream (0 to 4294967295)")
	flags.StringVar(&settings.streamName, "stream", "r", "random stream: r (reproduces R's sample(), 1 worker) or block (parallel)")
	flags.IntVar(&settings.workers, "workers", 0, "goroutines (default: 1 for -stream r, the number of CPUs for -stream block)")
	flags.StringVar(&settings.format, "format", "text", "output: text (for people) or tsv (one record line, as the R scripts write)")
	flags.StringVar(&settings.mediansPath, "medians", "", "also write the B resample medians to this file, one per line")
	flags.StringVar(&settings.cpuProfile, "cpuprofile", "", "write a CPU profile to this file")
	flags.StringVar(&settings.memProfile, "memprofile", "", "write a memory profile to this file")
	flags.BoolVar(&settings.quiet, "quiet", false, "do not write log lines to standard error")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return settings, err
		}
		return settings, fmt.Errorf("%w: %v", errInvalid, err)
	}

	workersGiven := false
	flags.Visit(func(given *flag.Flag) { workersGiven = workersGiven || given.Name == "workers" })
	settings.seed = uint32(seed)
	return settings, checkOptions(&settings, flags.Args(), seed, workersGiven)
}

// checkOptions validates the parsed values and fills in the default worker count.
func checkOptions(settings *options, extra []string, seed uint64, workersGiven bool) error {
	switch {
	case len(extra) > 0:
		return fmt.Errorf("%w: unexpected argument %q (the data file is given with -data)", errInvalid, extra[0])
	case settings.replicates < 2:
		return fmt.Errorf("%w: -replicates is %d; at least 2 resamples are needed for a standard error", errInvalid, settings.replicates)
	case seed > 1<<32-1:
		return fmt.Errorf("%w: -seed is %d; it must be between 0 and 4294967295", errInvalid, seed)
	case settings.streamName != "r" && settings.streamName != "block":
		return fmt.Errorf("%w: -stream is %q; choose r or block", errInvalid, settings.streamName)
	case settings.format != "text" && settings.format != "tsv":
		return fmt.Errorf("%w: -format is %q; choose text or tsv", errInvalid, settings.format)
	case workersGiven && settings.workers < 1:
		return fmt.Errorf("%w: -workers is %d; at least 1 is needed", errInvalid, settings.workers)
	case workersGiven && settings.workers > 1 && settings.streamName == "r":
		return fmt.Errorf("%w: -stream r draws R's numbers in order and runs with 1 worker; "+
			"use -stream block with -workers %d", errInvalid, settings.workers)
	}
	if !workersGiven {
		settings.workers = 1
		if settings.streamName == "block" {
			settings.workers = runtime.NumCPU()
		}
	}
	return nil
}

// analyse reads the data, runs the bootstrap and writes the result.
func analyse(settings options, stdout, stderr io.Writer) error {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if !settings.quiet {
		logger = slog.New(slog.NewTextHandler(stderr, nil))
	}

	readStart := time.Now()
	name, data, err := readData(settings.dataPath)
	if err != nil {
		return err
	}
	logger.Info("data read", "dataset", name, "n", len(data), "read_ms", milliseconds(time.Since(readStart)))

	stopProfile, err := startCPUProfile(settings.cpuProfile)
	if err != nil {
		return err
	}
	computeStart := time.Now()
	config := bootstrap.Config{Replicates: settings.replicates, Stream: newStream(settings), Workers: settings.workers}
	// The mean is listed first: the median reorders the resample it is given.
	values, err := bootstrap.Replicates(data, []bootstrap.Statistic{stats.Mean, stats.MedianInPlace}, config)
	stopProfile()
	if err != nil {
		return err
	}
	means, medians := values[0], values[1]
	result := record.Record{
		Dataset: name, Observations: len(data), Replicates: settings.replicates,
		Seed: settings.seed, Stream: settings.streamName,
		SEMedian: stats.StandardDeviation(medians), SEMean: stats.StandardDeviation(means),
	}
	logger.Info("bootstrap computed", "replicates", settings.replicates, "stream", settings.streamName,
		"workers", settings.workers, "compute_ms", milliseconds(time.Since(computeStart)))

	result.SampleMedian, result.SampleMean = stats.Median(data), stats.Mean(data)
	if err := writeOutputs(settings, result, medians, stdout); err != nil {
		return err
	}
	return writeMemoryProfile(settings.memProfile)
}

// readData returns the data set's name and values: the file at path, or the built-in copy.
func readData(path string) (string, []float64, error) {
	if path == "" {
		data, err := dataset.Read(bytes.NewReader(builtInData), builtInName)
		return builtInName, data, err
	}
	data, err := dataset.ReadFile(path)
	if err != nil && !errors.Is(err, dataset.ErrInvalid) {
		// A file that cannot be opened is the user's to fix, like any other bad input.
		err = fmt.Errorf("%w: %v", errInvalid, err)
	}
	return filepath.Base(path), data, err
}

// newStream returns the random stream named on the command line.
func newStream(settings options) bootstrap.Stream {
	if settings.streamName == "block" {
		return bootstrap.BlockStream(settings.seed)
	}
	return bootstrap.RStream(settings.seed)
}

// milliseconds converts a duration to milliseconds with three decimals, for the log.
func milliseconds(elapsed time.Duration) float64 {
	return float64(elapsed.Microseconds()) / 1000
}

// writeOutputs prints the result and, if asked, saves the resample medians.
func writeOutputs(settings options, result record.Record, medians []float64, stdout io.Writer) error {
	if settings.format == "tsv" {
		fmt.Fprintln(stdout, result)
	} else {
		fmt.Fprint(stdout, summary(result))
	}
	if settings.mediansPath == "" {
		return nil
	}
	var lines strings.Builder
	for _, median := range medians {
		lines.WriteString(record.Exact(median))
		lines.WriteByte('\n')
	}
	if err := os.WriteFile(settings.mediansPath, []byte(lines.String()), 0o644); err != nil {
		return fmt.Errorf("cannot write the resample medians: %w", err)
	}
	return nil
}

// summary describes a result in plain words: what was estimated, from which data, and how.
func summary(result record.Record) string {
	streamNote := "reproduces R's sample() for this seed"
	if result.Stream == "block" {
		streamNote = "parallel stream; agrees with R up to Monte Carlo error"
	}
	return fmt.Sprintf(`Bootstrap standard errors
  Data set:       %s (n = %d observations)
  Resamples:      B = %d, seed %d, stream %s (%s)
  Sample median:  %-12.6g bootstrap SE of the median: %.6g
  Sample mean:    %-12.6g bootstrap SE of the mean:   %.6g
`, result.Dataset, result.Observations, result.Replicates, result.Seed, result.Stream, streamNote,
		result.SampleMedian, result.SEMedian, result.SampleMean, result.SEMean)
}

// startCPUProfile starts a CPU profile if a path was given and returns the function that
// stops it; with no path, both do nothing.
func startCPUProfile(path string) (stop func(), err error) {
	if path == "" {
		return func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("cannot write the CPU profile: %w", err)
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("cannot start the CPU profile: %w", err)
	}
	return func() {
		pprof.StopCPUProfile()
		file.Close()
	}, nil
}

// writeMemoryProfile saves a heap profile if a path was given.
func writeMemoryProfile(path string) error {
	if path == "" {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot write the memory profile: %w", err)
	}
	defer file.Close()
	runtime.GC() // so the profile shows what is still in use, not garbage
	if err := pprof.Lookup("allocs").WriteTo(file, 0); err != nil {
		return fmt.Errorf("cannot write the memory profile: %w", err)
	}
	return nil
}
