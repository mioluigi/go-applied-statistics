# Plan 001: how the spec will be met

| Field | Value |
|---|---|
| Status | **APPROVED 2026-11-14** by Luis |
| Module | `github.com/mioluigi/go-applied-statistics`, `go 1.24`, **standard library only** (if Q2 = a): no `go.sum` needed |

## 1. Layout

```
main.go, main_test.go        thin command-line program "bootse": flags, validation, output, logging, profiles
rrand/                       R-compatible generator: set.seed scrambling, unif_rand, sample index   (FR-5, FR-R3)
stats/                       Mean, StandardDeviation (two-pass), Median, MedianInPlace              (FR-R4)
bootstrap/                   the resampling engine: data + statistics + B + stream + workers       (FR-R1, FR-R2, FR-6)
dataset/                     read and validate a one-column CSV; record format shared with R        (FR-4, FR-8)
internal/quality/            test: doc comments on exported identifiers, function length            (NFR-8)
cmd/compare/                 R record against Go record by spec §6                                  (FR-7)
cmd/tables/                  results + prices -> results/tables.md (tables, headline, cost, effort) (FR-13, FR-15, NFR-7)
r/original/                  the course's program and listing, unchanged
r/generate-data.R            seeded generator of the 12 data sets and their manifest                (FR-1)
r/bootstrap-se.R             idiomatic base R: sample() loop                                        (FR-2)
r/bootstrap-se-boot.R        the same with the CRAN boot package                                    (FR-2)
r/startup-empty.R, r/startup-boot.R   start-up conditions                                           (NFR-3)
data/                        12 CSV files + MANIFEST.tsv
testdata/r-reference/        R's saved reference values (from the spike), with provenance
prices/aws-ec2.tsv           saved prices with source, region, instance, date                       (FR-13)
scripts/                     check-inputs.sh · compare.sh · measure.sh · readme.sh · build.sh · build-macos-app.sh · acceptance.sh · review-bundle.sh · env-probe.sh
results/                     raw runs, comparison, benchmarks, profiles, machine, tables            (official machine only)
```

## 2. The generator (`rrand`)

The spike code, productised: `New(seed)` reproduces `set.seed`; `Float64()` is `unif_rand()`; `Index(n)` is
R's rejection sampling. Tests read `testdata/r-reference/` (the 14,000 draws and the `runif` values R wrote
on the Mac). A performance step precomputes the bit count for n instead of taking a logarithm per draw;
the reference test proves the faster form draws the same indices.

## 3. The engine (`bootstrap`)

```go
// A Statistic maps one resample to a number. It may reorder its argument.
type Statistic func(resample []float64) float64

// IndexSource draws an index in [0, n).
type IndexSource interface{ Index(n int) int }

type Config struct {
    Replicates int
    Seed       uint32
    Stream     Stream // RStream or BlockStream
    Workers    int
}

// Replicates returns, for each statistic, its value on each of the B resamples.
func Replicates(data []float64, statistics []Statistic, config Config) ([][]float64, error)
// StandardErrors is the standard deviation (B − 1) of each row of Replicates.
func StandardErrors(...) ([]float64, error)
```

- **One path.** Replicate numbers 0…B−1 are split into contiguous ranges, one per worker. Each worker owns
  one resample buffer and writes replicate b's statistics into position b of preallocated result slices.
  No channel carries results, so the reduction order is fixed (review DR-6).
- **R stream:** one `rrand` generator, one worker, replicates in order: R's exact resamples.
- **Block stream:** replicate b gets its own standard-library generator (`math/rand/v2` PCG) whose two seed
  words come from mixing (seed, b) with SplitMix64. Cheap to create, independent of worker count, and it is
  the second random source of FR-R2.
- Statistics are evaluated in the given order on the same resample; the command line passes mean, then
  median (the median may reorder the buffer).

## 4. Statistics (`stats`)

- `Mean` and `StandardDeviation`: R's two-pass form (mean, then a correction pass; deviations about the
  corrected mean; n − 1).
- `Median` (copies, sorts): the plain, obviously correct version; kept as the reference and the "before".
- `MedianInPlace` (selection, O(n) on average, reorders its argument): the "after". A property test checks
  it equals `Median` on random inputs, including ties and even n.

## 5. Command line (`bootse`)

`bootse [-data FILE] [-replicates 1000] [-seed 431] [-stream r|block] [-workers N] [-format text|tsv]
[-medians FILE] [-cpuprofile FILE] [-memprofile FILE] [-quiet]`

No arguments: the built-in copy (`go:embed`) of `data/lognormal-1000.csv`. Log lines (`log/slog`, standard
error): settings, `read_ms`, `compute_ms`. Exit codes 0 / 1 / 2 (FR-9).

## 6. Measurement design (`scripts/measure.sh`, Mac only)

1. Guards: Mac, clean and current copy, data gate, R and `boot` present. Records machine, versions, commit, load average.
2. **Correctness first** (`scripts/compare.sh`): R (`sample()` loop), R (`boot`), Go (R stream), Go (block
   stream, 1 and all workers) on all 12 data sets; `cmd/compare` applies §6. Any failure stops the script.
3. **Whole-process runs** with `/usr/bin/time -l`, conditions interleaved (R, Go, R, Go …) so drift affects
   all equally: start-up ×3; five implementations × four sizes on the lognormal data; three shapes at
   n = 10,000; the course program. Each run appends one line to `results/runs.tsv`
   (condition, n, run, real, user, sys, max RSS, peak footprint, read_ms, compute_ms).
4. **Go benchmarks** (`-benchmem`, repeated) for generator, median (before/after), engine (before/after per
   step, 1 and all workers), and the course-study volume (P4) → `results/benchmarks.txt`.
5. **Profiles** of one n = 100,000 run → `results/cpu.prof`, `results/mem.prof`, with `go tool pprof -top`
   text saved beside them.

## 7. The suggested package (`seehuhn/mt19937`), without importing it

The README shows, for seed 431, the first uniform numbers from R, from our generator and from the suggested
package. The third column comes from a stand-alone snippet under `docs/spikes/` with its own `go.mod`, run
once on the Mac, output saved with provenance. The product never imports the GPL package.

## 8. Cost model (`cmd/tables`)

From `results/runs.tsv` (mean whole-process seconds) and `prices/aws-ec2.tsv` (USD per hour, vCPUs):

- right-sized: cost per job = seconds × (hourly price ÷ vCPUs) ÷ 3600 × vCPUs used (1 for R and one-core Go);
- multi-vCPU instance: R runs one job per vCPU at once; Go runs jobs one at a time on all vCPUs;
- outputs: USD per 1,000,000 jobs, percentage saved, jobs per dollar of rewrite cost; per data size.

## 9. Deadline mode: order of cuts (from review R1, accepted)

1. Narrow the timing matrix (200 runs only at n = 1,000; 50 at n = 100 and 10,000; 20 at n = 100,000 and the course program).
2. Limit performance history to two steps (buffer reuse + selection; workers).
3. Time `boot` at n = 1,000 only (still compared for correctness everywhere).
4. n = 100,000 for one shape only.

Never cut: R-stream parity tests, deterministic parallel stream, input validation and no-argument run,
build scripts, profiles and logging, generated README numbers, AI dialogs.

## 10. Round trips with the Mac (kept to four)

| # | Mac step | Delivers |
|---|---|---|
| M1 | Generate data; smoke-run the R scripts | `data/`, manifest; first real test of R code written without R |
| M2 | `scripts/measure.sh` (close other programs first) | `results/` |
| M3 | Export AI dialogs to `docs/ai-dialogs/`; Gemini R2 | dialogs; review |
| M4 | `scripts/acceptance.sh` on a fresh clone | acceptance record; submit |
