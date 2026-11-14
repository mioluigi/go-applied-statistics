# Assignment Requirements and Rubric (summary)

> Summary of the Week 8 assignment, *Modern Applied Statistics with Go* (100 points, due
> **2026-11-15, 11:59 PM ET**; assignment page revised July 3, 2025), in our own words, so every tool and
> reviewer can check the spec against it. The Canvas page is the authoritative source.

## Scenario

A **research consultancy** has used **R** for years, relying on CRAN for modelling techniques, many of them
computer-intensive. Worried about **cloud computing costs** on large data sets, and having heard that Go is
fast and uses multicore processors well, it is considering moving much of its programming from R to Go. It
wants an **independent data scientist** to answer two questions:

- **M1** How easy is it to use Go in place of R?
- **M2** How much money will the firm save in cloud computing costs?

## Required steps (numbered for traceability)

| # | Step | Source wording, condensed |
|---|---|---|
| S1 | Read Gelman & Vehtari and **select one computer-intensive method** from its list, one that is **fully implemented in R** | list: counterfactual causal inference; bootstrapping and simulation-based inference; overparameterized models; regularization; Bayesian multilevel models; generic algorithms; adaptive decision analysis |
| S2 | From **CRAN**, identify the **R package(s)** for the method; **name the method and show its URLs**; **demonstrate an analysis in R** | "identify the R package(s) … demonstrate an analysis in R. Name the method and show its web addresses" |
| S3 | **Implement the method in Go**, with a third-party package or our own program (look at Awesome Go and pkg.go.dev) | "either by drawing on a third-party Go package or developing your own" |
| S4 | Employ Go **testing, benchmarking, profiling and logging**; say **what was done to improve the Go implementation's performance** | "What did you do to improve the performance…?" |
| S5 | *(Consider)* the **Mersenne Twister** in Go for random test data, the generator R and Python use | `github.com/seehuhn/mt19937` |
| S6 | **Compare R and the final Go implementation on the same input data**; **ensure results are comparable**; report **memory requirements and processing times** of both | also: read the Canvas page *Performance Comparisons of R, Python, and Go* |
| S7 | Repository holds **complete data and code for R and Go**; data sets of **at least 100 observations** | deliverables |
| S8 | README: **efforts in finding R and Go packages**; **process of building** the Go implementation; **experiences** with testing, benchmarking, profiling, logging | "describe your efforts … Review your process … Review your experiences" |
| S9 | README **recommendation**: when Go should replace R for this method; **select a cloud IaaS provider**; **note its VM costs**; **what percentage of cloud cost might be saved** | "Under what circumstances … What percentage…?" |
| S10 | README is **complete documentation**, with **instructions for running and testing**, and for **creating an executable** | deliverables; rubric [A] |
| S11 | **GenAI Tools** section + conversation logs | deliverables |
| S12 | Submit the cloneable **`.git` URL** in the Canvas comments | deliverables |

### The assignment's own suggested example (not mandatory)

Bootstrap sampling for a **robust, distribution-free estimate of the standard error of the median**, for
distributions of **various shapes (positively skewed, symmetric, negatively skewed)**: "Start by implementing
in R and then refactor into Go." It adds that bootstrap-median code **may be reusable in the Week 9
assignment**. Jump-start material named on the page (Canvas login required, **not yet in our hands**):

- `run-bootstrap-median.R`: R code for the standard error of the median from a normal distribution.
- `listing-from-100-run-bootstrap-median.txt`: its sample output.
- CRAN packages mentioned: **bootstrap** (Efron & Tibshirani 1993), **boot** (Davison & Hinkley 1997).

## Deliverables

- Cloneable GitHub URL ending in `.git`, in the Canvas comments.
- `README.md` as complete documentation (roles of programs and data, how to run and test).
- Complete **data and code for R and Go**; R and Go run on the **same data sets** (≥ 100 observations),
  comparing **run times and memory**.
- **GenAI Tools** section and conversation logs (plain text or links).

## Rubric: five criteria, 20 points each

Ratings per criterion: Outstanding 20 (all criteria) · Meets 15 (most) · Needs Improvement 10 · Partially
Acceptable 5 · Unacceptable 0.

| Criterion | What earns full marks |
|---|---|
| **[C] Coding rules, organization, aesthetics** | Go modules, idiomatic Go; meaningful names; no useless temporaries; clean blocks; one task per block; simple control flow; distribute the NOT; self-documenting code (McConnell 2004, 777–817). |
| **[T] Testing and software metrics** | Unit tests of critical components with synthetic data; logs and profiles; memory and processing measured for components and the whole program; **Monte Carlo performance benchmark if the requirements call for one**. We treat it as called for: the method is itself a Monte Carlo procedure and S4/S6 require benchmarks. |
| **[D] Design and development** | Clean, efficient design meeting every requirement; **general-purpose modules reusable in other projects**. |
| **[Doc] Documentation** | Effective Git/GitHub use; README with roles of programs and data, how to test and use. |
| **[A] Application** | **Instructions for creating** an executable (`.exe` or `.app`); *changed from Week 7, which asked for delivery of one*. Runs to completion; validates user input and explains mistakes; clean UX/UI. |

## Differences from Week 7 that matter

1. **Two languages.** R code is a deliverable and the baseline; it runs only on the Mac (the workspace has no R).
2. **Results must be comparable**, not only timed: the comparison is a correctness question before a speed one.
3. **A business answer is required** (S9): a cloud provider, its prices, and a percentage. Prices are
   third-party facts and need provenance (source URL, date).
4. No external system is called at run time; the outside world enters through **R's behaviour** (random
   numbers, quantile definition) and **cloud price lists**.

## Owner's emphasis (Luis, 2026-11-14)

Enhance the methodology, reuse what works, speed with quality first, learn to industrialize and keep improving.

## Who runs what

| Work | Where | Why |
|---|---|---|
| Documents, Go code, Go tests, scripts, review bundles | **Claude** (workspace) | Go 1.24.7 available; `mt19937` fetched directly from GitHub |
| Anything that runs **R** (baseline, R↔Go comparison, R timing and memory) | **Mac** | No R in the workspace, and none installable |
| `go mod tidy` to produce `go.sum` | **Mac** | Workspace cannot reach `proxy.golang.org` (L-W7-06) |
| Official measurements, profiles, builds (`.app`, `.exe`) | **Mac** | Official machine (SFEDD §5.3) |
| Downloading the Canvas jump-start files and the Canvas performance page | **Mac** (Luis, by hand) | Course login |
| Cloud price capture | **Claude** by web lookup, **confirmed by Luis** | Third-party fact with provenance |
| Gemini reviews R1 and R2; submission | **Luis** | Roles |
