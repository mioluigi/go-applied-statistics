# Spike 002: is the 10 % Monte Carlo limit safe for a correct implementation?

| Field | Value |
|---|---|
| Date | 2026-11-14, before any official comparison |
| Question | Spec §6 (v2) allowed standard errors from different random draws to differ by 10 % at B = 1,000. Would a **correct** implementation ever fail that? |
| Why now | The full comparison passed in the workspace with a largest difference of 6.5 %. That is closer to the limit than a rule with "about four standard deviations of margin" should come on 48 checks. |
| Program | `docs/spikes/002-tolerance/main.go`; output saved in `docs/spikes/002-tolerance/output.tsv` |
| Reproduce | `go run ./docs/spikes/002-tolerance data/normal-100.csv data/lognormal-100.csv data/mirrored-lognormal-100.csv data/normal-1000.csv data/lognormal-1000.csv data/mirrored-lognormal-1000.csv` |

## Method

Reference: the R stream's standard errors (seed 431). Candidates: the block stream with seeds 1 to 200.
Relative difference of each candidate from the reference, for the mean and the median.

## Result

| Data set | Statistic | SD of the relative difference | Largest | Share above 10 % |
|---|---|---|---|---|
| normal-100 | mean | 2.29 % | 9.14 % | 0 |
| normal-100 | median | 2.17 % | 6.51 % | 0 |
| lognormal-100 | mean | 2.48 % | 9.98 % | 0 |
| lognormal-100 | median | 2.38 % | 7.99 % | 0 |
| mirrored-lognormal-100 | mean | 2.11 % | 7.77 % | 0 |
| mirrored-lognormal-100 | median | 2.28 % | 7.50 % | 0 |
| normal-1000 | mean | 2.27 % | 9.51 % | 0 |
| normal-1000 | median | 2.25 % | **10.70 %** | **1.5 %** |
| lognormal-1000 | mean | 2.24 % | 7.82 % | 0 |
| lognormal-1000 | median | 2.35 % | 9.23 % | 0 |
| mirrored-lognormal-1000 | mean | 2.28 % | 6.50 % | 0 |
| mirrored-lognormal-1000 | median | 2.61 % | 9.17 % | 0 |

(Numbers copied from `output.tsv`; this is a spike record, not a README result.)

## Findings

1. The spread of one estimate is what the normal approximation predicted: 2.1 % to 2.6 % against 2.2 %.
   Review R1 expected the median to be much worse because of its discrete bootstrap distribution; on these
   data it is not (2.2 % to 2.6 %).
2. The largest differences are four times that, because the comparison is against **one reference run that
   has its own error**. Where the reference happens to sit far from the centre (normal-1000, median), every
   candidate inherits the offset.
3. So 10 % was not "four standard deviations": for an unlucky reference it is nearer two.

## Decision

Limit raised to **15 %** (spec §6, `docs/decisions.md`). The separate check against exact theory keeps 10 %.

## Prediction

P7: on the Mac, with the real `boot` package and all 12 data sets, every different-draws check passes at
15 %, and the largest difference is under 11 %. Assumes B = 1,000 and seed 431.
