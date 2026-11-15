package report

import (
	"fmt"
	"math"
	"strings"
)

// costRow holds the cost figures for one data size.
type costRow struct {
	n                              int
	rSmall, goSmall                float64 // USD per million jobs, 1-vCPU instance, one job at a time
	goBlockSmall                   float64 // the same for Go's block stream on one worker
	savingMean, savingMedian       float64 // right-sized saving, from mean and from median times
	jobsPerDollar                  float64
	rLarge, goOneCoreLarge, goMany float64 // USD per million jobs on the multi-vCPU instance
	goBlockLarge                   float64 // block stream, one worker per job, one job per vCPU
}

// costs computes the cost figures for every size.
func (inputs *Inputs) costs() ([]costRow, Price, Price, error) {
	small, err := inputs.price(1)
	if err != nil {
		return nil, Price{}, Price{}, err
	}
	var large Price
	for _, price := range inputs.Prices {
		if price.VCPUs > large.VCPUs {
			large = price
		}
	}
	if large.VCPUs < 2 {
		return nil, Price{}, Price{}, fmt.Errorf("prices/aws-ec2.tsv has no multi-vCPU instance")
	}
	var rows []costRow
	for _, n := range Sizes {
		loop, err := inputs.scaling("r-loop", n)
		if err != nil {
			return nil, Price{}, Price{}, err
		}
		goR, err := inputs.scaling("go-r", n)
		if err != nil {
			return nil, Price{}, Price{}, err
		}
		goBlock, err := inputs.scaling("go-block-1", n)
		if err != nil {
			return nil, Price{}, Price{}, err
		}
		goMulti, err := inputs.scaling("go-block-multi", n)
		if err != nil {
			return nil, Price{}, Price{}, err
		}
		row := costRow{n: n,
			rSmall: CostPerJobs(loop.Mean, small, 1), goSmall: CostPerJobs(goR.Mean, small, 1),
			goBlockSmall: CostPerJobs(goBlock.Mean, small, 1),
			// A single-threaded program fills the large instance with one job per vCPU.
			rLarge: CostPerJobs(loop.Mean, large, large.VCPUs), goOneCoreLarge: CostPerJobs(goR.Mean, large, large.VCPUs),
			goBlockLarge: CostPerJobs(goBlock.Mean, large, large.VCPUs),
			// The parallel Go program uses every vCPU for one job at a time.
			goMany: CostPerJobs(goMulti.Mean, large, 1),
		}
		row.savingMean = Saving(row.rSmall, row.goSmall)
		row.savingMedian = Saving(CostPerJobs(loop.Median, small, 1), CostPerJobs(goR.Median, small, 1))
		row.jobsPerDollar = JobsPerDollarSaved(row.rSmall, row.goSmall)
		rows = append(rows, row)
	}
	return rows, small, large, nil
}

// costTables shows the saved prices and the cost of a million jobs under two scenarios.
func (inputs *Inputs) costTables() (string, error) {
	rows, small, large, err := inputs.costs()
	if err != nil {
		return "", err
	}
	var priceRows [][]string
	for _, price := range []Price{small, large} {
		priceRows = append(priceRows, []string{price.Provider + " `" + price.Instance + "`", fmt.Sprint(price.VCPUs),
			fmt.Sprintf("%g", price.MemoryGiB), price.Region, fmt.Sprintf("$%.4f", price.USDPerHour),
			fmt.Sprintf("$%.6f", price.USDPerHour/float64(price.VCPUs)), price.Source, price.Captured})
	}
	var smallRows, largeRows [][]string
	for _, row := range rows {
		smallRows = append(smallRows, []string{thousands(row.n), dollars(row.rSmall), dollars(row.goSmall),
			"**" + percent(row.savingMean) + "**", percent(row.savingMedian), jobsText(row.jobsPerDollar),
			dollars(row.goBlockSmall), percent(Saving(row.rSmall, row.goBlockSmall))})
		largeRows = append(largeRows, []string{thousands(row.n), dollars(row.rLarge),
			dollars(row.goOneCoreLarge), percent(Saving(row.rLarge, row.goOneCoreLarge)),
			dollars(row.goBlockLarge), percent(Saving(row.rLarge, row.goBlockLarge)),
			dollars(row.goMany), percent(Saving(row.rLarge, row.goMany))})
	}
	workers := inputs.machineValue("parallel workers")
	var out strings.Builder
	out.WriteString("**Published prices used** (on-demand, Linux)\n\n")
	out.WriteString(table([]string{"Instance", "vCPUs", "Memory (GiB)", "Region", "USD per hour", "USD per vCPU-hour", "Source", "Captured"}, priceRows))
	out.WriteString(fmt.Sprintf("\n**Scenario A, right-sized: one job at a time on `%s` (1 vCPU).** Cost of one million jobs (B = 1,000), from mean whole-process times.\n\n", small.Instance))
	out.WriteString(table([]string{"n", "R, `sample()` loop", "Go, R stream", "Saving (from means)", "Saving (from medians)", "Jobs to save one dollar",
		"Go, block stream, 1 worker", "Saving, block stream"}, smallRows))
	out.WriteString(fmt.Sprintf("\n**Scenario B, one `%s` (%d vCPUs) kept busy.** One-core programs run %d jobs at once; parallel Go runs one job at a time on %s workers. Savings are against R.\n\n",
		large.Instance, large.VCPUs, large.VCPUs, workers))
	atOnce := fmt.Sprint(large.VCPUs) + " jobs at once"
	out.WriteString(table([]string{"n", "R, " + atOnce, "Go, R stream, " + atOnce, "Saving",
		"Go, block stream, 1 worker each, " + atOnce, "Saving", "Go, block stream, " + workers + " workers, 1 job at a time", "Saving"}, largeRows))
	return out.String(), nil
}

// jobsText formats the number of jobs needed to save one dollar. When Go is not cheaper
// there is no such number, and the cell says so.
func jobsText(jobsPerDollar float64) string {
	if math.IsInf(jobsPerDollar, 0) || math.IsNaN(jobsPerDollar) || jobsPerDollar <= 0 {
		return "no saving"
	}
	return thousands(int(math.Round(jobsPerDollar)))
}

// verdict words a prediction's outcome.
func verdict(held bool) string {
	if held {
		return "**Held**"
	}
	return "**Did not hold**"
}

// predictions checks the spike's predictions P4 to P7 against the measurements.
func (inputs *Inputs) predictions() (string, error) {
	volume, err := inputs.benchmark("CourseStudyVolume")
	if err != nil {
		return "", err
	}
	empty, err := inputs.startup("r-empty")
	if err != nil {
		return "", err
	}
	computeRatio, startupShare := map[int]float64{}, map[int]float64{}
	for _, n := range Sizes {
		loop, err := inputs.scaling("r-loop", n)
		if err != nil {
			return "", err
		}
		goR, err := inputs.scaling("go-r", n)
		if err != nil {
			return "", err
		}
		computeRatio[n], startupShare[n] = loop.MeanComputeMs/goR.MeanComputeMs, empty.Mean/loop.Mean
	}
	differentDraws, allPassed, largest := 0, true, 0.0
	for _, check := range inputs.Checks {
		if strings.Contains(check.Comparison, "different draws") && strings.HasPrefix(check.Quantity, "SE of") {
			differentDraws++
			allPassed = allPassed && check.Pass
			largest = math.Max(largest, check.Difference)
		}
	}
	volumeSeconds := volume.NanosPerOp / 1e9
	inRange := func(ratio float64) bool { return ratio >= 2 && ratio <= 10 }
	rows := [][]string{
		{"P4", "Go does the course study's resampling volume in under 0.20 s of computation",
			seconds(volumeSeconds) + " s", verdict(volumeSeconds < 0.20)},
		{"P5", "Computation only, R ÷ Go (one core) is between 2× and 10× at n = 1,000 and n = 10,000",
			times(computeRatio[1000]) + " and " + times(computeRatio[10000]),
			verdict(inRange(computeRatio[1000]) && inRange(computeRatio[10000]))},
		{"P6", "R's start-up is more than 50 % of its whole-process time at n = 100 and under 20 % at n = 100,000",
			percent(startupShare[100]) + " and " + percent(startupShare[100000]),
			verdict(startupShare[100] > 0.5 && startupShare[100000] < 0.2)},
		{"P7", "Every different-draws check passes at 15 %, and the largest difference is under 11 %",
			fmt.Sprintf("%d checks, largest %s", differentDraws, percent(largest)),
			verdict(differentDraws > 0 && allPassed && largest < 0.11)},
	}
	return table([]string{"#", "Prediction (with its conditions)", "Measured", "Outcome"}, rows), nil
}

// headline states the few numbers the recommendation rests on.
func (inputs *Inputs) headline() (string, error) {
	costRows, small, _, err := inputs.costs()
	if err != nil {
		return "", err
	}
	passed, total := inputs.checksPassed()
	identical, compared := inputs.medianIdentity()
	smallest, largest := Sizes[0], Sizes[len(Sizes)-1]
	summaries := map[string]Summary{}
	for _, key := range []struct {
		implementation string
		n              int
	}{{"r-loop", smallest}, {"go-r", smallest}, {"r-loop", largest}, {"go-r", largest}, {"go-block-1", largest}, {"go-block-multi", largest}} {
		summary, err := inputs.scaling(key.implementation, key.n)
		if err != nil {
			return "", err
		}
		summaries[fmt.Sprintf("%s/%d", key.implementation, key.n)] = summary
	}
	empty, err := inputs.startup("r-empty")
	if err != nil {
		return "", err
	}
	loopSmall, goSmall := summaries[fmt.Sprintf("r-loop/%d", smallest)], summaries[fmt.Sprintf("go-r/%d", smallest)]
	loopLarge, goLarge := summaries[fmt.Sprintf("r-loop/%d", largest)], summaries[fmt.Sprintf("go-r/%d", largest)]
	oneWorker, manyWorkers := summaries[fmt.Sprintf("go-block-1/%d", largest)], summaries[fmt.Sprintf("go-block-multi/%d", largest)]
	lowSaving, highSaving := 1.0, 0.0
	for _, row := range costRows {
		lowSaving, highSaving = math.Min(lowSaving, row.savingMean), math.Max(highSaving, row.savingMean)
	}
	lines := []string{
		fmt.Sprintf("**Same results.** %d of %d comparison checks passed; %s of %s resample medians are identical between R and Go.",
			passed, total, thousands(identical), thousands(compared)),
		fmt.Sprintf("**Computation only, one core** (n = %s, B = 1,000): R %s ms, Go %s ms: Go is **%s** faster.",
			thousands(largest), milliseconds(loopLarge.MeanComputeMs), milliseconds(goLarge.MeanComputeMs),
			times(loopLarge.MeanComputeMs/goLarge.MeanComputeMs)),
		fmt.Sprintf("**Whole process, small job** (n = %s): R %s s, of which %s is R starting up; Go about %s s, "+
			"which is below the timer's 0.01 s step, so the ratio is approximate: **%s** from medians, **%s** from means.",
			thousands(smallest), seconds(loopSmall.Mean), percent(empty.Mean/loopSmall.Mean), seconds(goSmall.Mean),
			times(loopSmall.Median/goSmall.Median), times(loopSmall.Mean/goSmall.Mean)),
		fmt.Sprintf("**Whole process, large job** (n = %s): R %s s, Go %s s (**%s**).",
			thousands(largest), seconds(loopLarge.Mean), seconds(goLarge.Mean), times(loopLarge.Mean/goLarge.Mean)),
		fmt.Sprintf("**Parallel Go** (%s workers, n = %s): computation **%s** faster than one worker.",
			inputs.machineValue("parallel workers"), thousands(largest), times(oneWorker.MeanComputeMs/manyWorkers.MeanComputeMs)),
		fmt.Sprintf("**Memory** (n = %s, peak footprint): R %s MB, of which %s MB is R just starting; Go %s MB.",
			thousands(largest), megabytes(loopLarge.MeanFootprint), megabytes(empty.MeanFootprint), megabytes(goLarge.MeanFootprint)),
		fmt.Sprintf("**Cloud cost** (`%s`, one job at a time): Go saves **%s to %s** of the compute cost across the four sizes.",
			small.Instance, percent(lowSaving), percent(highSaving)),
	}
	return "- " + strings.Join(lines, "\n- ") + "\n", nil
}
