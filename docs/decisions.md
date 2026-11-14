# Decisions log (newest first)

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
