package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// verifyAgainstR checks the Go re-implementation against the reference files R wrote on the Mac
// (docs/spikes/001-evidence). It returns the number of mismatches.
func verifyAgainstR(evidenceFolder string) int {
	return verifyIndexDraws(filepath.Join(evidenceFolder, "sample-indices.tsv")) +
		verifyBootstrap(filepath.Join(evidenceFolder, "bootstrap-se.tsv"))
}

// dataLines returns the tab-separated fields of each line after the comment and header lines.
func dataLines(path string) [][]string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("cannot open", path, err)
		os.Exit(2)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	var rows [][]string
	for lineNumber := 0; scanner.Scan(); lineNumber++ {
		if lineNumber >= 2 {
			rows = append(rows, strings.Split(scanner.Text(), "\t"))
		}
	}
	return rows
}

func verifyIndexDraws(path string) (mismatches int) {
	checked := 0
	for _, row := range dataLines(path) {
		n, _ := strconv.Atoi(row[0])
		twister := newRMersenneTwister(431)
		for _, field := range strings.Fields(row[1]) {
			want, _ := strconv.Atoi(field)
			if twister.unifIndex(n)+1 != want {
				mismatches++
			}
			checked++
		}
	}
	fmt.Printf("index draws: %d checked, %d mismatches\n", checked, mismatches)
	return mismatches
}

func verifyBootstrap(path string) (mismatches int) {
	for _, row := range dataLines(path) {
		n, _ := strconv.Atoi(row[0])
		replicates, _ := strconv.Atoi(row[1])
		seed, _ := strconv.Atoi(row[2])
		wantMedianSE, _ := strconv.ParseFloat(row[3], 64)
		wantMeanSE, _ := strconv.ParseFloat(row[4], 64)

		data := make([]float64, n)
		for i := range data {
			data[i] = float64((37 * (i + 1)) % 101)
		}
		twister := newRMersenneTwister(uint32(seed))
		medians := make([]float64, replicates)
		means := make([]float64, replicates)
		resample := make([]float64, n)
		for b := range medians {
			sum := 0.0
			for i := range resample {
				resample[i] = data[twister.unifIndex(n)]
				sum += resample[i]
			}
			sort.Float64s(resample)
			medians[b] = (resample[(n-1)/2] + resample[n/2]) / 2
			means[b] = sum / float64(n)
		}
		gotMedianSE, gotMeanSE := standardDeviation(medians), standardDeviation(means)
		fmt.Printf("n=%d B=%d  SE median: R %.17g  Go %.17g (rel. diff %.1e)   SE mean: R %.17g  Go %.17g (rel. diff %.1e)\n",
			n, replicates, wantMedianSE, gotMedianSE, relativeDifference(gotMedianSE, wantMedianSE),
			wantMeanSE, gotMeanSE, relativeDifference(gotMeanSE, wantMeanSE))
		if relativeDifference(gotMedianSE, wantMedianSE) > 1e-12 || relativeDifference(gotMeanSE, wantMeanSE) > 1e-12 {
			mismatches++
		}
	}
	return mismatches
}

// standardDeviation uses the n-1 denominator, as R's sd() does.
func standardDeviation(values []float64) float64 {
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	sumOfSquares := 0.0
	for _, value := range values {
		sumOfSquares += (value - mean) * (value - mean)
	}
	return math.Sqrt(sumOfSquares / float64(len(values)-1))
}

func relativeDifference(got, want float64) float64 {
	return math.Abs(got-want) / math.Abs(want)
}
