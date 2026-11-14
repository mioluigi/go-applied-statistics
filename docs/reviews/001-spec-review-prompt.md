# Prompt for Gemini: spec review 001 (SFEDD §7.1)

**Upload one file:** `docs/reviews/001-spec-review-bundle.txt`. It contains four files, each after a
`===== FILE: path =====` header: `docs/assignment-and-rubric.md`, `docs/constitution.md`,
`docs/spikes/001-environment-and-baseline.md`, `specs/001-bootstrap-median/spec.md`.

Then paste everything below the line.

---

Before anything else, list every file you received inside the bundle, by name, and the commit named on its
first line. The bundle must contain these four: `docs/assignment-and-rubric.md`, `docs/constitution.md`,
`docs/spikes/001-environment-and-baseline.md`, `specs/001-bootstrap-median/spec.md`. If any is missing, stop
and say which.

You are two reviewers at once: a **strict grader** applying the rubric in the bundle, and a **senior
engineer who knows both R and Go and has done statistical computing and benchmarking in production**. Review
the draft spec (with its assignment summary, constitution and spike) for a project that implements the
bootstrap standard error of the median in R and in Go, shows the results agree, measures time and memory,
and recommends when Go should replace R, with a cloud-cost saving. The deadline is tomorrow night.

Report, in this order:

1. **Coverage gaps**: any assignment step (S1–S12), management question (M1–M2), benchmarking guidance item
   (G1–G9) or rubric criterion not covered, or covered only weakly.
2. **Contradictions** between spec, constitution, spike and assignment summary.
3. **Untestable requirements**: anything that cannot be checked by a test, a script or a reading of a file.
4. **Domain risks**, statistical and engineering: the bootstrap definition used and how it differs from the
   course's R program; the comparability rule in spec §6 and whether its tolerances are justified; exact
   reproduction of R's random draws and its limits; floating-point parsing and summation; parallel random
   streams; fairness of the R-versus-Go timing (start-up, package loading, idiomatic R, single core versus
   many); memory measurement; and the soundness of turning local timings into a cloud-cost percentage.
5. **Evidence check**: is each spike finding supported by the evidence it cites? Is any prediction (P1–P6)
   unjustified, or missing the conditions it assumes?
6. **A view on each open decision Q1–Q10**: agree with the recommendation or argue for another option.
   For Q2 in particular, weigh the assignment's suggestion of `seehuhn/mt19937` against the spike's finding.
7. **Scope cuts** if the deadline is at risk, in the order you would cut them.

For each finding give: **severity** (High / Medium / Low), **where** (file and section; quote the text),
**problem**, **suggested change**. Express worries as **properties a test can check**, not as a required
mechanism. For every claim you make about the documents, quote the file and section you rely on.

Finish with a verdict: *approve*, *approve with changes*, or *rework*.
