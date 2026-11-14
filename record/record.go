// Package record defines the one-line, tab-separated result that the Go program and the R
// scripts both write, so that a program can compare them field by field.
package record

import (
	"fmt"
	"strconv"
	"strings"
)

// Header names the fields of a record, in order, separated by tabs.
const Header = "dataset\tn\treplicates\tseed\tstream\tsample_median\tse_median\tsample_mean\tse_mean"

// fieldCount is the number of tab-separated fields in a record.
const fieldCount = 9

// Record is the result of one bootstrap run on one data set.
type Record struct {
	Dataset      string  // file name of the data set
	Observations int     // n
	Replicates   int     // B
	Seed         uint32  // seed of the random stream
	Stream       string  // which random stream produced the resamples: "r", "block" or "boot"
	SampleMedian float64 // median of the data
	SEMedian     float64 // bootstrap standard error of the median
	SampleMean   float64 // mean of the data
	SEMean       float64 // bootstrap standard error of the mean
}

// Exact formats a number with 17 significant digits, enough to recover the same float64.
// It matches sprintf("%.17g", x) in the R scripts.
func Exact(value float64) string {
	return strconv.FormatFloat(value, 'g', 17, 64)
}

// String returns the record as one tab-separated line, without a line ending.
func (r Record) String() string {
	return strings.Join([]string{
		r.Dataset,
		strconv.Itoa(r.Observations),
		strconv.Itoa(r.Replicates),
		strconv.FormatUint(uint64(r.Seed), 10),
		r.Stream,
		Exact(r.SampleMedian),
		Exact(r.SEMedian),
		Exact(r.SampleMean),
		Exact(r.SEMean),
	}, "\t")
}

// Parse reads one record line, as written by String or by the R scripts.
func Parse(line string) (Record, error) {
	fields := strings.Split(strings.TrimRight(line, "\r\n"), "\t")
	if len(fields) != fieldCount {
		return Record{}, fmt.Errorf("record has %d fields, want %d: %q", len(fields), fieldCount, line)
	}
	var parsed Record
	var err error
	fail := func(field string, cause error) (Record, error) {
		return Record{}, fmt.Errorf("record field %s: %w", field, cause)
	}
	parsed.Dataset, parsed.Stream = fields[0], fields[4]
	if parsed.Observations, err = strconv.Atoi(fields[1]); err != nil {
		return fail("n", err)
	}
	if parsed.Replicates, err = strconv.Atoi(fields[2]); err != nil {
		return fail("replicates", err)
	}
	seed, err := strconv.ParseUint(fields[3], 10, 32)
	if err != nil {
		return fail("seed", err)
	}
	parsed.Seed = uint32(seed)
	numbers := []*float64{&parsed.SampleMedian, &parsed.SEMedian, &parsed.SampleMean, &parsed.SEMean}
	for i, target := range numbers {
		if *target, err = strconv.ParseFloat(fields[5+i], 64); err != nil {
			return fail(strings.Split(Header, "\t")[5+i], err)
		}
	}
	return parsed, nil
}
