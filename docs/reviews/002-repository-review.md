# Review 002: Gemini repository review (R2) and triage

| Field | Value |
|---|---|
| Reviewer | Gemini (web), independent reviewer |
| Reviewed | `docs/reviews/002-repository-review-bundle.txt`, commit `94a7609` (18 files) |
| Receipt check | **Passed**: the review opens by listing all 18 files and the commit |
| Score | **96 / 100** ([C] 19, [T] 19, [D] 19, [Doc] 19, [A] 20) |
| Verdict | **Submit after the listed changes** (documentation and price-file edits; no re-measurement) |
| Full text | `docs/reviews/002-repository-review-gemini.txt` (text extracted from the PDF Luis exported, 24 pages) |
| Triage | Claude, 2026-11-14 |

## The constraint on fixes

The final measurement was taken at commit `3a19148`. Results are valid only for the code they measured, and
`scripts/acceptance.sh` enforces that the program, the R code and the data are unchanged since. So findings
about the measured code are **recorded as known limits in the README**, not patched: patching would either
invalidate the results or need a third 30-minute measurement for defects that affect no reported number.
The reviewer recommends the same. Findings about the README, the report generator and the price file are
fixed.

## Claims tested before deciding

| Reviewer claim | Test | Result |
|---|---|---|
| RC-1 (a): for n = 1, R's `R_unif_index` returns at once without drawing, so `Index(1)` diverges from R by consuming a number | R's own output on the Mac, `testdata/r-reference/sample-indices-edge.tsv`: after 500 draws of `sample.int(1, …)` the next `runif(1)` is 0.25645866966806352, the same as after 500 draws of `sample.int(2, …)`, which draw one number each. Test `TestIndexMatchesSampleIntAtBitBoundaries` checks exactly this and passes. | **Wrong.** R 4.4.2 does consume one number per draw for n = 1, and the Go code matches it. The reviewer quoted a line of R source that this version of R does not behave as. |
| RC-1 (b): `Index(0)` never returns | Read the code: bit count 64, mask all ones, no candidate is below 0 | **Right.** Outside the documented precondition (n ≥ 1) and unreachable from the engine (n ≥ 2). |
| Claim 2: the Go start-up row's standard deviation comes from one slow first run, not only the timer step | `results/runs.tsv`: go-help run 1 took 0.26 s; the next slowest of 200 took 0.02 s | **Right.** My README sentence blamed only the timer. |
| Claim 1: the "8 jobs at once" costs are unmeasured | Read `internal/report/render_cost.go`: time of one process ÷ 8 | **Right.** My sentence "cost less at every size" claimed more than the table can show. |

## Findings and decisions

| ID | Sev. | Finding (short) | Decision | Reason and change |
|---|---|---|---|---|
| [C]-1, RC-5 | Low | `runtime.GC()` comment describes a heap profile, but an allocation profile is written; with `-cpuprofile`, `compute_ms` includes writing the profile | **Accept as a known limit** | True. In measured code; the official timings ran without profiling. README "Known limits". |
| RC-1 (a) | Med | `Index(1)` consumes a number where R does not | **Reject** | Contradicted by R's own output (above). |
| RC-1 (b) | Med | `Index(n)` hangs for n ≤ 0 | **Accept as a known limit** | Outside the documented precondition; README "Known limits". |
| [T]-1, RC-6 (1), friction 2 | Low | Final run started at load 6.25, above the script's warning level | **Accept (document)** | README "Results" now says so plainly and that the cause was not pinned down. A third run was not made: every R and large-job condition has a standard error of the mean at or below 0.6 %. |
| [T]-2, Claim 2 | Med | Start-up outlier | **Accept** | The start-up table gains a "Slowest run" column (generated); README explains both limits. |
| RC-6 (2) | Low | `awk` extractions lack `tail -n 1`; no warm-up of the new binary | **Accept as a known limit** | `measure.sh` already fails if time or memory cannot be read. Changing it now would leave a script that differs from the one that produced the results. Warm-up noted in "Known limits"; both go to the SFEDD lessons. |
| [D]-1, RC-2 | Med | Block stream allocates two objects per resample; a comment says "no allocation inside the loop" | **Accept as a known limit** | True (the comment refers to the buffer). The benchmark table already shows the allocation counts. |
| [D]-2, RC-3 | Low | Statistics share one mutable slice; a statistic that changes values would corrupt the next; idle goroutines when workers exceed replicates | **Accept as a known limit (the contract), reject the redesign** | Sharing is the documented design (it is what makes "one set of resamples" cheap). README states that a statistic must not change values. Copying the resample per statistic would undo performance step 2. |
| RC-4, [A] | Low | A header reading `nan`/`inf`, or a leading blank line, is rejected with a header message | **Accept as a known limit** | The message is clear and the R script has the same header rule. |
| [Doc]-1, Claim 1, cost 1 | **High** | Scenario B conclusion unsupported; memory of eight R processes ignored | **Accept** | README rewritten: the columns are an unmeasured best case. A generated line gives the memory eight processes would need against one 8-worker Go process. |
| [Doc]-2a, friction 1 | Low | Price file carries a "to be confirmed by Luis" note | **Accept** | Reworded as a statement of provenance. The other listings that showed the same prices on 2026-11-14: cloud-credits.org/prices/ec2, doit.com (c7g.medium), calculator.holori.com (c7g.2xlarge). |
| [Doc]-2b, S5, friction 3 | Med | The one-seed table for the suggested package, promised in spec Q2, is missing | **Accept** | New spike `docs/spikes/003-suggested-generator/` runs the package stand-alone; the README table is generated (R's saved values, ours computed live, the package's saved output), with a test that ours equal R's and the package's do not. |
| Claim 3 | Med | Ratios and dollars at n = 100 shown without the ⚠ | **Accept** | README notes which ratios and cost figures inherit the uncertainty. |
| Claim 4 | Low | 8 workers in timing but all CPUs in correctness and benchmarks; 48 checks against 60 + 60 | **Accept** | README explains both. |
| Cost 2 | Med | Laptop cores are not cloud vCPUs; cost per job assumes a busy instance | **Accept, without the reviewer's figure** | Both caveats added. The reviewer's "1.3× to 1.6× higher" is not adopted: it cites no source. |
| Cost 3, M2 | Med | Percentages sound large; connect them to dollars and break-even | **Accept** | Headline gains a generated "In dollars" line; the Recommendation uses it. |
| Friction 4 | Low | `Rscript r/bootstrap-se.R` with no arguments exits with a usage message | **Reject** | That is the right behaviour for a script that needs a data file; the README shows the command. |
| S12 | | Submission is a manual step | Noted | Luis submits. |

## Reviewer statements that are inaccurate or unverified

1. R returns without drawing for n = 1 (RC-1): contradicted by R 4.4.2's output.
2. "An Apple M1 Max performance core … is significantly faster … so cloud costs would be roughly 1.3× to 1.6× higher" (cost 2): no source; direction plausible, figure not adopted.
3. "The comment in `bootstrap.go` contradicts the allocations" ([D]-1): the comment is about the resample buffer, which is not allocated in the loop; the generator objects are. The finding stands, the "contradiction" is a reading.

## Accepted findings that fit the time left

All documentation, report and price-file findings are fixed in this commit. No finding required changing
measured code, and none changes a reported number.
