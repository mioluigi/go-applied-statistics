# Prompt for Gemini: repository review 002 (SFEDD §7.2)

**Upload one file:** `docs/reviews/002-repository-review-bundle.txt`. It contains 18 files, each after a
`===== FILE: path =====` header.

Then paste everything below the line.

---

Before anything else, list every file you received inside the bundle, by name, and the commit named on its
first line. The bundle must contain these 18 files: `docs/assignment-and-rubric.md`, `README.md`,
`results/tables.md`, `results/machine.txt`, `prices/aws-ec2.tsv`, `specs/001-bootstrap-median/spec.md`,
`docs/decisions.md`, `docs/mutation-checks.txt`, `main.go`, `rrand/rrand.go`, `stats/stats.go`,
`bootstrap/bootstrap.go`, `bootstrap/stream.go`, `dataset/dataset.go`, `r/bootstrap-se.R`,
`r/bootstrap-se-boot.R`, `scripts/measure.sh`, `scripts/acceptance.sh`. If any is missing, stop and say which.

You are two reviewers at once: a **strict grader** applying the rubric in `docs/assignment-and-rubric.md`,
and a **senior engineer who knows both R and Go and has done statistical computing and benchmarking in
production**. This is the finished repository for an assignment that implements the bootstrap standard
error of the median in R and in Go, shows the results agree, measures time and memory, and recommends when
Go should replace R, with a cloud-cost saving. It is due tomorrow night. You reviewed its specification
earlier; judge the result on what is in the bundle, not on that memory.

Report, in this order:

1. **Score each rubric criterion** ([C], [T], [D], [Doc], [A]; 20 points each) with the evidence for the
   score and what, specifically, keeps it from 20.
2. **Audit the assignment steps S1–S12, the management questions M1–M2 and the benchmarking guidance
   G1–G9**: for each, met, partly met or not met, and where in the bundle.
3. **Review the risky code**: the R-compatible generator (`rrand/rrand.go`), the selection median
   (`stats/stats.go`), the engine and its streams (`bootstrap/`), input handling (`dataset/dataset.go`,
   `main.go`), and the measurement script (`scripts/measure.sh`). Look for defects, not style.
4. **Check every conclusion in the README against `results/tables.md`**: quote each claim that the tables do
   not support, or that is stronger than the numbers allow. Pay particular attention to the Recommendation,
   the cost section and its two scenarios, and the statement about the timer's 0.01-second step.
5. **Is the cost model sound?** Consider what it assumes about running eight jobs at once, the use of laptop
   timings with cloud prices, and whether the "percentage saved" answers the consultancy's question.
6. **Grader friction**: anything that would slow down or confuse someone who clones the repository and
   follows the README.
7. **The three changes that would raise the score most and take under 20 minutes each.**

For every claim you make, **quote the file and section (or function) it relies on**. If you cannot find the
evidence in the bundle, say so instead of assuming. Express worries as properties a test or a reading of a
file can check.

Finish with a total score out of 100 and a verdict: *submit as is*, *submit after the listed changes*, or
*rework*.
