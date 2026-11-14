package record

import (
	"os"
	"strings"
	"testing"
)

func TestStringAndParseRoundTrip(t *testing.T) {
	original := Record{
		Dataset: "lognormal-1000.csv", Observations: 1000, Replicates: 1000, Seed: 431, Stream: "r",
		SampleMedian: 86.6483, SEMedian: 1.503050518008813, SampleMean: 97.8793036, SEMean: 1.7145995070818436,
	}
	parsed, err := Parse(original.String())
	if err != nil {
		t.Fatal(err)
	}
	if parsed != original {
		t.Errorf("round trip changed the record:\n got %+v\nwant %+v", parsed, original)
	}
	if got, want := len(strings.Split(original.String(), "\t")), len(strings.Split(Header, "\t")); got != want {
		t.Errorf("record has %d fields, header has %d", got, want)
	}
}

// TestParseReadsWhatRWrote uses the record R wrote on the Mac (testdata/r-reference).
func TestParseReadsWhatRWrote(t *testing.T) {
	content, err := os.ReadFile("../testdata/r-reference/lognormal-1000-record.tsv")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if lines[1] != Header {
		t.Errorf("R's header differs from ours:\n got %q\nwant %q", lines[1], Header)
	}
	parsed, err := Parse(lines[2])
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Dataset != "lognormal-1000.csv" || parsed.Observations != 1000 || parsed.Stream != "r" ||
		parsed.SampleMedian != 86.648300000000006 || parsed.SEMean != 1.7145995070818436 {
		t.Errorf("unexpected record %+v", parsed)
	}
	// Formatting what was parsed must give back R's own text: both sides print the same digits.
	if parsed.String() != lines[2] {
		t.Errorf("Go formats R's record differently:\n got %q\nwant %q", parsed.String(), lines[2])
	}
}

func TestParseRejectsMalformedLines(t *testing.T) {
	for _, line := range []string{
		"",
		"a\t1\t2",
		"a\tx\t1000\t431\tr\t1\t2\t3\t4",
		"a\t10\t1000\t-1\tr\t1\t2\t3\t4",
		"a\t10\t1000\t431\tr\t1\ttwo\t3\t4",
	} {
		if _, err := Parse(line); err == nil {
			t.Errorf("accepted malformed record %q", line)
		}
	}
}
