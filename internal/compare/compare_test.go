package compare

import (
	"math"
	"strings"
	"testing"

	"github.com/mioluigi/go-applied-statistics/record"
)

var base = record.Record{
	Dataset: "d.csv", Observations: 1000, Replicates: 1000, Seed: 431, Stream: "r",
	SampleMedian: 86.6483, SEMedian: 1.5, SampleMean: 97.88, SEMean: 1.7,
}

// failed lists the quantities that did not pass.
func failed(checks []Check) string {
	var names []string
	for _, check := range checks {
		if !check.Pass {
			names = append(names, check.Quantity)
		}
	}
	return strings.Join(names, ", ")
}

func TestSameDraws(t *testing.T) {
	lastBits := base
	lastBits.SEMedian = math.Nextafter(base.SEMedian, 2)
	lastBits.SEMean = base.SEMean * (1 + 5e-13)
	lastBits.SampleMean = math.Nextafter(base.SampleMean, 0)
	if checks, _ := Records(base, lastBits, SameDraws); failed(checks) != "" {
		t.Errorf("last-bit differences in summed quantities should pass; failed: %s", failed(checks))
	}

	cases := []struct {
		name       string
		change     func(*record.Record)
		wantFailed string
	}{
		{"median off by one bit", func(r *record.Record) { r.SampleMedian = math.Nextafter(r.SampleMedian, 100) }, "sample median"},
		{"n differs", func(r *record.Record) { r.Observations = 999 }, "n"},
		{"SE of the median off by 1e-11", func(r *record.Record) { r.SEMedian *= 1 + 1e-11 }, "SE of the median"},
		{"SE of the mean off by 1 %", func(r *record.Record) { r.SEMean *= 1.01 }, "SE of the mean"},
		{"sample mean off by 1e-11", func(r *record.Record) { r.SampleMean *= 1 + 1e-11 }, "sample mean"},
	}
	for _, testCase := range cases {
		candidate := base
		testCase.change(&candidate)
		checks, err := Records(base, candidate, SameDraws)
		if err != nil || failed(checks) != testCase.wantFailed {
			t.Errorf("%s: failed %q (error %v), want %q", testCase.name, failed(checks), err, testCase.wantFailed)
		}
	}
}

func TestDifferentDraws(t *testing.T) {
	near := base
	near.SEMedian, near.SEMean = base.SEMedian*1.09, base.SEMean*0.92
	if checks, _ := Records(base, near, DifferentDraws); failed(checks) != "" {
		t.Errorf("9 %% and 8 %% apart should pass; failed: %s", failed(checks))
	}
	far := base
	far.SEMedian = base.SEMedian * 1.11
	if checks, _ := Records(base, far, DifferentDraws); failed(checks) != "SE of the median" {
		t.Errorf("11 %% apart: failed %q, want the SE of the median", failed(checks))
	}
	// The data are the same, so the sample median must still be identical.
	shifted := base
	shifted.SampleMedian += 1e-9
	if checks, _ := Records(base, shifted, DifferentDraws); failed(checks) != "sample median" {
		t.Errorf("shifted median: failed %q", failed(checks))
	}
}

func TestSameProgramRequiresIdentity(t *testing.T) {
	if checks, _ := Records(base, base, SameProgram); failed(checks) != "" {
		t.Errorf("identical records failed: %s", failed(checks))
	}
	lastBit := base
	lastBit.SEMean = math.Nextafter(base.SEMean, 2)
	if checks, _ := Records(base, lastBit, SameProgram); failed(checks) != "SE of the mean" {
		t.Errorf("a one-bit difference between worker counts must fail; failed %q", failed(checks))
	}
}

func TestRecordsOfDifferentRunsAreNotCompared(t *testing.T) {
	for _, change := range []func(*record.Record){
		func(r *record.Record) { r.Dataset = "other.csv" },
		func(r *record.Record) { r.Replicates = 500 },
		func(r *record.Record) { r.Seed = 1 },
	} {
		candidate := base
		change(&candidate)
		if _, err := Records(base, candidate, SameDraws); err == nil {
			t.Errorf("compared records of different runs: %+v", candidate)
		}
	}
}

func TestResampleMedians(t *testing.T) {
	reference := []float64{1, 2, 3, 4}
	if check := ResampleMedians("d.csv", reference, []float64{1, 2, 3, 4}); !check.Pass {
		t.Error("identical lists failed")
	}
	if check := ResampleMedians("d.csv", reference, []float64{1, 2, 3.0000000001, 4}); check.Pass || check.Difference != 1 {
		t.Errorf("one differing value: pass=%v, difference=%v", check.Pass, check.Difference)
	}
	if check := ResampleMedians("d.csv", reference, []float64{1, 2}); check.Pass || check.Difference != 2 {
		t.Errorf("short list: pass=%v, difference=%v", check.Pass, check.Difference)
	}
	if check := ResampleMedians("d.csv", nil, nil); check.Pass {
		t.Error("two empty lists must not pass: nothing was compared")
	}
}

func TestMeanFormula(t *testing.T) {
	// Data 1..10: s = sqrt(55/6); theory = sqrt(9/10) * s / sqrt(10) = sqrt(0.825).
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	theory := math.Sqrt(0.825)
	result := base
	result.SEMean = theory * 1.05
	check := MeanFormula(result, data)
	if !check.Pass || math.Abs(check.Reference-theory) > 1e-15 {
		t.Errorf("5 %% from theory: pass=%v, theory computed as %v, want %v", check.Pass, check.Reference, theory)
	}
	result.SEMean = theory * 1.2
	if MeanFormula(result, data).Pass {
		t.Error("20 % from theory passed")
	}
}

func TestCheckLineFormat(t *testing.T) {
	line := quantity("d.csv", "sample median", 2, 2, Identical).String()
	if line != "d.csv\tsample median\t2\t2\t0\t0\tPASS" {
		t.Errorf("unexpected line %q", line)
	}
}
