# Spec 001: Bootstrap standard error of the median, in R and in Go

| Field | Value |
|---|---|
| Status | **DRAFT v2, 2026-11-14**: R1 findings applied (`docs/reviews/001-spec-review.md`); awaiting owner answers to Q1–Q10 and approval |
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
(spike §6–§7; narrowed after review R1 to *the same resample medians*, with the standard error equal to 1e-12, see §6). The spike also showed the course's jump-start R program spends most of its 5 seconds on data
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
- **One set of resamples per run** (review DR-4): after seeding, resample b (b = 1…B) is the next n
  consecutive index draws, and **every statistic is evaluated on that same resample**. The SE of the median
  and the SE of the mean in one run therefore come from the same B resamples, using B × n draws in total.
- The course's R program reports a different quantity (the spread, across 100 fresh samples, of the average
  bootstrap median; spike §4). It is run unchanged and reported as the course's baseline; it is **not** the
  quantity our R and Go programs compute, and the README says so.
- **R stream**: resample indices drawn exactly as R 4.4.2 draws them after `set.seed(seed)` with default
  settings. **Block stream**: every replicate has its own generator, derived only from the seed and the
  replicate number, so replicates can run on different cores in any order.
- **Computation time**: the time to evaluate the B replicates on data already in memory. Reading the file
  and writing the result are outside it (`read_ms` and `compute_ms` in the log).

## 4. Functional requirements

| ID | Requirement |
|---|---|
| **FR-1 Data sets** | The repository holds generated data sets for three shapes (symmetric, positively skewed, negatively skewed) at the sizes in **Q5**, each a one-column CSV with a header, values written with **4 decimal places**. A committed, seeded R script generates them; a manifest records file, shape, n, generator call, seed, size and SHA-256. |
| **FR-2 R analysis** | An R script, run with `Rscript`, reads a data file and reports the sample median, the bootstrap SE of the median and the bootstrap SE of the mean, for given B and seed, using a `sample()` loop with preallocated `numeric(B)` vectors and no data frames (idiomatic base R). A second R script does the same with the CRAN **boot** package. The course's `run-bootstrap-median.R` stays unchanged in `r/original/`. |
| **FR-3 Go analysis** | A Go program reads the same data file and reports the same quantities for the same B and seed. |
| **FR-4 Same output format** | R and Go write the same machine-readable record (tab-separated: data set, n, B, seed, stream, sample median, SE median, sample mean, SE mean), numbers with 17 significant digits, so a program can compare them; on request both also write the B resample medians, one per line. The default terminal output is a human-readable summary that states the **data set, n, B, seed, stream, sample median, bootstrap SE of the median, sample mean and bootstrap SE of the mean**. |
| **FR-5 R-compatible random draws** | With the R stream, the Go program draws the same resample indices as R for the same seed (**Q2**). |
| **FR-6 Parallel resampling** | There is one execution path: replicates are handed to `-workers` workers (at least 1), and replicate b's statistics are stored at position b, so the final reduction never depends on scheduling. With the **block stream** the result for a given seed is **bit-for-bit identical for every worker count**. The **R stream** is one sequential generator and runs with one worker. `-workers` defaults to **1 for the R stream** and to **the number of CPUs for the block stream**. The two streams give different draws; their results differ within the §6 tolerance and the README says so. |
| **FR-7 Comparison** | A committed program compares the R and Go records for every data set by the rule in §6 and writes a pass/fail table to `results/`. The official measurement refuses to time anything until this passes. |
| **FR-8 Input validation** | Bad input is rejected before any computation with a message naming the problem: missing or unreadable file, non-numeric value (file, line), fewer than 2 observations, B < 2, workers < 1, unknown stream, an explicit `-workers` above 1 with the R stream. NaN and infinite values are rejected. |
| **FR-9 Exit codes** | **0** success · **2** invalid input or usage · **1** any other failure. Tested on the built binary. |
| **FR-10 No-argument run** | With no arguments the program analyses a built-in copy of one committed data set and runs to completion (so a double-clicked `.app` works). |
| **FR-11 Logging and profiles** | Structured log lines on standard error (what was read, settings, time per phase); `-cpuprofile` and `-memprofile` write pprof files; `-quiet` silences the log. Results go to standard output. |
| **FR-12 Performance history** | Each performance step is kept as a benchmark pair (before, after) with allocation counts; a faster variant must return the same values as the plain one (constitution 1.2), so the README's account of what improved the program is generated from saved benchmark output (constitution 3.5). |
| **FR-13 Cost model** | Cloud prices for the provider in **Q8** are saved in a file with source URL, region, instance type, vCPUs, memory, price and capture date. A program computes from **mean** measured whole-process times and the saved prices: (1) **cost in USD per 1,000,000 jobs** for R and Go, where the price does not cancel; (2) the **percentage saved**; for two scenarios: **right-sized** (each program billed per vCPU it uses: one-core R against one-core Go) and **multi-vCPU instance** (R running one job per vCPU against Go using all vCPUs for one job at a time); (3) **jobs needed to recover each dollar of rewrite cost** (1 ÷ saving per job), so M1 and M2 connect without inventing a developer cost. Median-based figures are shown beside the mean-based ones. The README's figures come from this program. |
| **FR-14 Builds** | Scripts and README instructions create a macOS binary, a Windows `.exe` and a macOS `.app` (**Q9**); the acceptance script builds all three into a temporary folder and checks they exist. |
| **FR-15 Effort measure (M1)** | A generated table counts the lines of R and of Go needed for the same analysis, and lists what Go had to supply that R provides (generator, `sample`, median, standard deviation, bootstrap loop). |

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
| **NFR-2 Enough runs (G6, G7)** | Each timed condition is run the number of times in **Q10**. Tables report, per condition, the number of runs N, mean, median, standard deviation and the **relative standard error of the mean** (flagged when above 2 %). The raw per-run log is committed. The README states why the multi-second conditions use N = 30. |
| **NFR-3 Three kinds of time (G8)** | For each language: **whole-process** time; **start-up** time, three conditions (an empty R script; an R script that only loads `boot`; the Go binary exiting at once); **computation** time (`compute_ms`, §3) and **read** time (`read_ms`) logged inside both programs. Every ratio in the README says which kind it is. |
| **NFR-4 Memory** | From `/usr/bin/time -l`, **both** fields (maximum resident set size and peak memory footprint) are recorded for every run of both languages, **including the start-up runs**, so memory above each language's empty-process baseline can be reported, with its spread. Component level: saved `go test -bench -benchmem` output has rows (ns/op, B/op, allocs/op) for the **random generator**, the **median**, and the **bootstrap engine**. |
| **NFR-5 Reported separately** | Go on one core against R (like for like), and Go on all cores (what Go adds). The course's unchanged R program is a third, separate comparison. |
| **NFR-6 Freshness** | The measurement script refuses a stale or modified copy and records machine, CPU count, Go and R versions, R package versions and the commit. No user name, host name or home path. It records the **load average before and after** the run and warns when the machine is busy at the start (G2). |
| **NFR-7 Generated numbers** | All README tables, headline figures and the saving percentage are generated from `results/` and the price file. |
| **NFR-8 Quality gates** | `gofmt`, `go vet`, `go test -race ./...` clean; tests need neither R nor the internet; critical tests mutation-checked. A test requires a **doc comment on every exported identifier** and **no function longer than 60 lines**. Mean and standard deviation use a two-pass form and are accurate to 1e-14 relative on badly conditioned data (n = 100,000, values near 1e6). |
| **NFR-9 Dependencies** | Per **Q2**. `go.sum`, if any, is generated on the Mac. |
| **NFR-10 Privacy and provenance** | Constitution Article 8; the course's files are credited to their author. |

## 6. Comparability rule (written before measuring; constitution 1.1)

| Quantity | R (`sample()` loop) against Go (R stream) | Against the block stream or `boot` |
|---|---|---|
| n | equal | equal |
| Sample median | **identical** (same double) | identical |
| Sample mean | relative difference ≤ 1e-12 | ≤ 1e-12 |
| **Each of the B resample medians** | **identical** (same double), all B of them | not applicable (different draws) |
| Bootstrap SE of the median | relative difference ≤ **1e-12** | relative difference ≤ **10 %** at B = 1,000 |
| Bootstrap SE of the mean | relative difference ≤ 1e-12 | ≤ 10 % at B = 1,000 |

Why these numbers.

- **Identity for medians.** The same indices select the same values, and a median is a selection (or one
  exact average of two), so no rounding can differ. This is the strong claim: Go drew R's resamples.
- **1e-12 for anything summed** (means, and standard errors, which are standard deviations of B values).
  The sum can differ in the last bit with the width of the accumulator, which depends on the platform; the
  spike saw identity on Apple silicon, and a simulation with extended precision saw at most 1.4e-16 (review
  R1, DR-1). 1e-12 leaves a wide margin and still fails on any real difference.
- **10 % between different random streams.** Two correct bootstraps with different draws differ by Monte
  Carlo error: about SE ÷ √(2B) ≈ 2.2 % at B = 1,000 for a near-normal statistic, and more for the median,
  whose bootstrap distribution sits on a few order statistics and can have heavy tails, most at n = 100.
  10 % allows for that. The rule is **verified on the committed seeds before anything is timed** (FR-7); if a
  correct implementation fails it, the rule is examined first (SFEDD principle 6) and any change is recorded.
- A wrong denominator (B instead of B − 1) hides inside any such tolerance, so it is covered by exact tests.

A second, independent check, on **all 12 data sets** (it holds for any shape): the bootstrap SE of the
**mean** must be within 10 % of √((n − 1) ÷ n) × s ÷ √n, where s is the sample standard deviation.

## 7. Risks

| # | Risk | Handling |
|---|---|---|
| R1 | Exact R parity depends on R's version and settings (spike §7 limits) | The R scripts record `R.version.string` and `RNGkind()` in their output; the comparison reports them; the README states the conditions |
| R2 | R and Go may parse the same decimal text to different doubles | The values are not exact in binary; what matters is that both parsers return the **same** double. Short decimals (4 places) make that likely; it is **checked, not assumed**: the sample median, the sample mean and all B resample medians are compared on every data set (§6) |
| R3 | Whole-process timing of sub-second jobs mostly measures start-up | NFR-3 separates the three kinds of time; sizes in Q5 include jobs long enough for computation to dominate |
| R4 | A parallel run that is fast but not reproducible | FR-6 is a tested property: same result for 1, 2 and all workers |
| R5 | Per-replicate generators seeded from neighbouring numbers could be correlated | Property to test: block-stream SEs agree with R-stream SEs within the §6 tolerance on all data sets; the README states the block stream is for speed, the R stream for audit |
| R6 | The cost percentage could be read as general | The README states it holds for this method, these sizes, this instance type, at the capture date |
| R7 | Time: hundreds of runs per condition | Q10 fixes a budget of about 20 minutes of machine time; deadline-mode cuts are ordered in `plan.md` §9 |
| R8 | Exact agreement of summed quantities is platform-dependent | §6 requires identity only for medians; summed quantities use 1e-12 |

## 8. Acceptance criteria

| ID | Criterion | Checked by |
|---|---|---|
| AC-1 | Every data set in the manifest exists with the recorded hash and has n ≥ 100 | `scripts/check-inputs.sh` |
| AC-2 | The course's R program, unchanged, prints output that contains the course listing as **one contiguous block** (the program also prints four console lines the listing never had) | acceptance script, where R is present |
| AC-3 | R (`sample()` loop) and Go (R stream) agree by §6 on every data set, including all B resample medians | `results/comparison` table, all rows pass |
| AC-4 | Go reproduces R's saved reference draws and bootstraps (spike evidence) without R installed | `go test ./...` |
| AC-5 | Block-stream results are bit-for-bit identical for 1, 2, 4 and 8 workers, repeated 10 times, under the race detector | `go test -race -count=10` |
| AC-6 | Block stream and `boot` agree with the R stream by §6; the bootstrap SE of the mean passes the √((n−1)/n)·s/√n check on all 12 data sets | comparison table |
| AC-7 | Median, mean and standard deviation match independently known answers, including even n, odd n, ties, n = 2, and badly conditioned data to 1e-14 | unit tests; mutation-checked |
| AC-8 | Each invalid input of FR-8 gives its message and exit code 2 on the built binary | acceptance script |
| AC-9 | No-argument run exits 0 and its output names every field listed in FR-4 | acceptance script |
| AC-10 | Profiles and logs are produced on request; committed profiles exist; saved benchmark output has `-benchmem` rows for the generator, the median and the engine, and the before/after pairs of FR-12 | acceptance script |
| AC-11 | `results/` holds the raw per-run log with the run counts of Q10 (all three shapes at n = 10,000; three start-up conditions; both memory fields), the load averages, and machine details without personal data | acceptance script |
| AC-12 | README tables, headline and saving percentage equal the generated ones | `scripts/readme.sh --check` |
| AC-13 | The price file has source URL, region, instance type, price and date for each price used | acceptance script |
| AC-14 | README contains the exact CRAN URLs of `boot` and `bootstrap`, pkg.go.dev, Awesome Go, the Gelman and Vehtari citation, the generated line-count table (FR-15), and sections for: method and CRAN URLs; package search (R and Go); build process; experiences with testing, benchmarking, profiling, logging; what improved performance; R against Go results; recommendation with provider, prices and percentage; run, test and build instructions; GenAI Tools | acceptance script (headings and required strings) and review R2 |
| AC-15 | A statistic defined outside the package and a second random source both work with the engine | unit tests (FR-R1, FR-R2) |
| AC-16 | `gofmt`, `go vet`, `go test -race ./...` clean on a fresh clone, including the doc-comment and function-length test; all three builds succeed | acceptance script |
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
| M1 how easy | FR-15 (generated line counts; what Go had to supply); README experience sections; AC-14 |
| M2 how much saved | FR-13 |
| G1–G9 benchmarking guidance | NFR-1 (G1, G3, G4, G5); hand-off step and NFR-6 load average (G2); NFR-2 (G6, G7); NFR-3 (G8); FR-5 (G9) |

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
| **Q2** | Random generator in Go | (a) our own R-compatible 32-bit Mersenne Twister, standard library only · (b) `seehuhn/mt19937` as the assignment suggests (64-bit variant, GPL-3.0; results comparable only statistically) · (c) both: (a) as default, (b) as an optional source | **(a)**, with a standard-library generator as the second source for FR-R2, and the README explaining, with a small table for one seed, why the suggested package cannot match R (*R1 agrees*). (c) honours the suggestion literally but puts the program under GPL terms |
| **Q3** | R baselines | (a) course program only · (b) idiomatic base R + `boot`, with the course program reported separately | **(b)**: the headline saving rests on idiomatic R; comparing only with the course program would overstate it |
| **Q4** | Install the CRAN `bootstrap` package | (a) no: cite it with its URL, use base R and `boot` (already installed) · (b) install and add a third R script | **(a)** |
| **Q5** | Data sizes and shapes | (a) n = 100, 1,000, 10,000, 100,000 × three shapes (12 files, about 3 MB): Normal(100, 10); lognormal (positive skew); 200 − lognormal (negative skew) · (b) stop at 10,000 (9 files, 0.3 MB) | **(a)**: the firm's concern is large data sets, and at 100,000 computation dominates start-up |
| **Q6** | Replicates B | 1,000 (usual for a standard error) · 10,000 | **1,000** for comparison and timing |
| **Q7** | Default stream of the Go program | (a) R stream (auditable against R, one core) · (b) block stream (all cores) | **(a)** as default, (b) with `-stream block`; workers default per FR-6 (*R1 agrees with that fix*) |
| **Q8** | Cloud provider and price basis | (a) AWS EC2, on-demand, Linux, US East, compute-optimised **arm64 (Graviton)** instances, matching the Mac's architecture: a 1-vCPU and an 8-vCPU size · (b) Google Compute Engine · (c) Azure | **(a)**, prices captured with source and date; cost model as in FR-13 (*R1: expanded so the price does not cancel*) |
| **Q9** | Application builds | (a) scripts and instructions for binary, `.exe`, `.app`; binaries not committed · (b) also commit built binaries | **(a)**: the rubric now asks for instructions |
| **Q10** | Runs per timed condition | (a) **200** whole-process runs per condition for n ≤ 10,000 and **30** for n = 100,000, on the positively skewed data sets, **plus all three shapes at n = 10,000**; 200 runs of each start-up condition; 30 runs of the course program; about 20 minutes · (b) 1,000 runs per condition; over an hour | **(a)** |

## 11. Out of scope

- Reproducing R's `rnorm` in Go (data is generated once, by R, and read by both).
- Confidence intervals, BCa, other bootstrap variants; other statistics beyond median and mean in the command-line program (the engine accepts any).
- Python. The course guidance mentions it; the assignment compares R and Go.
- A third, idiomatic R version of the course's *study* (review DR-2, rejected part). The course program is compared with Go at the same resampling volume and labelled as such.
- Running anything in the cloud. Costs are computed from local measurements and published prices, and the README says so.

## 12. Changes from v1 (review R1)

All 24 findings were decided in `docs/reviews/001-spec-review.md`: 19 accepted, 4 adjusted, 1 partly
rejected. The four High findings: default flags contradiction fixed (FR-6, FR-8); comparability rule
narrowed to identical *medians* and 1e-12 for standard errors (§6); one set of resamples per run defined
(§3); component benchmarks required (NFR-4) and the cost model expanded so the price matters (FR-13).
