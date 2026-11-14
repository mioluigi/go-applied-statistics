package main

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// binary is the built program. Exit codes are tested on it, never through "go run", which
// reports every failure as 1 (SFEDD §5.5).
var binary string

func TestMain(m *testing.M) {
	folder, err := os.MkdirTemp("", "bootse")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(folder, "bootse")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		panic("cannot build the program: " + string(output))
	}
	code := m.Run()
	os.RemoveAll(folder)
	os.Exit(code)
}

// execute runs the built program and returns its exit code and both output streams.
func execute(t *testing.T, arguments ...string) (code int, stdout, stderr string) {
	t.Helper()
	command := exec.Command(binary, arguments...)
	var out, errOut bytes.Buffer
	command.Stdout, command.Stderr = &out, &errOut
	err := command.Run()
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), out.String(), errOut.String()
	}
	if err != nil {
		t.Fatal(err)
	}
	return 0, out.String(), errOut.String()
}

// TestNoArgumentsRunsToCompletion is FR-10 and AC-9: the summary names every field.
func TestNoArgumentsRunsToCompletion(t *testing.T) {
	code, stdout, stderr := execute(t)
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr)
	}
	for _, want := range []string{
		"Data set:", "lognormal-1000.csv", "n = 1000", "B = 1000", "seed 431", "stream r",
		"Sample median:", "bootstrap SE of the median:", "Sample mean:", "bootstrap SE of the mean:",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("summary lacks %q:\n%s", want, stdout)
		}
	}
	for _, want := range []string{"read_ms=", "compute_ms="} {
		if !strings.Contains(stderr, want) {
			t.Errorf("log lacks %q:\n%s", want, stderr)
		}
	}
}

// TestRecordEqualsRsRecord: with -format tsv the program prints, character for character,
// the line the R script printed on the Mac for the same data set, B and seed.
func TestRecordEqualsRsRecord(t *testing.T) {
	reference, err := os.ReadFile("testdata/r-reference/lognormal-1000-record.tsv")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(reference)), "\n")
	wantFields := strings.Split(lines[len(lines)-1], "\t")

	code, stdout, stderr := execute(t, "-data", "data/lognormal-1000.csv", "-format", "tsv", "-quiet")
	if code != 0 || stderr != "" {
		t.Fatalf("exit code %d, stderr %q (with -quiet nothing should be logged)", code, stderr)
	}
	gotFields := strings.Split(strings.TrimSpace(stdout), "\t")
	if len(gotFields) != len(wantFields) {
		t.Fatalf("got %d fields, want %d", len(gotFields), len(wantFields))
	}
	// Identical text for the first six fields (through the sample median). The summed
	// quantities may differ in the last digit between platforms (spec §6), so they are
	// compared as numbers: this also catches the two standard errors being swapped.
	for field := 0; field <= 5; field++ {
		if gotFields[field] != wantFields[field] {
			t.Errorf("field %d: got %q, R wrote %q", field, gotFields[field], wantFields[field])
		}
	}
	for field := 6; field <= 8; field++ {
		got, _ := strconv.ParseFloat(gotFields[field], 64)
		want, _ := strconv.ParseFloat(wantFields[field], 64)
		if math.Abs(got-want) > 1e-12*math.Abs(want) {
			t.Errorf("field %d: got %q, R wrote %q", field, gotFields[field], wantFields[field])
		}
	}
}

func TestMediansFileAndProfiles(t *testing.T) {
	folder := t.TempDir()
	medians := filepath.Join(folder, "medians.txt")
	cpu, memory := filepath.Join(folder, "cpu.prof"), filepath.Join(folder, "mem.prof")
	code, _, stderr := execute(t, "-replicates", "50", "-medians", medians, "-cpuprofile", cpu, "-memprofile", memory)
	if code != 0 {
		t.Fatalf("exit code %d: %s", code, stderr)
	}
	content, err := os.ReadFile(medians)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(string(content), "\n"); lines != 50 {
		t.Errorf("medians file has %d lines, want 50", lines)
	}
	for _, profile := range []string{cpu, memory} {
		if info, err := os.Stat(profile); err != nil || info.Size() == 0 {
			t.Errorf("profile %s missing or empty (%v)", filepath.Base(profile), err)
		}
	}
}

func TestBlockStreamRunsInParallel(t *testing.T) {
	_, oneWorker, _ := execute(t, "-stream", "block", "-workers", "1", "-format", "tsv", "-quiet")
	code, fourWorkers, stderr := execute(t, "-stream", "block", "-workers", "4", "-format", "tsv", "-quiet")
	if code != 0 {
		t.Fatalf("exit code %d: %s", code, stderr)
	}
	if oneWorker != fourWorkers || !strings.Contains(oneWorker, "\tblock\t") {
		t.Errorf("1 worker and 4 workers differ, or the stream is mislabelled:\n%s%s", oneWorker, fourWorkers)
	}
}

// TestInvalidInputIsExplained is FR-8 and AC-8: each bad request gets exit code 2, a message
// naming the problem, no result on standard output, and no stack trace.
func TestInvalidInputIsExplained(t *testing.T) {
	folder := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(folder, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	cases := []struct {
		name          string
		arguments     []string
		wantInMessage string
	}{
		{"missing file", []string{"-data", filepath.Join(folder, "absent.csv")}, "cannot read data set"},
		{"non-numeric value", []string{"-data", write("text.csv", "value\n1.5\noops\n2.5\n")}, `text.csv, line 3: "oops" is not a number`},
		{"too few observations", []string{"-data", write("one.csv", "value\n1.5\n")}, "has 1 observation(s); at least 2 are needed"},
		{"NaN", []string{"-data", write("nan.csv", "value\n1.5\nNaN\n2.5\n")}, "is not a finite number"},
		{"missing header", []string{"-data", write("bare.csv", "1.5\n2.5\n")}, "where the header line should be"},
		{"one replicate", []string{"-replicates", "1"}, "at least 2 resamples are needed"},
		{"replicates not a number", []string{"-replicates", "many"}, "invalid value"},
		{"no workers", []string{"-stream", "block", "-workers", "0"}, "-workers is 0; at least 1 is needed"},
		{"R stream with workers", []string{"-workers", "4"}, "use -stream block with -workers 4"},
		{"unknown stream", []string{"-stream", "fast"}, `-stream is "fast"; choose r or block`},
		{"unknown format", []string{"-format", "json"}, `-format is "json"; choose text or tsv`},
		{"seed too large", []string{"-seed", "4294967296"}, "must be between 0 and 4294967295"},
		{"negative seed", []string{"-seed", "-1"}, "invalid value"},
		{"unknown flag", []string{"-speed", "9"}, "flag provided but not defined"},
		{"stray argument", []string{"data/normal-100.csv"}, "the data file is given with -data"},
	}
	for _, testCase := range cases {
		code, stdout, stderr := execute(t, testCase.arguments...)
		if code != 2 {
			t.Errorf("%s: exit code %d, want 2", testCase.name, code)
		}
		if !strings.Contains(stderr, testCase.wantInMessage) {
			t.Errorf("%s: message lacks %q:\n%s", testCase.name, testCase.wantInMessage, stderr)
		}
		if stdout != "" || strings.Contains(stderr, "goroutine ") {
			t.Errorf("%s: printed a result or a stack trace:\n%s%s", testCase.name, stdout, stderr)
		}
	}
}

// TestFailureThatIsNotTheUsersInputExitsOne: a result file that cannot be written.
func TestFailureThatIsNotTheUsersInputExitsOne(t *testing.T) {
	code, _, stderr := execute(t, "-replicates", "10", "-medians", filepath.Join(t.TempDir(), "no-such-folder", "m.txt"))
	if code != 1 || !strings.Contains(stderr, "cannot write the resample medians") {
		t.Errorf("exit code %d, stderr %q; want 1 and an explanation", code, stderr)
	}
}

func TestHelpExitsZero(t *testing.T) {
	code, _, stderr := execute(t, "-h")
	if code != 0 || !strings.Contains(stderr, "-replicates") {
		t.Errorf("exit code %d, usage: %s", code, stderr)
	}
}
