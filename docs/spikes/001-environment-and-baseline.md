# Spike 001: environment, R baseline and R↔Go comparability

| Field | Value |
|---|---|
| Status | **Complete** (phase 1 gate met 2026-11-14). P4–P6 are checked by the official measurement. |
| Date | 2026-11-14 |

## 1. Environment probe (SFEDD §5.7)

### 1.1 Claude's workspace (measured 2026-11-14)

| Check | Result | Consequence |
|---|---|---|
| Go toolchain | go 1.24.7 linux/amd64 | Go code and tests can be developed here |
| R | **absent**; `apt-get install r-base-core` has no candidate | Everything that runs R is a **Mac** task |
| Python | 3.13 with numpy 2.5.3, scipy 1.18.1 | An independent source of known answers for tests |
| `proxy.golang.org` | **unreachable** (name does not resolve) | `go.sum` is generated on the Mac (L-W7-06) |
| `github.com/seehuhn/mt19937` | v1.0.0 fetched with `GOPROXY=direct` | The suggested generator can be used and tested here |
| GitHub read | works (`git ls-remote`) | |
| GitHub push | CLI token invalid; push works only for repositories attached to the session | The repository must exist and be attached before the first push |
| Canvas (jump-start R code, performance page) | course login required | Luis downloads them |

### 1.2 Luis's Mac

Measured 2026-11-14 by `scripts/env-probe.sh`; raw output in `docs/spikes/env-mac.txt` (commit `80060ad`).

| Check | Result | Consequence |
|---|---|---|
| Machine | MacBook Pro, Apple M1 Max, 10 CPUs (8 performance), 64 GB | Official machine; multi-core runs are meaningful |
| Go | go 1.27.2 darwin/arm64 (workspace: 1.24.7) | `go.mod` declares `go 1.24`, so both build it; the Mac is the official build |
| R | 4.4.2; RNG `Mersenne-Twister / Inversion / Rejection` | The defaults assumed in §3 are the ones in use |
| R packages | `boot` 1.3.32, `bench` 1.1.4, `microbenchmark` 1.5.0 present; **`bootstrap` missing** | Installing `bootstrap` is an owner decision (Q in the spec); `boot` alone covers the method |
| `median(c(1, 2, 3, 4))` | 2.5 | R averages the two middle values (assumption 2 confirmed for R) |
| `/usr/bin/time -l` | present | Whole-process time and peak memory for R and Go, same tool |
| Go module proxy | ok | `go.sum` can be generated there |
| GitHub CLI | missing | Mac steps use `git` and the browser only |

**Phase 0 gate: met.** Steps and rubric written down, push access verified by a fresh clone, work tagged
Claude or Mac.

## 2. Inputs requested at intake (all received 2026-11-14: the two R files in commit `80060ad`; the guidance page read the same day)

| Input | Needed for | Owner |
|---|---|---|
| `run-bootstrap-median.R` (Canvas) | "Run inherited code unchanged" (phase 1 gate) | Luis |
| `listing-from-100-run-bootstrap-median.txt` (Canvas) | Expected output of the inherited code | Luis |
| Canvas page *Performance Comparisons of R, Python, and Go* | S6 says to check it; it may set the expected measuring method | Luis |

Week 7's baseline code never arrived (L-W7-05). This week the baseline is a required deliverable, so these
files are requested at intake.

## 3. Riskiest assumptions to test in phase 1 (recorded now, tested next)

1. **"Same seed gives the same random draws in R and Go."** Both can use the Mersenne Twister, but R
   scrambles the seed in `set.seed`, turns generator output into a uniform number its own way, and since R
   3.6.0 draws `sample` indices by rejection sampling. If draws differ, results are comparable
   **statistically, not bit for bit**, and the spec's comparison rule must say so.
2. **"Median means the same thing on both sides."** For an even number of observations R averages the two
   middle values; a Go library may not.
3. **"Go is faster and saves cloud cost."** The assignment's premise. R's `sample` and `median` run in
   compiled C; the gap may be smaller than assumed, and whole-process time includes R's start-up.

## 4. Inherited code, read before it is run

`r/original/run-bootstrap-median.R` (T. W. Miller, revised May 23, 2024) and its listing arrived in commit
`80060ad`. Reading it (running it is the Mac part, §7):

1. **What it computes.** 100 iterations × 4 sample sizes (25, 100, 225, 400) × B = 100 resamples of a
   Normal(100, 10) sample. For each sample it keeps the **mean of the B bootstrap medians**, and reports the
   **standard deviation of that mean across the 100 samples**. That is a Monte Carlo estimate of how the
   median varies from sample to sample. The textbook bootstrap standard error is different: the standard
   deviation of the B bootstrap medians **within one sample**. Both are legitimate; the spec must say which
   one each number is, and the R and Go programs must compute the same one.
2. **Its cost is mostly not statistics.** Each of the 40,000 resamples builds a one-row data frame and
   `rbind`s it onto a growing one. That is R's slowest idiom. A comparison of Go against this program measures
   data-frame handling more than resampling. A fair comparison (constitution 3.4) also needs an **idiomatic R**
   version (preallocated vectors, or the `boot` package).
3. **Only the symmetric case.** It draws from a normal distribution; the assignment's example asks for
   positively skewed, symmetric and negatively skewed shapes.
4. **Seed 9999, R's default generator.** The listing should be reproduced exactly on R 4.4.2 (prediction P1).

## 5. Dependency audit: `github.com/seehuhn/mt19937` v1.0.0 (the assignment's suggestion)

| Property | Finding (from its source) | Consequence |
|---|---|---|
| Algorithm | **64-bit** Mersenne Twister (MT19937-64: 312 words of 64 bits) | **Not R's generator.** R uses the 32-bit MT19937 (624 words). The same seed cannot give the same stream. |
| Seeding | `Seed(int64)` (Knuth-style initialisation), `SeedFromSlice` | R scrambles the seed with a linear congruential generator instead |
| Default | An unseeded generator uses a fixed default seed | Reproducible, but silently identical across runs |
| Concurrency | Not safe for concurrent use (documented) | One generator per worker |
| Licence | **GPL-3.0** | Importing it puts our program under GPL terms; an owner decision |
| Interface | `math/rand` `Source64` | Works with `math/rand` version 1 |

The assignment says the Mersenne Twister "is the random number generator algorithm used in R and Python".
True of the family; this package is a different member of it. Using it would give *a* Mersenne Twister, not
*R's* numbers.

## 6. Riskiest assumption tested: can Go draw R's bootstrap samples?

`docs/spikes/001-rparity/main.go` (standard library only, 130 lines) re-implements from the R sources the
three steps between `set.seed()` and `sample()`: the seed scrambling, `unif_rand()`, and the rejection
sampling of an index.

| Probe (seed 431) | R 4.4.2 on the Mac | Go spike in the workspace |
|---|---|---|
| `runif(3)` | 0.75096705253235996 0.58476094505749643 0.80907909828238189 | identical, all 17 digits |
| `sample(10, 5, replace = TRUE)` | 3 6 10 6 8 | identical |

Reproduce: `go run docs/spikes/001-rparity/main.go` and compare with `docs/spikes/env-mac.txt`.

**Result: assumption 1 of §3 is overturned in the favourable direction.** A Go program can draw exactly R's
resamples, so the bootstrap can be compared **number for number**, not only statistically. This rests on
eight draws so far; §7 extends it to 14,000 index draws and two whole bootstraps before the spec relies on it.

Workspace timing of a study-shaped workload in plain Go (same resampling volume as the R program,
`sort.Float64s` for the median, 2 slow virtual CPUs): 0.65–0.72 s over three runs. The sort dominates, which
names the first performance steps: a faster sort or a selection algorithm, and no per-draw logarithm.

## 7. Mac part: the R program run unchanged, and R's reference values

`scripts/spike-r-baseline.sh`, run on the Mac at commit `bf5b5ba`; raw evidence in `docs/spikes/001-evidence/`.

**Inherited code, unchanged, three runs** (`/usr/bin/time -l Rscript r/original/run-bootstrap-median.R`):

| Run | Whole-process time | User CPU | Maximum resident set | Peak memory footprint |
|---|---|---|---|---|
| 1 | 5.24 s | 5.01 s | 100.6 MB | 80.7 MB |
| 2 | 5.12 s | 4.93 s | 113.7 MB | 89.1 MB |
| 3 | 5.08 s | 4.92 s | 111.1 MB | 93.8 MB |

- The three outputs are byte-identical, and equal the course listing line for line. The only difference is
  four console lines the listing file never contained (the "Running study" header and the "Run Complete"
  footer). Check: `diff r/original/listing-from-100-run-bootstrap-median.txt docs/spikes/001-evidence/original-run-1.txt`.
- **Memory varies between identical runs** (81 to 94 MB footprint): R's garbage collection timing. Memory
  needs repeated runs and a spread, like time.

**Reference values reproduced by the Go spike** (`cd docs/spikes/001-rparity && go run . -verify ../001-evidence`):

| Reference from R 4.4.2 | Go spike |
|---|---|
| 14,000 index draws: `sample.int(n, 2000, replace = TRUE)` for n = 10, 25, 100, 225, 400, 1,000, 100,000 | **14,000 of 14,000 identical** |
| Bootstrap SE of the median, n = 25 (odd), B = 200: 9.6785923952695132 | 9.6785923952695132 (identical) |
| Bootstrap SE of the median, n = 100 (even), B = 200: 5.0143412419232352 | 5.0143412419232352 (identical) |
| Bootstrap SE of the mean, n = 25: 5.3149252669218976 | identical |
| Bootstrap SE of the mean, n = 100: 2.9716934244711761 | 2.9716934244711748 (relative difference 4.5e-16: last bits, from summation order) |

The verifier was mutation-checked: widening the rejection mask by one bit gives 13,665 mismatches of 14,000.

**Conclusion.** For the same data and seed, a Go program gives **R's bootstrap standard error of the median
digit for digit**; the mean agrees to the last bits. The spec can require exact agreement for the median and
a tight tolerance for the mean.

### Limits of what was tested (conditions)

- R 4.4.2 with its default generator settings, on Apple silicon. An R older than 3.6.0, or
  `RNGkind(sample.kind = "Rounding")`, draws indices differently.
- n up to 100,000. For n ≥ 2^31 R switches to a different path; out of scope.
- Data that is exact in floating point (integers). Real data read from text must parse to the same doubles
  in both languages; see risk R2 in the spec.
- The `boot` package draws all indices in one call and assigns them to replicates in a different order than
  a `sample()` loop. Its result is expected to agree statistically, not digit for digit (untested).

## 8. Predictions, with the conditions they assume

| # | Prediction | Assumes |
|---|---|---|
| P1 | The unchanged R program reproduces the course listing exactly | R ≥ 3.6 produced the listing (same `sample` method); seed 9999 |
| P2 | The unchanged R program takes 3–15 s per run | 40,000 data-frame `rbind`s at roughly 0.1–0.3 ms each on an M1 Max; R start-up about 0.2 s |
| P3 | The Go spike reproduces all 14,000 reference index draws and both bootstrap standard errors (to 1e-12 relative) | R 4.4.2 defaults as probed; `sd` differs from Go at most in the last bits |
| P4 | Plain single-core Go does the study's resampling in under 0.2 s on the Mac, so **Go is at least 15× faster than the unchanged R program** | M1 Max about 4× the workspace CPU; P2 holds |
| P5 | Against **idiomatic R** the single-core gap shrinks to roughly 2–10× | R's `sample`, `sort` and `median` are compiled C; the remaining gap is interpreter overhead per resample, largest for small n |
| P6 | For runs this short, **whole-process** time is dominated by R's start-up, so the cost saving depends on job size | R start-up about 0.2 s against a few ms for Go |

### Predictions checked so far (2026-11-14)

| # | Outcome | Evidence |
|---|---|---|
| P1 | **Held.** Listing reproduced exactly. | §7 |
| P2 | **Held.** 5.1–5.2 s per run (predicted 3–15 s). | §7 |
| P3 | **Held, and stronger than predicted:** the median's SE is identical, not merely within 1e-12. | §7 |
| P4–P6 | Open until the official measurement on the Mac. | |

---

## Addendum A (2026-11-14, after review R1): the identity claim is narrower than §7 said

§7 concluded that Go gives R's bootstrap standard error of the median "digit for digit". The evidence
supports less than that sentence, and review R1 (DR-1, EC-1) was right to challenge it:

- What is **proved identical** is the sequence of index draws, and therefore each resample and **each
  resample median**.
- The **standard error** is a standard deviation of B medians: a sum. A simulation with extended-precision
  accumulators (what R uses on Intel) differs from plain doubles in the last bit in about one case in five
  (maximum relative difference 1.4e-16), for integer data as much as decimal data. The identity observed in
  §7 holds on Apple silicon, where R's `long double` is a plain double; it is not portable.
- The spec (§6, v2) therefore requires identity for medians and 1e-12 for standard errors.
- The reference data in §7 was integer-valued. Decimal data is covered by the official comparison on the
  12 data sets (spec AC-3), which is still to run.

## Addendum B (2026-11-14, after review R1): P4–P6 restated with checkable conditions

The original wording in §8 stays as written. Checked versions (review EC-3):

| # | Restated prediction | Checked by |
|---|---|---|
| P4 | On the M1 Max, single-core Go does the course study's resampling volume (100 × (25 + 100 + 225 + 400) × 100 resamples, median and mean) in **under 0.20 s of computation time**: at least 25× less than the unchanged R program's 5.1 s whole-process time. Same resampling volume, not the same program. | A Go benchmark in the official measurement |
| P5 | **Computation time** (`compute_ms`) of the final single-core Go against idiomatic base R (`sample()` loop), B = 1,000, is between **2× and 10×** in Go's favour for n = 1,000 and n = 10,000. | Generated tables |
| P6 | At n = 100, B = 1,000, R's start-up is **more than 50 %** of its whole-process time; at n = 100,000 it is **under 20 %**. | Generated tables |
