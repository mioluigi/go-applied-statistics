// Spike 002: how far apart are two correct bootstrap standard errors that used different
// random draws? Run before the official comparison, to check the 10 % limit of spec §6
// against evidence rather than a normal approximation.
//
// For each data set it takes the R-stream standard errors (seed 431, B = 1,000) as the
// reference and compares the block stream's for 200 other seeds. It prints, per data set and
// statistic, the standard deviation and the largest relative difference.
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/mioluigi/go-applied-statistics/bootstrap"
	"github.com/mioluigi/go-applied-statistics/dataset"
	"github.com/mioluigi/go-applied-statistics/stats"
)

func main() {
	const replicates, seeds = 1000, 200
	statistics := []bootstrap.Statistic{stats.Mean, stats.Median}
	fmt.Println("dataset\tstatistic\tsd_of_relative_difference\tlargest_relative_difference\tshare_above_10_percent")
	for _, name := range os.Args[1:] {
		data, err := dataset.ReadFile(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		reference, _ := bootstrap.StandardErrors(data, statistics,
			bootstrap.Config{Replicates: replicates, Stream: bootstrap.RStream(431), Workers: 1})
		differences := [2][]float64{}
		for seed := uint32(1); seed <= seeds; seed++ {
			got, _ := bootstrap.StandardErrors(data, statistics,
				bootstrap.Config{Replicates: replicates, Stream: bootstrap.BlockStream(seed), Workers: 2})
			for s := range got {
				differences[s] = append(differences[s], (got[s]-reference[s])/reference[s])
			}
		}
		for s, label := range []string{"mean", "median"} {
			largest, above := 0.0, 0
			for _, difference := range differences[s] {
				largest = math.Max(largest, math.Abs(difference))
				if math.Abs(difference) > 0.10 {
					above++
				}
			}
			fmt.Printf("%s\t%s\t%.4f\t%.4f\t%.3f\n", name, label,
				stats.StandardDeviation(differences[s]), largest, float64(above)/seeds)
		}
	}
}
