package bootstrap

import "github.com/mioluigi/go-applied-statistics/rrand"

// rStream draws every replicate from one generator, in order: exactly what a loop of
// sample(x, n, replace = TRUE) does in R after set.seed.
type rStream struct {
	generator *rrand.Generator
}

// RStream returns the stream that reproduces R: for the same seed and data, replicate b is
// the b-th resample R would draw after set.seed(seed). It is sequential (one worker).
func RStream(seed uint32) Stream {
	return rStream{generator: rrand.New(seed)}
}

// Source returns the shared generator: every replicate continues where the last one stopped.
func (stream rStream) Source(int) IndexSource { return stream.generator }

// Sequential reports true: the draws of one replicate follow those of the one before.
func (stream rStream) Sequential() bool { return true }
