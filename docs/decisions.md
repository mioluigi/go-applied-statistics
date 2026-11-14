# Decisions log (newest first)

## 2026-11-14 · After official run 1: performance step 4, and the measurement is repeated
- **Evidence from run 1** (`results/run-1/`, commit `202d3cc`): all 264 comparison checks passed. The CPU profile showed the random generator taking 59 % of the time at n = 100,000 (the median 30 %). Go on one core with the R stream was only 1.4 times faster than R in computation at n = 100,000; spike prediction P5 (2× to 10× at n = 10,000) did not hold at that size.
- **Decision (Claude, for Luis to overturn):** add performance step 4, which the profile points to: refill the generator's state without remainder operations, and take the 16 random bits per draw directly from the 32-bit output (equal to R's `floor(unif_rand() * 65536)`, proved by test on every earlier version). Because the program changed, the official measurement must be repeated: results are only valid for the code they measured.
- **Second reason to repeat:** run 1 recorded a 1-minute load average of 117 at the start. The timings were stable (standard error of the mean about 1 % or less for R conditions, no drift, the course program at 5.08 s as in the spike), so the figure may not reflect real contention, but an official run should not carry that doubt. Run 1 is kept as evidence and as the "before" of step 4.
- **Measurement tool limit, recorded:** `/usr/bin/time` reports whole-process time in steps of 0.01 s, and the Go program runs small jobs in about that time. Whole-process ratios at n = 100 and n = 1,000 are therefore approximate; the README marks conditions whose standard error exceeds 2 % and points to the in-program `compute_ms` for precise comparison. A finer timer was considered and not adopted: the conclusions do not depend on it (the saving at small sizes is above 90 % either way), and the course guidance asks for one command-line tool for both languages.
- **Report changes:** cost Scenario B now also shows the block stream run one worker per job, so one-core and parallel Go are compared stream for stream; the README shows step 4 from the two official runs.
- **Documents updated:** `docs/mutation-checks.txt`, `results/run-1/README.md`, README draft.

## 2026-11-14 · T10: Monte Carlo limit in spec §6 raised from 10 % to 15 %, before the official comparison
- **Evidence:** `docs/spikes/002-tolerance.md`: 2,400 comparisons of correct block-stream runs (200 seeds × 6 data sets × 2 statistics) against the R-stream reference at B = 1,000. Standard deviation of the relative difference 2.1–2.6 %; largest 10.7 %; 3 of 2,400 above 10 % (all on one data set, whose reference run sits far from the centre).
- **Why it matters:** the official run makes 48 such comparisons (block stream and `boot`, 12 data sets, 2 statistics). At 10 % a correct implementation could fail, stop the measurement, and cost a round trip on deadline day. Review R1 (DR-3) had warned that a correct implementation could exceed 10 %.
- **Decision:** 15 % for standard errors from different draws. The check against exact theory stays at 10 %. The rule is still written before measuring: no official comparison has run.
- **Documents updated:** spec §6; `internal/compare` constants and tests.

## 2026-11-14 · T03: accuracy bound for the standard deviation relaxed from 1e-14 to 1e-12
- **Evidence:** On 100,000 values near 1e6, the two-pass standard deviation differs from exact rational arithmetic by 1.9e-14 relative (`stats` test, first run). Review R1 (DR-5) said a two-pass algorithm would meet 1e-14 "effortlessly"; for a sum of 100,000 squared deviations in double precision it does not (rounding grows with the number of terms). The mean does meet 1e-14.
- **Decision:** Spec NFR-8 and AC-7 now require 1e-14 for the mean and 1e-12 for the standard deviation, the tolerance the R comparison (§6) actually uses. Compensated summation was considered and not adopted: R does not use it, and the bootstrap takes standard deviations of only B = 1,000 values.
- **The test still catches** the one-pass variance shortcut (error about 4e-6).
- **Documents updated:** spec NFR-8, AC-7.

## 2026-11-14 · Approval: spec v2, plan, tasks; Q1–Q10 as recommended
- **Decision (Luis):** "approved, all recommended". Q1 bootstrap SE of the median; Q2 own R-compatible generator, standard library only; Q3 base R and `boot`, course program separate; Q4 `bootstrap` package cited, not installed; Q5 n = 100 to 100,000 × three shapes; Q6 B = 1,000; Q7 R stream by default; Q8 AWS EC2 on-demand, US East, arm64, 1-vCPU and 8-vCPU; Q9 build scripts and instructions; Q10 200 runs per condition (30 for the largest and the course program).
- **Q8 clarified before approval:** only published prices are used; nothing runs in the cloud and no account is needed. The README states that costs are computed from local timings and published prices.
- **Documents updated:** spec, plan, tasks marked APPROVED.

## 2026-11-14 · Review R1 triaged; spec v2
- **Decision:** 24 findings decided (19 accepted, 4 adjusted, 1 partly rejected); see `docs/reviews/001-spec-review.md`.
- **Reason:** Four High findings were real: a default-flags contradiction, an identity claim wider than its evidence, an undefined resample order, and a cost model in which the price cancelled.
- **Documents updated:** spec (v2, §12), constitution (3.4 and 4.2 amended, dated), spike (addenda A and B), assignment summary (receipt of course files).

## 2026-11-14 · Repository name
- **Decision:** `go-applied-statistics` (the assignment title is not a valid repository name).
- **Reason:** Matches the earlier `go-…` repositories. Taken by Claude at intake; Luis created the repository under that name.

## 2026-11-14 · The Canvas guidance page is not committed
- **Decision:** *Performance Comparisons of R, Python, and Go* is summarised in our own words (G1–G9) and not stored.
- **Reason:** Course material behind a login; the public repository holds only what the assignment hands out as starting code.
