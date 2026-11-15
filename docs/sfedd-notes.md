# SFEDD notes from Week 8 (draft for Luis to confirm)

Working notes for the lessons log and for SFEDD v2.2. Each lesson cites where its evidence is. Nothing here
is final until Luis confirms it; the confirmed version goes to the course project's `SFEDD/` documents.

## Lessons (proposed IDs)

| ID | Category | What happened | Lesson | Evidence |
|---|---|---|---|---|
| L-W8-01 | Process | The kickoff arrived with last week's assignment PDF and due date; caught only because the lessons log named that assignment as finished. | At intake, compare the assignment's title and due date with the lessons log and today's date before anything else. | Conversation, first exchange |
| L-W8-02 | Hand-off | The first Mac step used the GitHub CLI, which the Mac does not have. | A Mac step may only use tools the Mac probe has confirmed. Probe first, then write the step. | `docs/spikes/env-mac.txt` (gh: MISSING) |
| L-W8-03 | Testing | Every mutant of the command-line program "survived": `go test` answered from its result cache. | Mutation runs use `-count=1`. A cached pass can only fake a survivor, never a catch. | `docs/mutation-checks.txt`, T09 notes |
| L-W8-04 | Testing | The 10 % Monte Carlo limit looked safe on paper (four standard deviations). Measured on 2,400 correct comparisons, 3 exceeded it. | Before fixing a tolerance, measure how far apart correct implementations actually are, and remember that a single reference run has its own error. | `docs/spikes/002-tolerance.md` |
| L-W8-05 | AI review | Two reviewer statements were wrong when tested: a two-pass standard deviation is accurate to 1e-14 at n = 100,000 (measured 1.9e-14); R draws no number for `sample.int(1)` (R's output shows it does). | A reviewer's number, or its reading of another system's source, is a claim. Test it against the system itself. | `docs/decisions.md` (T03); `docs/reviews/002-repository-review.md` |
| L-W8-06 | Process | The spike concluded "Go gives R's standard error digit for digit" from integer data on one platform. The evidence proved identical draws and medians, not identical sums. | State a spike finding as exactly what was shown, on which inputs and platform. | Spike 001 addendum A |
| L-W8-07 | Hand-off | Official run 1 started at a load average of 117 (a browser was open); the warning was printed and scrolled away. | Before an official run, look at the process list, not only the load average (which lags by minutes). The measurement script should stop, not warn, when the machine is busy, unless told to continue. | `results/run-1/machine.txt`; conversation |
| L-W8-08 | Technical | `/usr/bin/time` steps in 0.01 s and the Go program runs small jobs in about that. Found after measuring. | In the spike, compare the measuring tool's resolution with the smallest quantity to be measured. | `docs/decisions.md` (after run 1) |
| L-W8-09 | Technical | The profile on the official machine pointed at the random generator; the workspace profile had pointed at sorting. The fix needed a second 30-minute run. | Take a short profile on the official machine before the long measurement, not as part of it. | `results/run-1/cpu-profile.txt`; README step 4 |
| L-W8-10 | Process | Review R2 found small defects in the measured code after the final measurement. Fixing them would have invalidated the results, so they became "known limits". | Review the code before the final measurement and the conclusions after it: two reviews, not one. | `docs/reviews/002-repository-review.md` |
| L-W8-11 | Process | Twice a scripted edit failed part-way and the commit message still described the change. Both were caught on the next step. | After a scripted edit, check that the change is in the file before committing. | Commits "T10" and "T10 follow-up" |
| L-W8-12 | Testing | The goroutine-leak test failed once in about ten runs with nothing leaked: it counted goroutines the instant the engine returned. | A resource test may wait for release; it must not perform it. Run timing-sensitive tests many times before trusting them. | `docs/mutation-checks.txt`, note after R2 |
| L-W8-12b | Process | A commit changed a test file after the README's generated line counts were made, and the leak test's first fix was not run enough times; both surfaced only when acceptance was re-run the next morning. | Run the acceptance script as the last act before every hand-off, and run a fixed flaky test dozens of times before calling it fixed. | `docs/decisions.md` (2026-11-15) |
| L-W8-13 | Technical | Prediction P5 (Go 2× to 10× faster in computation at n = 10,000) failed in run 1. It assumed R's cost was interpreter overhead; at large n R runs compiled C. | When predicting against a baseline, ask where the baseline actually spends its time. A mature interpreter's inner loops are often compiled. | Spike 001 addendum C |
| L-W8-14 | Process | The README drew a conclusion from a cost column that was computed, not measured (eight jobs at once = one job's time ÷ 8). | Label every extrapolated number as such, in the table and in the sentence that uses it. | Review R2, claim 1 |
| L-W8-15 | Technical | A language's standard library can be the method: Go needed about 17 times the lines of the R script, mostly to supply what R has built in. | "How easy is the port" is answered by counting what the target language lacks, not by the difficulty of the algorithm. | README "Effort" |
| L-W8-16 | Technical | Another language's random numbers were reproduced exactly from its source, verified against its saved output; this turned "results are comparable" into a pass/fail check. | When two programs must agree, try for exact agreement on the random inputs first; statistical agreement is the fallback. | Spike 001 §6–§7; `rrand/` |

## Proposed changes for SFEDD v2.2

| # | Change | From |
|---|---|---|
| P1 | **Intake check**: assignment title and due date against the lessons log and today's date (phase 0 gate). | L-W8-01 |
| P2 | **Mac probe lists the tools each hand-off will use** (`git`, `gh`, `Rscript`, …); a step may use only confirmed tools (§5.1, §5.7). | L-W8-02 |
| P3 | **Mutation protocol**: `-count=1`; record survivors and how each was closed (§5.5). | L-W8-03 |
| P4 | **Tolerances are measured**: any statistical limit in a spec cites a spike that measured the spread of correct implementations (§3 phase 2). | L-W8-04 |
| P5 | **Measurement preflight** (§5.3): the script checks the process list and load, stops unless the machine is quiet or the owner overrides, warms up each fresh binary, and states the timer's resolution against the smallest expected time. | L-W8-07, L-W8-08 |
| P6 | **Official profile before the long run** (§3 phase 6): a two-minute profile on the official machine; act on it; then measure once. | L-W8-09 |
| P7 | **Split review R2**: R2a reviews the code before the final measurement; R2b reviews the README and conclusions after it. After the final measurement the measured code is frozen, and later code findings are listed as known limits (§3 phases 7–9). | L-W8-10 |
| P8 | **Verify scripted edits** before committing; a commit message describes what the diff shows (§5.6). | L-W8-11 |
| P9 | **Extrapolations are labelled** wherever they appear; generated tables mark computed-not-measured columns (§5.3). | L-W8-14 |

## Reusable assets from this week

- `rrand/`: R's `set.seed`, `runif`, `sample` in Go, with R reference files.
- `bootstrap/`, `stats/`: resampling engine and statistics (the assignment notes they may serve Week 9).
- `internal/quality`: doc-comment and function-length check as a Go test.
- `internal/report` + `cmd/tables`: named generated blocks in the README, with `--check`.
- `internal/compare` + `cmd/compare`: a comparability rule as code.
- `scripts/review-bundle.sh`: one-file review bundle with a file list.
- `scripts/acceptance.sh`: PASS / FAIL / SKIP per criterion, including "what was measured is what is submitted".
- `scripts/measure.sh`: interleaved conditions, fail-fast on unreadable timing, load record.
