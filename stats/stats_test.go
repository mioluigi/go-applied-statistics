package stats

import (
	"math"
	"math/big"
	"testing"

	"github.com/mioluigi/go-applied-statistics/rrand"
)

func TestMedianKnownAnswers(t *testing.T) {
	cases := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"one value", []float64{7}, 7},
		{"two values: mean of both", []float64{1, 4}, 2.5},
		{"odd count, unsorted", []float64{9, 1, 5}, 5},
		{"even count, unsorted (R: median(c(1,2,3,4)) = 2.5)", []float64{4, 1, 3, 2}, 2.5},
		{"ties in the middle", []float64{2, 2, 2, 1, 3}, 2},
		{"even count with tied middle", []float64{1, 5, 5, 9}, 5},
		{"negative values", []float64{-3, -1, -2}, -2},
		{"outlier does not move it", []float64{1, 2, 3, 4, 1e9}, 3},
	}
	for _, testCase := range cases {
		if got := Median(testCase.values); got != testCase.want {
			t.Errorf("%s: got %v, want %v", testCase.name, got, testCase.want)
		}
	}
}

func TestMedianLeavesItsArgumentUnchanged(t *testing.T) {
	values := []float64{3, 1, 2}
	Median(values)
	if values[0] != 3 || values[1] != 1 || values[2] != 2 {
		t.Errorf("argument reordered to %v", values)
	}
}

func TestMeanAndStandardDeviationKnownAnswers(t *testing.T) {
	// Hand-computed: mean 5; squared deviations 9+1+1+1+0+0+4+16 = 32; 32/7; sqrt.
	values := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	if got := Mean(values); got != 5 {
		t.Errorf("mean: got %v, want 5", got)
	}
	if got, want := StandardDeviation(values), math.Sqrt(32.0/7.0); got != want {
		t.Errorf("standard deviation: got %v, want %v (n - 1 denominator)", got, want)
	}
	if got := StandardDeviation([]float64{1, 3}); got != math.Sqrt2 {
		t.Errorf("two values: got %v, want sqrt(2)", got)
	}
}

func TestTooFewValuesGiveNaN(t *testing.T) {
	if !math.IsNaN(Mean(nil)) || !math.IsNaN(Median(nil)) {
		t.Error("mean and median of no values should be NaN")
	}
	if !math.IsNaN(StandardDeviation([]float64{1})) {
		t.Error("standard deviation of one value should be NaN")
	}
}

// exactMeanAndVariance computes the mean and the n-1 variance in exact rational arithmetic,
// independently of the code under test.
func exactMeanAndVariance(values []float64) (mean, variance float64) {
	count := big.NewRat(int64(len(values)), 1)
	sum := new(big.Rat)
	for _, value := range values {
		sum.Add(sum, new(big.Rat).SetFloat64(value))
	}
	exactMean := new(big.Rat).Quo(sum, count)

	sumOfSquares := new(big.Rat)
	for _, value := range values {
		deviation := new(big.Rat).Sub(new(big.Rat).SetFloat64(value), exactMean)
		sumOfSquares.Add(sumOfSquares, deviation.Mul(deviation, deviation))
	}
	exactVariance := sumOfSquares.Quo(sumOfSquares, big.NewRat(int64(len(values)-1), 1))

	mean, _ = exactMean.Float64()
	variance, _ = exactVariance.Float64()
	return mean, variance
}

// TestAccuracyOnBadlyConditionedData uses 100,000 values near one million with a spread of
// about one: the case where a careless variance formula loses most of its digits
// (spec NFR-8: mean to 1e-14 relative, standard deviation to 1e-12).
func TestAccuracyOnBadlyConditionedData(t *testing.T) {
	generator := rrand.New(431)
	values := make([]float64, 100000)
	for i := range values {
		values[i] = 1e6 + math.Round((generator.Float64()*4-2)*1e4)/1e4
	}
	wantMean, wantVariance := exactMeanAndVariance(values)
	wantDeviation := math.Sqrt(wantVariance)

	const meanTolerance, deviationTolerance = 1e-14, 1e-12
	if got := Mean(values); math.Abs(got-wantMean)/wantMean > meanTolerance {
		t.Errorf("mean: got %.17g, exact %.17g", got, wantMean)
	}
	if got := StandardDeviation(values); math.Abs(got-wantDeviation)/wantDeviation > deviationTolerance {
		t.Errorf("standard deviation: got %.17g, exact %.17g", got, wantDeviation)
	}
}

func BenchmarkMedian(b *testing.B) {
	generator := rrand.New(431)
	values := make([]float64, 10000)
	for i := range values {
		values[i] = generator.Float64()
	}
	b.ReportAllocs()
	for b.Loop() {
		Median(values)
	}
}
