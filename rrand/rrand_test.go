package rrand

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// referenceRows returns the tab-separated fields of each data line of a reference file,
// skipping its provenance line and its header line.
func referenceRows(t *testing.T, name string) [][]string {
	t.Helper()
	file, err := os.Open("../testdata/r-reference/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	var rows [][]string
	for line := 0; scanner.Scan(); line++ {
		if line >= 2 {
			rows = append(rows, strings.Split(scanner.Text(), "\t"))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatalf("%s has no data rows", name)
	}
	return rows
}

func atoi(t *testing.T, text string) int {
	t.Helper()
	value, err := strconv.Atoi(text)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

// TestFloat64MatchesRunif compares with set.seed(seed); runif(k), to the last digit.
func TestFloat64MatchesRunif(t *testing.T) {
	for _, row := range referenceRows(t, "runif.tsv") {
		generator := New(uint32(atoi(t, row[0])))
		for position, field := range strings.Fields(row[1]) {
			want, err := strconv.ParseFloat(field, 64)
			if err != nil {
				t.Fatal(err)
			}
			if got := generator.Float64(); got != want {
				t.Errorf("seed %s, draw %d: got %.17g, R gave %.17g", row[0], position+1, got, want)
			}
		}
	}
}

// TestIndexMatchesSampleInt compares with set.seed(431); sample.int(n, 2000, replace = TRUE)
// for sizes that need 4 to 17 random bits, so both one and two uniform numbers per candidate.
func TestIndexMatchesSampleInt(t *testing.T) {
	checked := 0
	for _, row := range referenceRows(t, "sample-indices.tsv") {
		n := atoi(t, row[0])
		generator := New(431)
		mismatches := 0
		for _, field := range strings.Fields(row[1]) {
			if generator.Index(n)+1 != atoi(t, field) {
				mismatches++
			}
			checked++
		}
		if mismatches > 0 {
			t.Errorf("n = %d: %d draws differ from R", n, mismatches)
		}
	}
	if checked != 14000 {
		t.Errorf("checked %d draws, want 14000 (is the reference file complete?)", checked)
	}
}

// TestIndexMatchesSampleIntAtBitBoundaries covers the sizes where R's 16-bits-per-uniform loop
// changes behaviour: n = 1 and 2, exactly 16 bits (40,000 and 65,536) and 17 bits (65,537).
// The uniform number R drew after the indices proves the same count of numbers was consumed.
func TestIndexMatchesSampleIntAtBitBoundaries(t *testing.T) {
	for _, row := range referenceRows(t, "sample-indices-edge.tsv") {
		n := atoi(t, row[0])
		generator := New(431)
		mismatches := 0
		for _, field := range strings.Fields(row[1]) {
			if generator.Index(n)+1 != atoi(t, field) {
				mismatches++
			}
		}
		if mismatches > 0 {
			t.Errorf("n = %d: %d draws differ from R", n, mismatches)
		}
		wantNext, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			t.Fatal(err)
		}
		if got := generator.Float64(); got != wantNext {
			t.Errorf("n = %d: next uniform %.17g, R gave %.17g (a different number of draws was consumed)", n, got, wantNext)
		}
	}
}

// TestIndexStaysInRange covers sizes the reference file does not, including n = 1 and
// exact powers of two, where the rejection mask is tightest.
func TestIndexStaysInRange(t *testing.T) {
	for _, n := range []int{1, 2, 3, 16, 17, 65536, 65537, 1 << 20} {
		generator := New(7)
		seen := map[int]bool{}
		for draw := 0; draw < 2000; draw++ {
			index := generator.Index(n)
			if index < 0 || index >= n {
				t.Fatalf("n = %d: index %d out of range", n, index)
			}
			seen[index] = true
		}
		if n <= 3 && len(seen) != n {
			t.Errorf("n = %d: saw only %d distinct indices in 2000 draws", n, len(seen))
		}
	}
}

// TestSeedsGiveDifferentStreams guards against a constructor that ignores its seed.
func TestSeedsGiveDifferentStreams(t *testing.T) {
	if New(1).Float64() == New(2).Float64() {
		t.Error("seeds 1 and 2 gave the same first number")
	}
	first, second := New(431), New(431)
	for draw := 0; draw < 1000; draw++ {
		if first.Float64() != second.Float64() {
			t.Fatal("the same seed gave different streams")
		}
	}
}

// TestFloat64IsNeverZeroOrOne forces the two extreme 32-bit outputs. An all-zero state word
// tempers to 0 and an all-ones word to the largest output, so the state is set directly.
func TestFloat64IsNeverZeroOrOne(t *testing.T) {
	generator := New(431)
	generator.next = 0
	generator.state[0] = 0
	if got, want := generator.Float64(), 0.5*toUnitInterval; got != want {
		t.Errorf("zero output: got %g, want R's nudged value %g", got, want)
	}
	if largest := float64(^uint32(0)) * toUnitInterval; largest >= 1 {
		t.Errorf("largest output maps to %g, which is not below 1", largest)
	}
}

func BenchmarkFloat64(b *testing.B) {
	generator := New(431)
	b.ReportAllocs()
	for b.Loop() {
		generator.Float64()
	}
}

func BenchmarkIndex(b *testing.B) {
	for _, n := range []int{100, 100000} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			generator := New(431)
			b.ReportAllocs()
			for b.Loop() {
				generator.Index(n)
			}
		})
	}
}
