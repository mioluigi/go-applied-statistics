package bootstrap

import (
	"math/rand/v2"

	"github.com/mioluigi/go-applied-statistics/rrand"
)

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

// blockStream gives every replicate its own generator, derived from the seed and the
// replicate number alone, so replicates can be drawn in any order and on any goroutine.
type blockStream struct {
	seed uint32
}

// BlockStream returns the stream for parallel runs. For a given seed its results do not
// depend on the number of workers. Its draws differ from R's, so its standard errors agree
// with RStream's only up to Monte Carlo error; use RStream when a result must be checked
// against R number for number.
func BlockStream(seed uint32) Stream {
	return blockStream{seed: seed}
}

// Source returns a new PCG generator (math/rand/v2) for one replicate. The pair (seed,
// replicate) is unique, and SplitMix64 turns it into two well-mixed seed words, so
// neighbouring replicates do not start from neighbouring generator states.
func (stream blockStream) Source(replicate int) IndexSource {
	key := uint64(stream.seed)<<32 | uint64(uint32(replicate))
	return pcgSource{rand.New(rand.NewPCG(splitMix64(key), splitMix64(^key)))}
}

// Sequential reports false: each replicate is independent of the others.
func (stream blockStream) Sequential() bool { return false }

// pcgSource adapts a standard-library generator to IndexSource.
type pcgSource struct {
	generator *rand.Rand
}

// Index returns an unbiased index in [0, n).
func (source pcgSource) Index(n int) int { return source.generator.IntN(n) }

// splitMix64 is the standard 64-bit mixing function (Steele, Lea and Flood 2014): it maps
// consecutive inputs to outputs that look unrelated.
func splitMix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}
