// Spike 003: what does the Mersenne Twister package the assignment suggests produce for the
// seed used throughout this project?
//
// A stand-alone program with its own go.mod, so the product never imports the package (it is
// licensed under the GPL). It prints the first three uniform numbers from
// github.com/seehuhn/mt19937 seeded with 431, through math/rand as that package documents.
package main

import (
	"fmt"
	"math/rand"

	"github.com/seehuhn/mt19937"
)

func main() {
	generator := rand.New(mt19937.New())
	generator.Seed(431)
	fmt.Println("# github.com/seehuhn/mt19937 v1.0.0 (64-bit Mersenne Twister); rand.New(mt19937.New()); Seed(431); Float64() three times")
	fmt.Println("seed\tvalues")
	fmt.Printf("431\t%.17g %.17g %.17g\n", generator.Float64(), generator.Float64(), generator.Float64())
}
