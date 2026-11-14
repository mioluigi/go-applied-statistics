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

// MedianInPlace returns the same value as Median without sorting: it finds the middle value
// by selection, which takes time proportional to the number of values instead of n log n.
// It reorders its argument, so it suits a scratch copy such as a bootstrap resample.
func MedianInPlace(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	upperMiddle := len(values) / 2
	selectPosition(values, upperMiddle)
	if len(values)%2 == 1 {
		return values[upperMiddle]
	}
	// After selection everything before upperMiddle is no larger than it, so the lower
	// middle value is the largest of that part.
	lowerMiddle := slices.Max(values[:upperMiddle])
	return (lowerMiddle + values[upperMiddle]) / 2
}

// selectPosition reorders values so that values[position] holds what a full sort would put
// there, with nothing larger before it and nothing smaller after it (Hoare's quickselect).
func selectPosition(values []float64, position int) {
	low, high := 0, len(values)-1
	for low < high {
		pivot := medianOfThree(values[low], values[low+(high-low)/2], values[high])
		left, right := low, high
		for left <= right {
			for values[left] < pivot {
				left++
			}
			for values[right] > pivot {
				right--
			}
			if left <= right {
				values[left], values[right] = values[right], values[left]
				left++
				right--
			}
		}
		// Now values[low..right] <= pivot <= values[left..high]; keep the part with position.
		switch {
		case position <= right:
			high = right
		case position >= left:
			low = left
		default:
			return
		}
	}
}

// medianOfThree returns the middle of three values: a pivot that splits sorted and
// reverse-sorted input evenly.
func medianOfThree(a, b, c float64) float64 {
	return max(min(a, b), min(max(a, b), c))
}
