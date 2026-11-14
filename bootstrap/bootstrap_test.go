package bootstrap

import (
	"bufio"
	"errors"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/mioluigi/go-applied-statistics/dataset"
	"github.com/mioluigi/go-applied-statistics/record"
	"github.com/mioluigi/go-applied-statistics/stats"
)

const reference = "../testdata/r-reference/"

// meanThenMedian is the statistic list the command-line program uses: R computes both on
// each resample (spec §3).
var meanThenMedian = []Statistic{stats.Mean, stats.Median}

func relativeDifference(got, want float64) float64 {
	return math.Abs(got-want) / math.Abs(want)
}

// referenceLines returns the lines of a reference file after its provenance and header lines.
func referenceLines(t *testing.T, name string) []string {
	t.Helper()
	file, err := os.Open(reference + name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for line := 0; scanner.Scan(); line++ {
		if line >= 2 {
			lines = append(lines, scanner.Text())
		}
	}
	return lines
}

// TestReproducesRsResampleMediansOnRealData is the central claim (spec §6): on a committed
// data set with decimal values, every one of the 1,000 resample medians equals R's, as the
// same float64, and the standard errors agree to 1e-12.
func TestReproducesRsResampleMediansOnRealData(t *testing.T) {
	want, err := record.Parse(referenceLines(t, "lognormal-1000-record.tsv")[0])
	if err != nil {
		t.Fatal(err)
	}
	data, err := dataset.ReadFile("../data/" + want.Dataset)
	if err != nil {
		t.Fatal(err)
	}
	config := Config{Replicates: want.Replicates, Stream: RStream(want.Seed), Workers: 1}
	values, err := Replicates(data, meanThenMedian, config)
	if err != nil {
		t.Fatal(err)
	}

	wantMedians := referenceLines(t, "lognormal-1000-medians.tsv")
	if len(wantMedians) != want.Replicates {
		t.Fatalf("reference has %d medians, want %d", len(wantMedians), want.Replicates)
	}
	different := 0
	for replicate, text := range wantMedians {
		wantMedian, err := strconv.ParseFloat(text, 64)
		if err != nil {
			t.Fatal(err)
		}
		if values[1][replicate] != wantMedian {
			different++
		}
	}
	if different > 0 {
		t.Errorf("%d of %d resample medians differ from R's", different, want.Replicates)
	}

	if got := stats.Median(data); got != want.SampleMedian {
		t.Errorf("sample median: got %.17g, R gave %.17g", got, want.SampleMedian)
	}
	checks := []struct {
		name      string
		got, want float64
	}{
		{"sample mean", stats.Mean(data), want.SampleMean},
		{"SE of the median", stats.StandardDeviation(values[1]), want.SEMedian},
		{"SE of the mean", stats.StandardDeviation(values[0]), want.SEMean},
	}
	for _, check := range checks {
		if relativeDifference(check.got, check.want) > 1e-12 {
			t.Errorf("%s: got %.17g, R gave %.17g", check.name, check.got, check.want)
		}
	}
}

// TestReproducesRsStandardErrorsOnExactData uses R's first reference: x = (37 i) mod 101,
// n = 25 (odd) and n = 100 (even), B = 200.
func TestReproducesRsStandardErrorsOnExactData(t *testing.T) {
	for _, line := range referenceLines(t, "bootstrap-se.tsv") {
		fields := strings.Split(line, "\t")
		n, _ := strconv.Atoi(fields[0])
		replicates, _ := strconv.Atoi(fields[1])
		seed, _ := strconv.Atoi(fields[2])
		wantMedianSE, _ := strconv.ParseFloat(fields[3], 64)
		wantMeanSE, _ := strconv.ParseFloat(fields[4], 64)

		data := make([]float64, n)
		for i := range data {
			data[i] = float64((37 * (i + 1)) % 101)
		}
		config := Config{Replicates: replicates, Stream: RStream(uint32(seed)), Workers: 1}
		got, err := StandardErrors(data, meanThenMedian, config)
		if err != nil {
			t.Fatal(err)
		}
		if relativeDifference(got[1], wantMedianSE) > 1e-12 {
			t.Errorf("n = %d, SE of the median: got %.17g, R gave %.17g", n, got[1], wantMedianSE)
		}
		if relativeDifference(got[0], wantMeanSE) > 1e-12 {
			t.Errorf("n = %d, SE of the mean: got %.17g, R gave %.17g", n, got[0], wantMeanSE)
		}
	}
}

// TestAcceptsAStatisticFromOutsideThePackage shows reuse (FR-R1): a 10 % trimmed mean,
// defined here, is bootstrapped without any change to the engine. Its standard error must
// lie between plausible bounds: positive, and below the spread of the data.
func TestAcceptsAStatisticFromOutsideThePackage(t *testing.T) {
	trimmedMean := func(resample []float64) float64 {
		slices.Sort(resample)
		trim := len(resample) / 10
		return stats.Mean(resample[trim : len(resample)-trim])
	}
	data, err := dataset.ReadFile("../data/lognormal-1000.csv")
	if err != nil {
		t.Fatal(err)
	}
	config := Config{Replicates: 500, Stream: RStream(431), Workers: 1}
	got, err := StandardErrors(data, []Statistic{trimmedMean}, config)
	if err != nil {
		t.Fatal(err)
	}
	if !(got[0] > 0 && got[0] < stats.StandardDeviation(data)) {
		t.Errorf("standard error of the trimmed mean = %v, want between 0 and the data's spread", got[0])
	}
}

// TestEveryStatisticSeesTheSameResample checks spec §3: one set of resamples per run.
// A statistic that records what it receives is listed twice; within a replicate both copies
// must receive the same resample, and different replicates must receive different ones.
func TestEveryStatisticSeesTheSameResample(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var seen [][]float64
	capture := func(resample []float64) float64 {
		seen = append(seen, slices.Clone(resample))
		return 0
	}
	config := Config{Replicates: 5, Stream: RStream(431), Workers: 1}
	if _, err := Replicates(data, []Statistic{capture, capture}, config); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 10 {
		t.Fatalf("statistics were called %d times, want 10", len(seen))
	}
	for replicate := 0; replicate < 5; replicate++ {
		if !slices.Equal(seen[2*replicate], seen[2*replicate+1]) {
			t.Errorf("replicate %d: the two statistics received different resamples", replicate)
		}
	}
	if slices.Equal(seen[0], seen[2]) {
		t.Error("replicates 0 and 1 received the same resample")
	}
}

func TestRejectsUnusableRequests(t *testing.T) {
	data := []float64{1, 2, 3}
	one := []Statistic{stats.Mean}
	cases := []struct {
		name          string
		data          []float64
		statistics    []Statistic
		config        Config
		wantInMessage string
	}{
		{"one observation", data[:1], one, Config{Replicates: 10, Stream: RStream(1), Workers: 1}, "at least 2 are needed"},
		{"no statistic", data, nil, Config{Replicates: 10, Stream: RStream(1), Workers: 1}, "no statistic"},
		{"one replicate", data, one, Config{Replicates: 1, Stream: RStream(1), Workers: 1}, "1 replicate(s)"},
		{"no stream", data, one, Config{Replicates: 10, Workers: 1}, "no random stream"},
		{"no workers", data, one, Config{Replicates: 10, Stream: RStream(1), Workers: 0}, "0 worker(s)"},
		{"R stream with two workers", data, one, Config{Replicates: 10, Stream: RStream(1), Workers: 2}, "sequential and runs with 1 worker"},
	}
	for _, testCase := range cases {
		_, err := StandardErrors(testCase.data, testCase.statistics, testCase.config)
		if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), testCase.wantInMessage) {
			t.Errorf("%s: got error %v, want ErrInvalid mentioning %q", testCase.name, err, testCase.wantInMessage)
		}
	}
}
