// Package bootstrap estimates standard errors by bootstrap resampling (Efron and Tibshirani
// 1993): draw B resamples of the data with replacement, evaluate a statistic on each, and
// take the standard deviation of the B values.
//
// The package knows no particular statistic and no particular random generator. A Statistic
// is any function of a sample, and a Stream supplies the random indices; RStream reproduces
// R's sample() exactly, BlockStream is built for parallel runs.
package bootstrap

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mioluigi/go-applied-statistics/stats"
)

// ErrInvalid marks every error caused by unusable arguments. Test for it with errors.Is.
var ErrInvalid = errors.New("invalid bootstrap request")

// A Statistic maps one resample to a number. It receives a scratch copy and may reorder it.
type Statistic func(resample []float64) float64

// An IndexSource draws a uniformly distributed index in [0, n).
type IndexSource interface {
	Index(n int) int
}

// A Stream supplies the random source for each replicate (replicates are numbered from 0).
type Stream interface {
	// Source returns the index source to use for one replicate.
	Source(replicate int) IndexSource
	// Sequential reports whether replicates must be drawn in order from one shared source.
	// A sequential stream runs with exactly one worker.
	Sequential() bool
}

// Config sets the size of a bootstrap run.
type Config struct {
	Replicates int    // B, the number of resamples; at least 2
	Stream     Stream // where the random indices come from
	Workers    int    // goroutines evaluating replicates; at least 1
}

// validate checks a request before any work is done.
func validate(data []float64, statistics []Statistic, config Config) error {
	switch {
	case len(data) < 2:
		return fmt.Errorf("%w: %d observation(s); at least 2 are needed", ErrInvalid, len(data))
	case len(statistics) == 0:
		return fmt.Errorf("%w: no statistic given", ErrInvalid)
	case config.Replicates < 2:
		return fmt.Errorf("%w: %d replicate(s); at least 2 are needed", ErrInvalid, config.Replicates)
	case config.Stream == nil:
		return fmt.Errorf("%w: no random stream given", ErrInvalid)
	case config.Workers < 1:
		return fmt.Errorf("%w: %d worker(s); at least 1 is needed", ErrInvalid, config.Workers)
	case config.Stream.Sequential() && config.Workers > 1:
		return fmt.Errorf("%w: this random stream is sequential and runs with 1 worker, not %d",
			ErrInvalid, config.Workers)
	}
	return nil
}

// Replicates evaluates every statistic on each of config.Replicates resamples of data.
// The result has one row per statistic; row s holds that statistic's value on resamples
// 0, 1, 2, … in that order, whatever the number of workers.
//
// Every statistic sees the same resample, in the order given; a statistic that reorders its
// argument therefore changes the order (not the content) seen by those after it.
func Replicates(data []float64, statistics []Statistic, config Config) ([][]float64, error) {
	if err := validate(data, statistics, config); err != nil {
		return nil, err
	}
	values := make([][]float64, len(statistics))
	for s := range values {
		values[s] = make([]float64, config.Replicates)
	}

	// Each worker owns a contiguous range of replicate numbers and writes each result at its
	// replicate's position, so no result depends on which goroutine finished first.
	var workers sync.WaitGroup
	for worker := 0; worker < config.Workers; worker++ {
		first := worker * config.Replicates / config.Workers
		last := (worker + 1) * config.Replicates / config.Workers
		workers.Add(1)
		go func() {
			defer workers.Done()
			for replicate := first; replicate < last; replicate++ {
				resample := drawResample(data, config.Stream.Source(replicate))
				for s, statistic := range statistics {
					values[s][replicate] = statistic(resample)
				}
			}
		}()
	}
	workers.Wait()
	return values, nil
}

// drawResample returns len(data) values drawn from data with replacement.
func drawResample(data []float64, source IndexSource) []float64 {
	resample := make([]float64, len(data))
	for i := range resample {
		resample[i] = data[source.Index(len(data))]
	}
	return resample
}

// StandardErrors returns the bootstrap standard error of each statistic: the standard
// deviation (denominator B - 1) of its values over the resamples.
func StandardErrors(data []float64, statistics []Statistic, config Config) ([]float64, error) {
	values, err := Replicates(data, statistics, config)
	if err != nil {
		return nil, err
	}
	standardErrors := make([]float64, len(values))
	for s, row := range values {
		standardErrors[s] = stats.StandardDeviation(row)
	}
	return standardErrors, nil
}
