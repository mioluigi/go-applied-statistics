// Spike 001: can Go reproduce R's random draws for the same seed?
//
// A throw-away program (not the product). It re-implements, from the R sources (src/main/RNG.c),
// the three steps R takes between set.seed() and sample():
//
//  1. set.seed scrambles the seed with a linear congruential generator and fills the 32-bit
//     Mersenne Twister state from it;
//  2. unif_rand turns a 32-bit output into a double in (0, 1);
//  3. sample() draws an index by rejection sampling on random bits (R >= 3.6.0).
//
// It prints the same three probes scripts/env-probe.sh ran in R, for comparison by eye and by test.
package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

const (
	stateSize   = 624
	middleWord  = 397
	matrixA     = 0x9908b0df
	upperMask   = 0x80000000
	lowerMask   = 0x7fffffff
	toUnitScale = 2.3283064365386963e-10 // 2^-32, as in R
)

// rMersenneTwister is the 32-bit MT19937 with R's seeding.
type rMersenneTwister struct {
	state [stateSize]uint32
	next  int
}

// newRMersenneTwister mirrors set.seed(seed) for RNGkind "Mersenne-Twister".
func newRMersenneTwister(seed uint32) *rMersenneTwister {
	for scramble := 0; scramble < 50; scramble++ {
		seed = 69069*seed + 1
	}
	// R fills 625 words: the first is the position counter, which FixupSeeds then sets to 624.
	seed = 69069*seed + 1
	twister := &rMersenneTwister{next: stateSize}
	for word := range twister.state {
		seed = 69069*seed + 1
		twister.state[word] = seed
	}
	return twister
}

func (twister *rMersenneTwister) refill() {
	twist := func(upper, lower uint32) uint32 {
		mixed := (upper & upperMask) | (lower & lowerMask)
		if mixed&1 == 1 {
			return (mixed >> 1) ^ matrixA
		}
		return mixed >> 1
	}
	state := &twister.state
	for word := 0; word < stateSize; word++ {
		state[word] = state[(word+middleWord)%stateSize] ^ twist(state[word], state[(word+1)%stateSize])
	}
	twister.next = 0
}

// uint32 returns the next tempered 32-bit output.
func (twister *rMersenneTwister) uint32() uint32 {
	if twister.next >= stateSize {
		twister.refill()
	}
	output := twister.state[twister.next]
	twister.next++
	output ^= output >> 11
	output ^= (output << 7) & 0x9d2c5680
	output ^= (output << 15) & 0xefc60000
	output ^= output >> 18
	return output
}

// unifRand mirrors R's unif_rand(): a double strictly inside (0, 1).
func (twister *rMersenneTwister) unifRand() float64 {
	value := float64(twister.uint32()) * toUnitScale
	if value <= 0 {
		return 0.5 * toUnitScale
	}
	if 1-value <= 0 {
		return 1 - 0.5*toUnitScale
	}
	return value
}

// unifIndex mirrors R_unif_index(n) with sample.kind "Rejection": an index in [0, n).
func (twister *rMersenneTwister) unifIndex(n int) int {
	bits := int(math.Ceil(math.Log2(float64(n))))
	for {
		var candidate uint64
		for filled := 0; filled <= bits; filled += 16 {
			candidate = 65536*candidate + uint64(math.Floor(twister.unifRand()*65536))
		}
		candidate &= (1 << bits) - 1
		if candidate < uint64(n) {
			return int(candidate)
		}
	}
}

func main() {
	if len(os.Args) == 3 && os.Args[1] == "-verify" {
		os.Exit(verifyAgainstR(os.Args[2]))
	}
	twister := newRMersenneTwister(431)
	fmt.Printf("set.seed(431); runif(3): %.17f %.17f %.17f\n",
		twister.unifRand(), twister.unifRand(), twister.unifRand())

	twister = newRMersenneTwister(431)
	fmt.Print("set.seed(431); sample(10, 5, replace = TRUE):")
	for draw := 0; draw < 5; draw++ {
		fmt.Printf(" %d", twister.unifIndex(10)+1)
	}
	fmt.Println()

	fmt.Printf("study-shaped workload (100 x {25,100,225,400} x B=100, sort-based median): %v\n",
		timeStudyShapedWorkload())
}

// timeStudyShapedWorkload does the same amount of resampling as the course's R program
// (uniform data instead of normal; only the cost matters here) and returns the time taken.
func timeStudyShapedWorkload() time.Duration {
	twister := newRMersenneTwister(9999)
	start := time.Now()
	var checksum float64
	for iteration := 0; iteration < 100; iteration++ {
		for _, n := range []int{25, 100, 225, 400} {
			sample := make([]float64, n)
			for i := range sample {
				sample[i] = twister.unifRand()
			}
			resample := make([]float64, n)
			for b := 0; b < 100; b++ {
				for i := range resample {
					resample[i] = sample[twister.unifIndex(n)]
				}
				sort.Float64s(resample)
				checksum += (resample[(n-1)/2] + resample[n/2]) / 2
			}
		}
	}
	_ = checksum
	return time.Since(start)
}
