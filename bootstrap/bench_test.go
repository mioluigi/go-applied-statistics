package bootstrap

import (
	"runtime"
	"slices"
	"strconv"
	"testing"

	"github.com/mioluigi/go-applied-statistics/rrand"
	"github.com/mioluigi/go-applied-statistics/stats"
)

// replicatesBefore is the engine as first written (task T07), kept only to measure and to
// check the performance step "reuse one buffer per worker": it allocates a new resample for
// every replicate.
func replicatesBefore(data []float64, statistics []Statistic, config Config) [][]float64 {
	values := make([][]float64, len(statistics))
	for s := range values {
		values[s] = make([]float64, config.Replicates)
	}
	for replicate := 0; replicate < config.Replicates; replicate++ {
		source := config.Stream.Source(replicate)
		resample := make([]float64, len(data))
		for i := range resample {
			resample[i] = data[source.Index(len(data))]
		}
		for s, statistic := range statistics {
			values[s][replicate] = statistic(resample)
		}
	}
	return values
}

// uniformData returns n reproducible values in (0, 1).
func uniformData(n int) []float64 {
	generator := rrand.New(431)
	data := make([]float64, n)
	for i := range data {
		data[i] = generator.Float64()
	}
	return data
}

// TestFasterEngineEqualsTheFirstVersion: buffer reuse and the selection median must not
// change a single replicate value (constitution 1.2).
func TestFasterEngineEqualsTheFirstVersion(t *testing.T) {
	for _, n := range []int{2, 3, 100, 101, 1000} {
		data := uniformData(n)
		before := replicatesBefore(data, []Statistic{stats.Mean, stats.Median},
			Config{Replicates: 300, Stream: RStream(431), Workers: 1})
		after, err := Replicates(data, []Statistic{stats.Mean, stats.MedianInPlace},
			Config{Replicates: 300, Stream: RStream(431), Workers: 1})
		if err != nil {
			t.Fatal(err)
		}
		for s, name := range []string{"mean", "median"} {
			if !slices.Equal(before[s], after[s]) {
				t.Errorf("n = %d: the %s replicates differ between the first and the faster engine", n, name)
			}
		}
	}
}

// Performance step "reuse one buffer per worker" (spec FR-12), same statistics on both sides.
func BenchmarkStep2Resample(b *testing.B) {
	data := uniformData(10000)
	statistics := []Statistic{stats.Mean, stats.MedianInPlace}
	b.Run("before=allocate", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			replicatesBefore(data, statistics, Config{Replicates: 100, Stream: RStream(431), Workers: 1})
		}
	})
	b.Run("after=reuse", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			Replicates(data, statistics, Config{Replicates: 100, Stream: RStream(431), Workers: 1})
		}
	})
}

// The whole engine, first version against final version, at a small and a large size:
// B = 100 resamples, mean and median.
func BenchmarkEngine(b *testing.B) {
	for _, n := range []int{1000, 100000} {
		data := uniformData(n)
		b.Run("before/n="+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				replicatesBefore(data, []Statistic{stats.Mean, stats.Median},
					Config{Replicates: 100, Stream: RStream(431), Workers: 1})
			}
		})
		b.Run("after/n="+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				Replicates(data, []Statistic{stats.Mean, stats.MedianInPlace},
					Config{Replicates: 100, Stream: RStream(431), Workers: 1})
			}
		})
	}
}

// The block stream on one worker and on all CPUs: what parallel resampling adds.
func BenchmarkWorkers(b *testing.B) {
	data := uniformData(10000)
	statistics := []Statistic{stats.Mean, stats.MedianInPlace}
	for _, workers := range []int{1, runtime.NumCPU()} {
		b.Run("workers="+strconv.Itoa(workers), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				Replicates(data, statistics, Config{Replicates: 1000, Stream: BlockStream(431), Workers: workers})
			}
		})
	}
}

// The resampling volume of the course's R program (spike prediction P4): 100 samples at each
// of four sizes, B = 100 resamples each, mean and median. Same volume, not the same program:
// the data here are uniform numbers, and the course program draws normal samples.
func BenchmarkCourseStudyVolume(b *testing.B) {
	samples := map[int][]float64{}
	for _, n := range []int{25, 100, 225, 400} {
		samples[n] = uniformData(n)
	}
	statistics := []Statistic{stats.Mean, stats.MedianInPlace}
	b.ReportAllocs()
	for b.Loop() {
		stream := RStream(9999)
		for iteration := 0; iteration < 100; iteration++ {
			for _, n := range []int{25, 100, 225, 400} {
				Replicates(samples[n], statistics, Config{Replicates: 100, Stream: stream, Workers: 1})
			}
		}
	}
}
