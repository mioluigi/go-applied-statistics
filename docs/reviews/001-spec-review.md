# Review 001: Gemini spec review (R1) and triage

| Field | Value |
|---|---|
| Reviewer | Gemini (web), independent reviewer |
| Reviewed | `docs/reviews/001-spec-review-bundle.txt`, commit `3fe1156` (4 files) |
| Receipt check | **Passed**: the review opens by listing all four files and the commit |
| Verdict | **Approve with changes** (4 High findings) |
| Full text | `docs/reviews/001-spec-review-gemini.txt` (text extracted from the PDF Luis exported, 20 pages) |
| Triage | Claude, 2026-11-14; decisions are Luis's to overturn at approval |

## Claims tested before deciding (SFEDD principle 5)

Simulation in the workspace (Python, numpy 80-bit `longdouble` standing in for R's accumulators on Intel):

| Reviewer claim | Test | Result |
|---|---|---|
| DR-1/EC-1: the SE of the median cannot be guaranteed identical on decimal data, because `sd()` sums B values | 200 trials of 1,000 values: R-style two-pass standard deviation with extended-precision accumulators against plain doubles | **Half right.** The result differs in the last bit in 39 of 200 trials (max relative difference 1.4e-16). But it differs just as often for integer and half-integer data (36 of 200), so the reviewer's explanation (exact sums for integers) is not the cause; the cause is accumulator precision, which is platform-dependent. On Apple silicon `long double` equals `double`, which is why the spike saw identity. Identity is real on the Mac but not portable. |
| DR-1: the difference can reach "1–10 ULPs" | same | Not observed: at most 1 ULP (1.4e-16). The 1e-12 tolerance has five orders of margin. |
| DR-5: a naive single-pass mean over 100,000 values can violate 1e-12 | n = 100,000, Normal(1e6, 1), 4 decimals, against exact rational arithmetic | **Not reproduced** for the naive mean (error 0). **Confirmed for the textbook one-pass variance** (Σx² − n·mean²): relative error 4e-6. A test with badly conditioned data is worth having because it catches that formula. |
| DR-5: R's `R_strtod` "guarantees correct rounding" | cannot be tested without R | **Not established.** Recorded as an unverified reviewer statement. Our handling does not depend on it: every data set's sample median, sample mean and all B resample medians are compared between R and Go. |

## Findings and decisions

| ID | Sev. | Finding (short) | Decision | Reason and change |
|---|---|---|---|---|
| CG-1 | High | No requirement for per-component benchmarks with allocations | **Accept** | Rubric [T] says components *and* whole program. NFR-4 and AC-10 now require saved `-benchmem` rows for the generator, the median, and the engine. |
| CG-2 | Med | Timing on one shape only | **Accept** | Q10 adds all three shapes at n = 10,000. |
| CG-3 | Med | Start-up measurement omits `library(boot)` | **Accept** | NFR-3 adds a third start-up condition: R loading `boot`. |
| CG-4 | Low | 30 runs is below "hundreds" | **Adjust** | Tables report N and the relative standard error of the mean time per condition, flagged when above 2 %; README states why N = 30 for multi-second conditions. A hard 2 % gate is not adopted: it could force an unbounded re-run on deadline day. |
| CG-5 | Low | AC-14 checks headings only | **Accept** | Acceptance checks the CRAN URLs, pkg.go.dev, Awesome Go, the Gelman and Vehtari citation, and a generated line-count table for M1. |
| CON-1 | High | Default stream R + default workers = CPUs is rejected by FR-8; breaks the no-argument run | **Accept** | A real defect in the draft. `-workers` defaults to 1 for the R stream and to the CPU count for the block stream; an explicit `-workers` above 1 with the R stream is exit code 2. |
| CON-2 | Med | Constitution 4.2 and 1.4 against two streams | **Accept** | Constitution 4.2 amended: one execution path; the stream is a setting. FR-6 states both properties. The difference between streams is reported (constitution 1.2). |
| CON-3 | Med | Medians (constitution) against averages (G6); cost formula does not say which | **Accept** | Constitution 3.4 amended: mean and median both reported. The **mean** drives cost (a bill is a sum of run times); the median is shown beside it. |
| CON-4 | Low | Stale "not yet in our hands" statements | **Accept** | Both documents updated with dated notes of receipt (commit `80060ad`; guidance page read 2026-11-14). |
| UT-1 | Low | Closing other programs leaves no evidence | **Adjust** | `measure.sh` records the load average before and after, and warns above a threshold. It warns rather than fails: the threshold is a judgement, and the record lets a reader judge. |
| UT-2 | Med | Default terminal output unspecified | **Accept** | FR-4 lists the fields; AC-9 checks them. |
| UT-3 | Low | Doc comments and function size unchecked | **Accept** | A Go test (standard library `go/ast`) requires a doc comment on every exported identifier and bounds function length at 60 lines. Names and "one task per function" stay a matter for review R2. |
| DR-1 | High | SE of the median: identity will fail on decimal data | **Accept (half right, see tests)** | §6: sample median and **each resample median identical**; SE of the median ≤ 1e-12. |
| DR-2 | Med | Course program does different work; P4 unverified | **Adjust** | Primary comparison is always FR-2 against FR-3 at identical (n, B). P4 is verified by a Go benchmark of the course study's resampling volume, reported as "same resampling volume, not the same program". **Rejected part:** a third, idiomatic R version of the course study; it adds a deliverable for a secondary comparison. |
| DR-3 | Med | Tolerance justification; s/√n holds for any shape | **Accept** | §6: the mean's bootstrap SE is checked against √((n−1)/n)·s/√n on all 12 data sets; the 10 % wording acknowledges the median's discrete bootstrap distribution, and the rule is verified on the committed seeds before measuring. |
| DR-4 | High | Unspecified whether both statistics use the same B resamples | **Accept** | §3 and FR-2/FR-3: one set of B resamples per run; resample b is n consecutive index draws; every statistic is evaluated on each resample. The engine takes a list of statistics. |
| DR-5 | Med | R2 wording; summation at n = 100,000 | **Accept, with a correction** | R2 reworded (my "parse exactly" was wrong: the values are not exact in binary; what matters is that both parsers give the same double). Mean and standard deviation use R's two-pass form, tested at 1e-14 on badly conditioned data. The reviewer's `R_strtod` statement is not relied on. |
| DR-6 | Med | Block size and seeding undefined; completion order could change the result | **Accept** | FR-6: each replicate has its own generator derived from (seed, replicate number); replicate b's statistic is stored at position b, so the reduction order never depends on scheduling. Tested for 1, 2, 4, 8 workers, repeated, under the race detector. |
| DR-7 | Med | "Computation time" boundary; algorithmic fairness | **Accept** | NFR-3: both programs log `read_ms` and `compute_ms`; compute covers only the B replicates on in-memory data. Buffer reuse and selection instead of sorting are the planned FR-12 steps. |
| DR-8 | Med | Which memory field; R's baseline | **Accept** | NFR-4: both `/usr/bin/time -l` fields recorded, for start-up runs too, so memory above the empty-process baseline can be shown. |
| DR-9 | High | Price cancels in the percentage; idle-core fallacy; break-even | **Accept, one part adjusted** | FR-13: cost per 1,000,000 jobs in USD; a per-vCPU (right-sized) scenario and a multi-vCPU scenario; arm64 (Graviton) instances to match the Mac. **Adjusted:** break-even is reported as *jobs per dollar of rewrite cost*, not as a dollar figure, because a developer cost would be an invented number. |
| EC-1 | High | Spike over-generalises from integer data | **Accept** | Spike addendum A (dated); conclusion narrowed. |
| EC-2 | Med | Listing diff is not zero because of four console lines | **Accept** | AC-2: the listing must appear as one contiguous block inside the output. |
| EC-3 | Med | P4–P6 lack checkable conditions | **Accept** | Spike addendum B restates P4–P6 with conditions; the originals stay as written. |
| Q1–Q10 | | Agrees with every recommendation; Q7 with the CON-1 fix; Q8 with Graviton and cost per job; Q10 with a shape check | **Accept** | Reflected in spec v2 §10. For Q2 the reviewer suggests a README table: suggested package against R against ours on one seed. Adopted as a spike-style test, without importing the GPL package into the product (plan §7). |
| Scope cuts | | Order: timing matrix breadth; number of optimisation steps; `boot` timing sizes; n = 100,000 for one shape only | **Accept as the deadline-mode order** | Recorded in `plan.md` §9. |

## Reviewer statements that are inaccurate or unverified

1. "Both R's `R_strtod` and Go's `strconv.ParseFloat` guarantee correct rounding" (DR-5): unverified for R; not relied on.
2. Identity in the spike "occurred only because the spike used small integers" (DR-1, EC-1): the simulation shows integer data differs as often as decimal data under extended precision; the cause is accumulator width.
3. "Can differ by 1–10 ULPs (~1e-15)" (DR-1): at most 1 ULP observed.
4. A naive single-pass mean reaching 1e-11 at n = 100,000 (DR-5): not reproduced; only the one-pass variance formula fails.

None of these changes a decision; each made the spec more careful.
