package dataset

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixtures copy the shape of the real files in data/: header "value", 4 decimals,
// negative values possible, "\n" line endings.

func TestReadAcceptsTheShapeOfTheRealFiles(t *testing.T) {
	values, err := Read(strings.NewReader("value\n-26.8707\n76.9465\n102.3419\n"), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{-26.8707, 76.9465, 102.3419}
	if len(values) != len(want) {
		t.Fatalf("got %d values, want %d", len(values), len(want))
	}
	for i := range want {
		if values[i] != want[i] {
			t.Errorf("value %d: got %v, want %v", i, values[i], want[i])
		}
	}
}

func TestReadToleratesBlankLinesAndWindowsLineEndings(t *testing.T) {
	values, err := Read(strings.NewReader("value\r\n1.5\r\n\r\n2.5\r\n\r\n"), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[0] != 1.5 || values[1] != 2.5 {
		t.Errorf("got %v, want [1.5 2.5]", values)
	}
}

func TestReadRejectsBadContentWithFileAndLine(t *testing.T) {
	cases := []struct {
		name, content, wantInMessage string
	}{
		{"text among numbers", "value\n1.0\nabc\n3.0\n", `data.csv, line 3: "abc" is not a number`},
		{"two numbers on a line", "value\n1.0, 2.0\n3.0\n", `data.csv, line 2: "1.0, 2.0" is not a number`},
		{"not a number", "value\n1.0\nNaN\n", `data.csv, line 3: "NaN" is not a finite number`},
		{"infinite", "value\n1.0\n-Inf\n", `data.csv, line 3: "-Inf" is not a finite number`},
		{"no header", "1.0\n2.0\n3.0\n", `data.csv, line 1: found the number "1.0" where the header line should be`},
		{"one observation", "value\n1.0\n", "data.csv has 1 observation(s); at least 2 are needed"},
		{"header only", "value\n", "data.csv has 0 observation(s)"},
		{"empty file", "", "data.csv has 0 observation(s)"},
	}
	for _, testCase := range cases {
		_, err := Read(strings.NewReader(testCase.content), "data.csv")
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: error %v is not ErrInvalid", testCase.name, err)
			continue
		}
		if !strings.Contains(err.Error(), testCase.wantInMessage) {
			t.Errorf("%s: message %q lacks %q", testCase.name, err, testCase.wantInMessage)
		}
	}
}

func TestReadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "two.csv")
	if err := os.WriteFile(path, []byte("value\n1.0\n2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if values, err := ReadFile(path); err != nil || len(values) != 2 {
		t.Errorf("got %v, %v", values, err)
	}
	_, err := ReadFile(filepath.Join(t.TempDir(), "missing.csv"))
	if err == nil || errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "cannot read data set") {
		t.Errorf("missing file: got %v, want a cannot-read error that is not ErrInvalid", err)
	}
}

// TestRealDataSetsParse reads every committed data set: the reader must accept what R wrote.
func TestRealDataSetsParse(t *testing.T) {
	paths, err := filepath.Glob("../data/*.csv")
	if err != nil || len(paths) != 12 {
		t.Fatalf("found %d data sets (%v), want 12", len(paths), err)
	}
	for _, path := range paths {
		values, err := ReadFile(path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if len(values) < 100 {
			t.Errorf("%s: %d observations, want at least 100", path, len(values))
		}
	}
}
