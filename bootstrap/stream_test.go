package bootstrap

import (
	"runtime"
	"slices"
	"testing"

	"github.com/mioluigi/go-applied-statistics/dataset"
	"github.com/mioluigi/go-applied-statistics/stats"
)

func readData(t *testing.T, name string) []float64 {
	t.Helper()
	data, err := dataset.ReadFile("../data/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestBlockStreamGivesTheSameResultForAnyWorkerCount is FR-6 and AC-5: bit-for-bit equal
// replicate values for 1, 2, 4 and 8 workers, and for a count that does not divide B.
func TestBlockStreamGivesTheSameResultForAnyWorkerCount(t *testing.T) {
	data := readData(t, "lognormal-1000.csv")
	run := func(workers int) [][]float64 {
		config := Config{Replicates: 1000, Stream: BlockStream(431), Workers: workers}
		values, err := Replicates(data, meanThenMedian, config)
		if err != nil {
			t.Fatal(err)
		}
		return values
	}
	oneWorker := run(1)
	for _, workers := range []int{2, 4, 8, 7} {
		got := run(workers)
		for s := range oneWorker {
			if !slices.Equal(got[s], oneWorker[s]) {
				t.Errorf("%d workers: statistic %d differs from the 1-worker result", workers, s)
			}
		}
	}
}

// TestBlockStreamAgreesWithRStream is risk R5 and spec §6: different draws, same answer up
// to Monte Carlo error (15 % at B = 1,000, spike 002), on every shape. The two larger sizes are left to
// the official comparison on the Mac, which runs all twelve data sets.
func TestBlockStreamAgreesWithRStream(t *testing.T) {
	for _, shape := range []string{"normal", "lognormal", "mirrored-lognormal"} {
		for _, size := range []string{"100", "1000"} {
			name := shape + "-" + size + ".csv"
			data := readData(t, name)
			fromR, err := StandardErrors(data, meanThenMedian, Config{Replicates: 1000, Stream: RStream(431), Workers: 1})
			if err != nil {
				t.Fatal(err)
			}
			fromBlocks, err := StandardErrors(data, meanThenMedian, Config{Replicates: 1000, Stream: BlockStream(431), Workers: 4})
			if err != nil {
				t.Fatal(err)
			}
			for s, statistic := range []string{"mean", "median"} {
				if difference := relativeDifference(fromBlocks[s], fromR[s]); difference > 0.15 {
					t.Errorf("%s, SE of the %s: block stream %v, R stream %v (%.1f %% apart)",
						name, statistic, fromBlocks[s], fromR[s], 100*difference)
				}
			}
		}
	}
}

// TestBlockStreamReplicatesAndSeedsDiffer guards against a source that ignores the
// replicate number or the seed.
func TestBlockStreamReplicatesAndSeedsDiffer(t *testing.T) {
	draw := func(stream Stream, replicate int) []int {
		source := stream.Source(replicate)
		indices := make([]int, 20)
		for i := range indices {
			indices[i] = source.Index(1000)
		}
		return indices
	}
	if slices.Equal(draw(BlockStream(431), 0), draw(BlockStream(431), 1)) {
		t.Error("replicates 0 and 1 draw the same indices")
	}
	if slices.Equal(draw(BlockStream(431), 0), draw(BlockStream(432), 0)) {
		t.Error("seeds 431 and 432 draw the same indices")
	}
	if !slices.Equal(draw(BlockStream(431), 5), draw(BlockStream(431), 5)) {
		t.Error("the same seed and replicate drew different indices")
	}
}

// countingStream is a random source written outside the package (FR-R2): a tiny linear
// congruential generator per replicate. Its quality does not matter here; that the engine
// runs it, in parallel, is the point.
type countingStream struct{}

type congruential struct{ state uint64 }

func (source *congruential) Index(n int) int {
	source.state = source.state*6364136223846793005 + 1442695040888963407
	return int((source.state >> 33) % uint64(n))
}

func (countingStream) Source(replicate int) IndexSource {
	return &congruential{state: uint64(replicate) + 1}
}
func (countingStream) Sequential() bool { return false }

func TestAcceptsARandomSourceFromOutsideThePackage(t *testing.T) {
	data := readData(t, "normal-1000.csv")
	config := Config{Replicates: 400, Stream: countingStream{}, Workers: 3}
	got, err := StandardErrors(data, []Statistic{stats.Mean}, config)
	if err != nil {
		t.Fatal(err)
	}
	expected := stats.StandardDeviation(data) / 31.622776601683793 // s / sqrt(1000)
	if relativeDifference(got[0], expected) > 0.15 {
		t.Errorf("SE of the mean with the outside source = %v, want about %v", got[0], expected)
	}
}

// TestNoGoroutineOutlivesARun: the count of goroutines after many parallel runs equals the
// count before. The test releases nothing itself (SFEDD §5.5).
func TestNoGoroutineOutlivesARun(t *testing.T) {
	data := readData(t, "normal-100.csv")
	before := runtime.NumGoroutine()
	for run := 0; run < 50; run++ {
		config := Config{Replicates: 64, Stream: BlockStream(uint32(run)), Workers: 8}
		if _, err := Replicates(data, meanThenMedian, config); err != nil {
			t.Fatal(err)
		}
	}
	if after := runtime.NumGoroutine(); after != before {
		t.Errorf("goroutines: %d before, %d after", before, after)
	}
}
