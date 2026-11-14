# Spec 001: Bootstrap standard error of the median, in R and in Go

| Field | Value |
|---|---|
| Status | **DRAFT v1, 2026-11-14**, for Gemini spec review (R1) and owner decisions Q1–Q10 |
| Inputs | `docs/assignment-and-rubric.md` (S1–S12, M1–M2, G1–G9) · `docs/constitution.md` · `docs/spikes/001-environment-and-baseline.md` |
| Deadline | 2026-11-15, 11:59 PM ET. Nothing is built before approval. |

## 1. Problem

A research consultancy runs computer-intensive statistics in R and pays for the cloud time. It asks an
independent data scientist two questions: **how easy is it to use Go in place of R (M1)**, and **how much
cloud cost would that save (M2)**. We answer for one method from Gelman and Vehtari's list, *bootstrapping
and simulation-based inference*, in the form the assignment suggests: a **distribution-free estimate of the
standard error of the median** by bootstrap resampling, for symmetric, positively skewed and negatively
skewed data.

The answer has three parts, and each must be reproducible by a command:

1. **The same analysis in R and in Go, on the same data, with comparable results.**
2. **Time and memory of both**, measured the same way, with start-up separated from computation.
3. **A recommendation with a cost figure** derived from the measurements and from a cloud provider's saved prices.

The spike changed what "comparable" can mean: Go can reproduce R's random draws exactly, so for the same data
and seed the two programs can be required to give **the same standard error of the median, digit for digit**
(spike §6–§7). The spike also showed the course's jump-start R program spends most of its 5 seconds on data
frames, so the comparison needs an idiomatic R baseline as well (spike §4).

## 2. Users and use

| User | Needs |
|---|---|
| The consultancy's managers | A recommendation they can act on: when Go pays off, when it does not, and a percentage with its assumptions. |
| An analyst at the consultancy | Run one command on a data file and get the median, its standard error, and the mean's as a check; trust that it matches R. |
| A Go developer (and Week 9) | Reuse the resampling engine with another statistic or another random source. |
| Grader | Build, run and test from the README; R and Go code and data in the repository; evidence for every claim. |

## 3. Definitions (so R and Go compute the same thing)

- **Data set**: n observations x₁…xₙ, n ≥ 2, read from a one-column CSV file.
- **Median**: for odd n the middle value; for even n **the mean of the two middle values** (R's `median`).
- **Bootstrap resample**: n values drawn from the data set with replacement.
- **Bootstrap standard error of a statistic** (the textbook definition, Efron and Tibshirani 1993): draw B
  resamples, compute the statistic on each, and take the **standard deviation of those B values, with
  denominator B − 1**.
- The course's R program reports a different quantity (the spread, across 100 fresh samples, of the average
  bootstrap median; spike §4). It is run unchanged and reported as the course's baseline; it is **not** the
  quantity our R and Go programs compute, and the README says so.
- **R stream**: resample indices drawn exactly as R 4.4.2 draws them after `set.seed(seed)` with default
  settings. **Block stream**: replicates split into fixed-size blocks, each block with its own generator
  derived from the seed and the block number, so blocks can run on different cores.

## 4. Functional requirements

| ID | Requirement |
|---|---|
| **FR-1 Data sets** | The repository holds generated data sets for three shapes (symmetric, positively skewed, negatively skewed) at the sizes in **Q5**, each a one-column CSV with a header, values written with **4 decimal places**. A committed, seeded R script generates them; a manifest records file, shape, n, generator call, seed, size and SHA-256. |
| **FR-2 R analysis** | An R script, run with `Rscript`, reads a data file and reports the sample median, the bootstrap SE of the median and the bootstrap SE of the mean, for given B and seed, using a `sample()` loop with preallocated vectors (idiomatic base R). A second R script does the same with the CRAN **boot** package. The course's `run-bootstrap-median.R` stays unchanged in `r/original/`. |
| **FR-3 Go analysis** | A Go program reads the same data file and reports the same quantities for the same B and seed. |
| **FR-4 Same output format** | R and Go write the same machine-readable record (tab-separated: data set, n, B, seed, stream, sample median, SE median, SE mean, sample mean), numbers with 17 significant digits, so a program can compare them. A human-readable summary is the default on the terminal. |
| **FR-5 R-compatible random draws** | With the R stream, the Go program draws the same resample indices as R for the same seed (**Q2**). |
| **FR-6 Parallel resampling** | With the block stream, the Go program can use several workers (`-workers`, default the number of CPUs). **The result for a given seed is identical for every worker count.** |
| **FR-7 Comparison** | A committed program compares the R and Go records for every data set by the rule in §6 and writes a pass/fail table to `results/`. The official measurement refuses to time anything until this passes. |
| **FR-8 Input validation** | Bad input is rejected before any computation with a message naming the problem: missing or unreadable file, non-numeric value (file, line), fewer than 2 observations, B < 2, workers < 1, unknown statistic or stream, workers > 1 with the R stream. NaN and infinite values are rejected. |
| **FR-9 Exit codes** | **0** success · **2** invalid input or usage · **1** any other failure. Tested on the built binary. |
| **FR-10 No-argument run** | With no arguments the program analyses a built-in copy of one committed data set and runs to completion (so a double-clicked `.app` works). |
| **FR-11 Logging and profiles** | Structured log lines on standard error (what was read, settings, time per phase); `-cpuprofile` and `-memprofile` write pprof files; `-quiet` silences the log. Results go to standard output. |
| **FR-12 Performance history** | Each performance step is kept as a benchmark pair (before, after) with allocation counts, so the README's account of what improved the program is generated from saved benchmark output (constitution 3.5). |
| **FR-13 Cost model** | Cloud prices for the provider in **Q8** are saved in a file with source URL, region, instance type, price and capture date. A program computes, from measured times and saved prices, the cost per job for R and Go and the percentage saved; the README's figures come from it. |
| **FR-14 Builds** | Scripts and README instructions create a macOS binary, a Windows `.exe` and a macOS `.app` (**Q9**). |

### Reuse requirements

| ID | Requirement | Shown by |
|---|---|---|
| **FR-R1** | The resampling engine is a Go package with no printing and no `os.Exit`. It takes data, **any statistic** (a function from a sample to a number), B, and a random source. | A test that bootstraps a statistic the package does not define (a trimmed mean) |
| **FR-R2** | The **random source is replaceable**: the R-compatible generator is one implementation of a small interface; a standard-library generator is another. | A test with each source; README example |
| **FR-R3** | The R-compatible generator is its own package with no dependency on the rest, usable by any Go program that must reproduce R's `runif` or `sample`. | Its tests use only R's saved reference files |
| **FR-R4** | Statistics (median, mean, standard deviation) are their own small package. | Used by engine and by `cmd` programs |

## 5. Non-functional requirements

| ID | Requirement |
|---|---|
| **NFR-1 Fair comparison (G1–G5)** | Same Mac, same data file, same B, same seed, batch execution (`Rscript`, built Go binary) from one shell script, both measured with `/usr/bin/time -l`. Neither side plots. |
| **NFR-2 Enough runs (G6, G7)** | Each timed condition is run the number of times in **Q10**; mean, median and spread are reported; the raw per-run log is committed. |
| **NFR-3 Three kinds of time (G8)** | For each language: **whole-process** time; **start-up** time (an empty R script, a Go binary that exits at once); **computation** time measured inside the program. Every ratio in the README says which kind it is. |
| **NFR-4 Memory** | Peak memory of the whole process from `/usr/bin/time -l` for both languages, reported with its spread (R's varies between identical runs, spike §7); Go allocations per operation from benchmarks. |
| **NFR-5 Reported separately** | Go on one core against R (like for like), and Go on all cores (what Go adds). The course's unchanged R program is a third, separate comparison. |
| **NFR-6 Freshness** | The measurement script refuses a stale or modified copy and records machine, CPU count, Go and R versions, R package versions and the commit. No user name, host name or home path. |
| **NFR-7 Generated numbers** | All README tables, headline figures and the saving percentage are generated from `results/` and the price file. |
| **NFR-8 Quality gates** | `gofmt`, `go vet`, `go test -race ./...` clean; tests need neither R nor the internet; critical tests mutation-checked. |
| **NFR-9 Dependencies** | Per **Q2**. `go.sum`, if any, is generated on the Mac. |
| **NFR-10 Privacy and provenance** | Constitution Article 8; the course's files are credited to their author. |

## 6. Comparability rule (written before measuring; constitution 1.1)

| Quantity | R (`sample()` loop) against Go (R stream) | Against the block stream or `boot` |
|---|---|---|
| n | equal | equal |
| Sample median | **identical** (same double) | identical |
| Sample mean | relative difference ≤ 1e-12 | ≤ 1e-12 |
| Bootstrap SE of the median | **identical** (same double) | relative difference ≤ **10 %** at B = 1,000 |
| Bootstrap SE of the mean | relative difference ≤ 1e-12 | ≤ 10 % at B = 1,000 |

Why these numbers. Identity for the median: the same indices select the same values, and a median is a
selection (or one exact average), so no rounding can differ; the spike observed identity. 1e-12 for means:
summation order differs; the spike observed 4.5e-16. 10 % between different random streams: two bootstrap
runs with different draws differ by Monte Carlo error, roughly SE ÷ √(2B) ≈ 2.2 % of the SE at B = 1,000 for
a near-normal statistic; 10 % is about four of those and still catches a wrong formula (a denominator of B
instead of B − 1 is invisible at any tolerance, so that is covered by exact tests instead).

A second, independent check: for the symmetric (normal) data sets the bootstrap SE of the **mean** must be
within 10 % of s ÷ √n, where s is the sample standard deviation.

## 7. Risks

| # | Risk | Handling |
|---|---|---|
| R1 | Exact R parity depends on R's version and settings (spike §7 limits) | The R scripts record `R.version.string` and `RNGkind()` in their output; the comparison reports them; the README states the conditions |
| R2 | R and Go may parse the same decimal text to different doubles | Values are written with 4 decimals (few significant digits parse exactly in both); the sample median and mean comparison in §6 detects any difference on every data set |
| R3 | Whole-process timing of sub-second jobs mostly measures start-up | NFR-3 separates the three kinds of time; sizes in Q5 include jobs long enough for computation to dominate |
| R4 | A parallel run that is fast but not reproducible | FR-6 is a tested property: same result for 1, 2 and all workers |
| R5 | Block streams seeded from neighbouring numbers could be correlated | Property to test: block-stream SEs agree with R-stream SEs within the §6 tolerance on all data sets; the README states the block stream is for speed, the R stream for audit |
| R6 | The cost percentage could be read as general | The README states it holds for this method, these sizes, this instance type, at the capture date |
| R7 | Time: hundreds of runs per condition | Q10 fixes a budget of about 15 minutes of machine time |

## 8. Acceptance criteria

| ID | Criterion | Checked by |
|---|---|---|
| AC-1 | Every data set in the manifest exists with the recorded hash and has n ≥ 100 | `scripts/check-inputs.sh` |
| AC-2 | The course's R program, unchanged, reproduces its listing | acceptance script, where R is present |
| AC-3 | R (`sample()` loop) and Go (R stream) agree by §6 on every data set | `results/comparison` table, all rows pass |
| AC-4 | Go reproduces R's saved reference draws and bootstraps (spike evidence) without R installed | `go test ./...` |
| AC-5 | Block-stream results are identical for 1, 2 and all workers, under the race detector | `go test -race` |
| AC-6 | Block stream and `boot` agree with the R stream by §6 | comparison table |
| AC-7 | Median, mean and standard deviation match independently known answers, including even n, odd n, ties, n = 2 | unit tests; mutation-checked |
| AC-8 | Each invalid input of FR-8 gives its message and exit code 2 on the built binary | acceptance script |
| AC-9 | No-argument run exits 0 and prints a result | acceptance script |
| AC-10 | Profiles and logs are produced on request; committed profiles exist | acceptance script |
| AC-11 | `results/` holds the raw per-run log with the run counts of Q10, and machine details without personal data | acceptance script |
| AC-12 | README tables, headline and saving percentage equal the generated ones | `scripts/readme.sh --check` |
| AC-13 | The price file has source URL, region, instance type, price and date for each price used | acceptance script |
| AC-14 | README contains: method and CRAN URLs; package search (R and Go); build process; experiences with testing, benchmarking, profiling, logging; what improved performance; R against Go results; recommendation with provider, prices and percentage; run, test and build instructions; GenAI Tools | acceptance script (headings) and review R2 |
| AC-15 | A statistic defined outside the package and a second random source both work with the engine | unit tests (FR-R1, FR-R2) |
| AC-16 | `gofmt`, `go vet`, `go test -race ./...` clean on a fresh clone | acceptance script |
| AC-17 | AI dialogs are in `docs/ai-dialogs/` with an index | acceptance script |

## 9. Traceability

### 9a. Assignment steps

| Step | Covered by |
|---|---|
| S1 select a method from Gelman and Vehtari | §1; README method section (AC-14) |
| S2 CRAN packages, URLs, analysis in R | FR-2 (base R and `boot`); AC-14; Q4 |
| S3 implement in Go | FR-3, FR-5, FR-6, FR-R1–R4 |
| S4 testing, benchmarking, profiling, logging; performance improvement | FR-11, FR-12, NFR-8; AC-7, AC-10 |
| S5 Mersenne Twister | FR-5, Q2; spike §5–§6 |
| S6 compare on the same data; comparable results; memory and time | FR-4, FR-7, §6, NFR-1–NFR-5; AC-3, AC-6, AC-11 |
| S7 complete data and code, n ≥ 100 | FR-1, FR-2, FR-3; AC-1 |
| S8 README: package search, build process, experiences | AC-14 |
| S9 recommendation, provider, prices, percentage | FR-13; AC-12, AC-13, AC-14 |
| S10 complete README, run, test and build instructions | FR-14; AC-14 |
| S11 GenAI Tools and logs | AC-14, AC-17 |
| S12 `.git` URL | Owner, at submission |
| M1 how easy | README: what had to be written in Go that R provides (generator, median, bootstrap), measured in lines and effort; AC-14 |
| M2 how much saved | FR-13 |
| G1–G9 benchmarking guidance | NFR-1 (G1, G3, G4, G5); hand-off step (G2); NFR-2 (G6, G7); NFR-3 (G8); FR-5 (G9) |

### 9b. Rubric

| Criterion | Covered by |
|---|---|
| [C] Coding | Constitution Art. 2; NFR-8; small packages (FR-R1–R4); readable R (constitution 2.9) |
| [T] Testing and metrics | AC-3–AC-7, AC-10, AC-11; FR-11, FR-12; NFR-2–NFR-4; the bootstrap itself is the Monte Carlo benchmark |
| [D] Design | FR-R1–R4, FR-6; same code path for 1 and many workers |
| [Doc] Documentation | AC-12, AC-14, AC-17; one task per commit |
| [A] Application | FR-8–FR-10, FR-14; AC-8, AC-9 |

## 10. Open decisions (owner)

| # | Decision | Options | Recommendation |
|---|---|---|---|
| **Q1** | Method | (a) bootstrap SE of the median, three shapes · (b) another method from the list | **(a)**: the assignment's own example, fully in R, and reusable in Week 9 |
| **Q2** | Random generator in Go | (a) our own R-compatible 32-bit Mersenne Twister, standard library only · (b) `seehuhn/mt19937` as the assignment suggests (64-bit variant, GPL-3.0; results comparable only statistically) · (c) both: (a) as default, (b) as an optional source | **(a)**, with a standard-library generator as the second source for FR-R2, and the README explaining why the suggested package cannot match R. (c) honours the suggestion literally but puts the program under GPL terms |
| **Q3** | R baselines | (a) course program only · (b) idiomatic base R + `boot`, with the course program reported separately | **(b)**: the headline saving rests on idiomatic R; comparing only with the course program would overstate it |
| **Q4** | Install the CRAN `bootstrap` package | (a) no: cite it with its URL, use base R and `boot` (already installed) · (b) install and add a third R script | **(a)** |
| **Q5** | Data sizes and shapes | (a) n = 100, 1,000, 10,000, 100,000 × three shapes (12 files, about 3 MB): Normal(100, 10); lognormal (positive skew); 200 − lognormal (negative skew) · (b) stop at 10,000 (9 files, 0.3 MB) | **(a)**: the firm's concern is large data sets, and at 100,000 computation dominates start-up |
| **Q6** | Replicates B | 1,000 (usual for a standard error) · 10,000 | **1,000** for comparison and timing |
| **Q7** | Default stream of the Go program | (a) R stream (auditable against R, one core) · (b) block stream (all cores) | **(a)** as default, (b) with `-stream block` |
| **Q8** | Cloud provider and price basis | (a) AWS EC2, on-demand, Linux, US East, a compute-optimised instance · (b) Google Compute Engine · (c) Azure | **(a)**, prices captured with source and date. Cost per job = measured time × price; saving = 1 − cost(Go) ÷ cost(R) on the same instance type, shown for one core and for all cores |
| **Q9** | Application builds | (a) scripts and instructions for binary, `.exe`, `.app`; binaries not committed · (b) also commit built binaries | **(a)**: the rubric now asks for instructions |
| **Q10** | Runs per timed condition | (a) **200** whole-process runs per condition for n ≤ 10,000 and **30** for n = 100,000, on the positively skewed data sets; 30 runs of the course program; about 15 minutes · (b) 1,000 runs per condition; over an hour | **(a)** |

## 11. Out of scope

- Reproducing R's `rnorm` in Go (data is generated once, by R, and read by both).
- Confidence intervals, BCa, other bootstrap variants; other statistics beyond median and mean in the command-line program (the engine accepts any).
- Python. The course guidance mentions it; the assignment compares R and Go.
- Running anything in the cloud. Costs are computed from local measurements and published prices, and the README says so.
