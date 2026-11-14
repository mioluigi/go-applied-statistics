// Package stats holds the few descriptive statistics the bootstrap needs, defined exactly as
// R defines them so that results can be compared with R's.
package stats

import (
	"math"
	"slices"
)

// Mean returns the arithmetic mean, or NaN for no values.
//
// Like R's mean(), it makes a second pass that adds the average of the deviations from the
// first estimate. The correction removes most of the rounding error a single sum leaves behind.
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	count := float64(len(values))
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	estimate := sum / count

	residual := 0.0
	for _, value := range values {
		residual += value - estimate
	}
	return estimate + residual/count
}

// StandardDeviation returns the sample standard deviation (denominator n - 1), as R's sd()
// does, or NaN for fewer than two values.
//
// It sums squared deviations from the mean. The shortcut "sum of squares minus n times the
// squared mean" is deliberately avoided: it loses most of its digits when the values are
// large compared with their spread.
func StandardDeviation(values []float64) float64 {
	if len(values) < 2 {
		return math.NaN()
	}
	mean := Mean(values)
	sumOfSquares := 0.0
	for _, value := range values {
		deviation := value - mean
		sumOfSquares += deviation * deviation
	}
	return math.Sqrt(sumOfSquares / float64(len(values)-1))
}

// Median returns the middle value, or for an even count the mean of the two middle values,
// as R's median() does. It returns NaN for no values and leaves its argument unchanged.
func Median(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	lowerMiddle, upperMiddle := (len(sorted)-1)/2, len(sorted)/2
	return (sorted[lowerMiddle] + sorted[upperMiddle]) / 2
}
