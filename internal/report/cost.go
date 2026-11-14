package report

import (
	"fmt"
	"strconv"
)

// A Price is one published instance price, from prices/aws-ec2.tsv.
type Price struct {
	Provider   string
	Instance   string
	VCPUs      int
	MemoryGiB  float64
	Region     string
	USDPerHour float64
	Source     string
	Captured   string
}

// ReadPrices reads the saved price file.
func ReadPrices(path string) ([]Price, error) {
	rows, column, err := readTable(path)
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"provider", "instance", "vcpus", "memory_gib", "region", "usd_per_hour", "source", "captured"} {
		if _, found := column[name]; !found {
			return nil, fmt.Errorf("%s lacks the column %q", path, name)
		}
	}
	prices := make([]Price, 0, len(rows))
	for _, row := range rows {
		vcpus, err := strconv.Atoi(row[column["vcpus"]])
		if err != nil || vcpus < 1 {
			return nil, fmt.Errorf("%s: instance %s has no usable vCPU count", path, row[column["instance"]])
		}
		usd := number(row[column["usd_per_hour"]])
		if !(usd > 0) || row[column["source"]] == "" || row[column["captured"]] == "" {
			return nil, fmt.Errorf("%s: instance %s needs a positive price, a source and a capture date", path, row[column["instance"]])
		}
		prices = append(prices, Price{
			Provider: row[column["provider"]], Instance: row[column["instance"]], VCPUs: vcpus,
			MemoryGiB: number(row[column["memory_gib"]]), Region: row[column["region"]],
			USDPerHour: usd, Source: row[column["source"]], Captured: row[column["captured"]],
		})
	}
	return prices, nil
}

const (
	secondsPerHour = 3600
	// JobsPerQuote is the batch size costs are quoted for: one million jobs.
	JobsPerQuote = 1e6
)

// CostPerJobs returns the cost in USD of JobsPerQuote jobs that each take the given seconds
// on an instance, when the instance runs concurrent jobs at once. One job at a time on a
// 1-vCPU instance is concurrent = 1; a single-threaded program filling an 8-vCPU instance
// with 8 independent jobs is concurrent = 8.
func CostPerJobs(seconds float64, price Price, concurrent int) float64 {
	instanceHours := seconds * JobsPerQuote / secondsPerHour / float64(concurrent)
	return instanceHours * price.USDPerHour
}

// Saving returns the share of cost saved by moving from the first cost to the second.
func Saving(costBefore, costAfter float64) float64 {
	return 1 - costAfter/costBefore
}

// JobsPerDollarSaved returns how many jobs must run before the saving adds up to one dollar:
// the number of jobs needed to recover each dollar spent on rewriting.
func JobsPerDollarSaved(costBefore, costAfter float64) float64 {
	return JobsPerQuote / (costBefore - costAfter)
}
