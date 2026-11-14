package report

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func near(got, want float64) bool {
	return math.Abs(got-want) <= 1e-9*math.Max(1, math.Abs(want))
}

// writeFile puts content in a temporary file and returns its path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// The fixture copies the shape of a real results/runs.tsv (header, empty read_ms and
// compute_ms on start-up rows) with times chosen so every statistic can be checked by hand.
const runsFixture = `group	implementation	dataset	n	run	real_s	user_s	sys_s	max_rss_bytes	peak_footprint_bytes	read_ms	compute_ms
startup	r-empty	-	0	1	0.10	0.08	0.01	50000000	40000000		
startup	r-empty	-	0	2	0.20	0.08	0.01	50000000	60000000		
scaling	r-loop	lognormal-1000.csv	1000	1	1.00	0.9	0.1	100000000	80000000	2.000	500.000
scaling	go-r	lognormal-1000.csv	1000	1	0.10	0.1	0.0	5000000	3000000	0.500	50.000
scaling	r-loop	lognormal-1000.csv	1000	2	2.00	0.9	0.1	100000000	80000000	4.000	700.000
scaling	go-r	lognormal-1000.csv	1000	2	0.30	0.1	0.0	5000000	5000000	0.500	150.000
scaling	r-loop	lognormal-1000.csv	1000	3	3.00	0.9	0.1	100000000	80000000	3.000	600.000
scaling	r-loop	lognormal-1000.csv	1000	4	4.00	0.9	0.1	100000000	80000000	3.000	600.000
`

func TestSummarizeAgainstHandComputedAnswers(t *testing.T) {
	runs, err := ReadRuns(writeFile(t, "runs.tsv", runsFixture))
	if err != nil {
		t.Fatal(err)
	}
	summaries := Summarize(runs)
	if len(summaries) != 3 {
		t.Fatalf("got %d conditions, want 3", len(summaries))
	}
	// r-loop times 1, 2, 3, 4: mean 2.5, median 2.5, sd sqrt(5/3), error of the mean sd/2/2.5.
	loop, found := Find(summaries, "scaling", "r-loop", "lognormal-1000.csv")
	deviation := math.Sqrt(5.0 / 3.0)
	if !found || loop.Count != 4 || loop.Observations != 1000 || !near(loop.Mean, 2.5) || !near(loop.Median, 2.5) ||
		!near(loop.Deviation, deviation) || !near(loop.RelativeError, deviation/2/2.5) ||
		!near(loop.MeanComputeMs, 600) || !near(loop.MeanReadMs, 3) || !near(loop.MeanFootprint, 80e6) {
		t.Errorf("r-loop summary wrong: %+v", loop)
	}
	// go-r times 0.1 and 0.3: mean 0.2, median 0.2; footprints 3 MB and 5 MB: mean 4 MB, sd sqrt(2) MB.
	goR, _ := Find(summaries, "scaling", "go-r", "lognormal-1000.csv")
	if !near(goR.Mean, 0.2) || !near(goR.Median, 0.2) || !near(goR.MeanComputeMs, 100) ||
		!near(goR.MeanFootprint, 4e6) || !near(goR.Spread, math.Sqrt2*1e6) {
		t.Errorf("go-r summary wrong: %+v", goR)
	}
	// Start-up rows log no read or compute time: those means are absent, not zero.
	empty, _ := Find(summaries, "startup", "r-empty", "-")
	if !near(empty.Mean, 0.15) || !math.IsNaN(empty.MeanComputeMs) || !math.IsNaN(empty.MeanReadMs) {
		t.Errorf("start-up summary wrong: %+v", empty)
	}
	if _, found := Find(summaries, "scaling", "r-boot", "lognormal-1000.csv"); found {
		t.Error("found a condition that is not in the file")
	}
}

func TestReadRunsRejectsUnusableFiles(t *testing.T) {
	for name, content := range map[string]string{
		"missing column": "group\timplementation\n",
		"ragged row":     "group\timplementation\tdataset\tn\trun\treal_s\tuser_s\tsys_s\tmax_rss_bytes\tpeak_footprint_bytes\tread_ms\tcompute_ms\nscaling\tgo-r\n",
		"no time":        "group\timplementation\tdataset\tn\trun\treal_s\tuser_s\tsys_s\tmax_rss_bytes\tpeak_footprint_bytes\tread_ms\tcompute_ms\nscaling\tgo-r\td\t1\t1\t\t\t\t1\t1\t\t\n",
		"empty":          "",
	} {
		if _, err := ReadRuns(writeFile(t, "runs.tsv", content)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := ReadRuns(filepath.Join(t.TempDir(), "absent.tsv")); err == nil {
		t.Error("absent file accepted")
	}
}

// The benchmark fixture has the shape of real "go test -bench -benchmem" output: header
// lines, a CPU-count suffix, sub-benchmark names with "/" and "=", repetitions, PASS/ok lines.
const benchmarkFixture = `goos: darwin
goarch: arm64
pkg: github.com/mioluigi/go-applied-statistics/stats
cpu: Apple M1 Max
BenchmarkStep1Median/before=sort-10         	    3000	    400000 ns/op	   81920 B/op	       1 allocs/op
BenchmarkStep1Median/before=sort-10         	    3000	    420000 ns/op	   81920 B/op	       1 allocs/op
BenchmarkStep1Median/before=sort-10         	    3000	    380000 ns/op	   81920 B/op	       1 allocs/op
BenchmarkStep1Median/after=select-10        	   30000	     40000 ns/op	       0 B/op	       0 allocs/op
BenchmarkFloat64-10                         	300000000	         4.5 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/mioluigi/go-applied-statistics/stats	12.345s
`

func TestReadBenchmarksTakesTheMedianOfRepetitions(t *testing.T) {
	benchmarks, err := ReadBenchmarks(writeFile(t, "benchmarks.txt", benchmarkFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(benchmarks) != 3 {
		t.Fatalf("got %d benchmarks, want 3", len(benchmarks))
	}
	before, found := FindBenchmark(benchmarks, "Step1Median/before=sort")
	if !found || before.Count != 3 || before.NanosPerOp != 400000 || before.BytesPerOp != 81920 || before.AllocsPerOp != 1 {
		t.Errorf("before: %+v", before)
	}
	if float64Benchmark, _ := FindBenchmark(benchmarks, "Float64"); float64Benchmark.NanosPerOp != 4.5 {
		t.Errorf("decimal ns/op not read: %+v", float64Benchmark)
	}
	if _, err := ReadBenchmarks(writeFile(t, "none.txt", "PASS\n")); err == nil {
		t.Error("a file without results was accepted")
	}
}

func TestCostModelAgainstHandComputedAnswers(t *testing.T) {
	small := Price{Instance: "c7g.medium", VCPUs: 1, USDPerHour: 0.0363}
	large := Price{Instance: "c7g.2xlarge", VCPUs: 8, USDPerHour: 0.29}
	// A 3.6-second job: one million jobs take 1,000 instance-hours one at a time.
	if got := CostPerJobs(3.6, small, 1); !near(got, 36.30) {
		t.Errorf("1 vCPU: got $%v, want $36.30", got)
	}
	// Eight at once on the 8-vCPU instance: 125 instance-hours at $0.29.
	if got := CostPerJobs(3.6, large, 8); !near(got, 36.25) {
		t.Errorf("8 vCPUs, 8 jobs at once: got $%v, want $36.25", got)
	}
	// One at a time on the 8-vCPU instance: the idle cores are paid for.
	if got := CostPerJobs(3.6, large, 1); !near(got, 290) {
		t.Errorf("8 vCPUs, one job at a time: got $%v, want $290", got)
	}
	if got := Saving(36.30, 3.63); !near(got, 0.9) {
		t.Errorf("saving: got %v, want 0.9", got)
	}
	// $32.67 saved per million jobs: 1e6 / 32.67 jobs recover one dollar.
	if got := JobsPerDollarSaved(36.30, 3.63); !near(got, 1e6/32.67) {
		t.Errorf("jobs per dollar: got %v, want %v", got, 1e6/32.67)
	}
}

func TestReadPrices(t *testing.T) {
	prices, err := ReadPrices("../../prices/aws-ec2.tsv")
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 2 || prices[0].Instance != "c7g.medium" || prices[0].VCPUs != 1 || prices[1].VCPUs != 8 ||
		!(prices[0].USDPerHour > 0) || prices[0].Source == "" || prices[0].Captured == "" {
		t.Errorf("unexpected prices: %+v", prices)
	}
	header := "provider\tinstance\tarchitecture\tvcpus\tmemory_gib\tregion\toperating_system\tpurchase\tusd_per_hour\tsource\tcaptured\n"
	for name, row := range map[string]string{
		"no source": "AWS\tx\tarm64\t1\t2\tr\tLinux\ton-demand\t0.1\t\t2026-11-14\n",
		"no date":   "AWS\tx\tarm64\t1\t2\tr\tLinux\ton-demand\t0.1\thttps://example.org\t\n",
		"no price":  "AWS\tx\tarm64\t1\t2\tr\tLinux\ton-demand\t\thttps://example.org\t2026-11-14\n",
		"no vcpus":  "AWS\tx\tarm64\t0\t2\tr\tLinux\ton-demand\t0.1\thttps://example.org\t2026-11-14\n",
	} {
		if _, err := ReadPrices(writeFile(t, "prices.tsv", header+row)); err == nil {
			t.Errorf("%s: a price without provenance was accepted", name)
		}
	}
}
