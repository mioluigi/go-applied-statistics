// Package compare applies the comparability rule of spec §6 to two results for the same
// data set: which quantities must be identical, which may differ in the last digits, and
// which may differ by Monte Carlo error.
package compare

import (
	"fmt"
	"math"

	"github.com/mioluigi/go-applied-statistics/record"
	"github.com/mioluigi/go-applied-statistics/stats"
)

// Limits on the relative difference between two results (spec §6).
const (
	Identical       = 0     // same float64
	SummationLimit  = 1e-12 // quantities that are sums: they may differ in the last bits
	MonteCarloLimit = 0.15  // standard errors from different random draws, at B = 1,000 (spike 002)
	TheoryLimit     = 0.10  // a bootstrap SE of the mean against its exact theoretical value
)

// Rule says how two results are expected to relate.
type Rule int

// The three relations between two runs on the same data, B and seed.
const (
	// SameDraws: both used the same resamples (R's sample() loop and Go's R stream).
	SameDraws Rule = iota
	// DifferentDraws: correct bootstraps with different random draws (block stream, boot).
	DifferentDraws
	// SameProgram: the same stream with a different worker count; everything is identical.
	SameProgram
)

// A Check is one compared quantity and its verdict.
type Check struct {
	Dataset    string
	Quantity   string
	Reference  float64
	Candidate  float64
	Difference float64 // relative difference; for counts, the number of differing items
	Limit      float64
	Pass       bool
}

// String formats a check as one tab-separated line.
func (c Check) String() string {
	verdict := "FAIL"
	if c.Pass {
		verdict = "PASS"
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%.3g\t%.3g\t%s", c.Dataset, c.Quantity,
		record.Exact(c.Reference), record.Exact(c.Candidate), c.Difference, c.Limit, verdict)
}

// relativeDifference is |candidate - reference| / |reference|, or the absolute difference
// when the reference is zero.
func relativeDifference(reference, candidate float64) float64 {
	if reference == 0 {
		return math.Abs(candidate)
	}
	return math.Abs(candidate-reference) / math.Abs(reference)
}

// quantity builds the check of one number against a limit.
func quantity(dataset, name string, reference, candidate, limit float64) Check {
	difference := relativeDifference(reference, candidate)
	return Check{
		Dataset: dataset, Quantity: name, Reference: reference, Candidate: candidate,
		Difference: difference, Limit: limit, Pass: difference <= limit,
	}
}

// Records compares two results for one data set under a rule. It returns an error when the
// two records do not describe the same run (different data set, B or seed), because then
// no comparison is meaningful.
func Records(reference, candidate record.Record, rule Rule) ([]Check, error) {
	if reference.Dataset != candidate.Dataset || reference.Replicates != candidate.Replicates ||
		reference.Seed != candidate.Seed {
		return nil, fmt.Errorf("records describe different runs: %s B=%d seed=%d against %s B=%d seed=%d",
			reference.Dataset, reference.Replicates, reference.Seed,
			candidate.Dataset, candidate.Replicates, candidate.Seed)
	}
	meanLimit, standardErrorLimit := SummationLimit, SummationLimit
	switch rule {
	case DifferentDraws:
		standardErrorLimit = MonteCarloLimit
	case SameProgram:
		meanLimit, standardErrorLimit = Identical, Identical
	}
	name := reference.Dataset
	return []Check{
		quantity(name, "n", float64(reference.Observations), float64(candidate.Observations), Identical),
		quantity(name, "sample median", reference.SampleMedian, candidate.SampleMedian, Identical),
		quantity(name, "sample mean", reference.SampleMean, candidate.SampleMean, meanLimit),
		quantity(name, "SE of the median", reference.SEMedian, candidate.SEMedian, standardErrorLimit),
		quantity(name, "SE of the mean", reference.SEMean, candidate.SEMean, standardErrorLimit),
	}, nil
}

// ResampleMedians checks that two lists of resample medians are identical, value by value.
// The check's Difference is the number of positions that differ (a length mismatch counts
// every missing position).
func ResampleMedians(dataset string, reference, candidate []float64) Check {
	different := max(len(reference), len(candidate)) - min(len(reference), len(candidate))
	for i := 0; i < min(len(reference), len(candidate)); i++ {
		if reference[i] != candidate[i] {
			different++
		}
	}
	return Check{
		Dataset: dataset, Quantity: fmt.Sprintf("resample medians differing (of %d)", len(reference)),
		Reference: 0, Candidate: float64(different), Difference: float64(different),
		Limit: Identical, Pass: different == 0 && len(reference) > 0,
	}
}

// MeanFormula checks a bootstrap standard error of the mean against the value theory gives
// for any data: sqrt((n-1)/n) * s / sqrt(n), where s is the sample standard deviation. It is
// an answer known independently of every implementation being compared.
func MeanFormula(result record.Record, data []float64) Check {
	count := float64(len(data))
	theory := math.Sqrt((count-1)/count) * stats.StandardDeviation(data) / math.Sqrt(count)
	return quantity(result.Dataset, "SE of the mean against sqrt((n-1)/n)*s/sqrt(n)", theory, result.SEMean, TheoryLimit)
}
