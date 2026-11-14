# Spike 001: environment, R baseline and R↔Go comparability

| Field | Value |
|---|---|
| Status | **Phase 0 part only** (environment probe, workspace side). Mac side and Phase 1 sections pending. |
| Date | 2026-11-14 |

## 1. Environment probe (SFEDD §5.7)

### 1.1 Claude's workspace (measured 2026-11-14)

| Check | Result | Consequence |
|---|---|---|
| Go toolchain | go 1.24.7 linux/amd64 | Go code and tests can be developed here |
| R | **absent**; `apt-get install r-base-core` has no candidate | Everything that runs R is a **Mac** task |
| Python | 3.13 with numpy 2.5.3, scipy 1.18.1 | An independent source of known answers for tests |
| `proxy.golang.org` | **unreachable** (name does not resolve) | `go.sum` is generated on the Mac (L-W7-06) |
| `github.com/seehuhn/mt19937` | v1.0.0 fetched with `GOPROXY=direct` | The suggested generator can be used and tested here |
| GitHub read | works (`git ls-remote`) | |
| GitHub push | CLI token invalid; push works only for repositories attached to the session | The repository must exist and be attached before the first push |
| Canvas (jump-start R code, performance page) | course login required | Luis downloads them |

### 1.2 Luis's Mac

Pending: `scripts/env-probe.sh` writes `docs/spikes/env-mac.txt`.

## 2. Inputs not yet in our hands

| Input | Needed for | Owner |
|---|---|---|
| `run-bootstrap-median.R` (Canvas) | "Run inherited code unchanged" (phase 1 gate) | Luis |
| `listing-from-100-run-bootstrap-median.txt` (Canvas) | Expected output of the inherited code | Luis |
| Canvas page *Performance Comparisons of R, Python, and Go* | S6 says to check it; it may set the expected measuring method | Luis |

Week 7's baseline code never arrived (L-W7-05). This week the baseline is a required deliverable, so these
files are requested at intake.

## 3. Riskiest assumptions to test in phase 1 (recorded now, tested next)

1. **"Same seed gives the same random draws in R and Go."** Both can use the Mersenne Twister, but R
   scrambles the seed in `set.seed`, turns generator output into a uniform number its own way, and since R
   3.6.0 draws `sample` indices by rejection sampling. If draws differ, results are comparable
   **statistically, not bit for bit**, and the spec's comparison rule must say so.
2. **"Median means the same thing on both sides."** For an even number of observations R averages the two
   middle values; a Go library may not.
3. **"Go is faster and saves cloud cost."** The assignment's premise. R's `sample` and `median` run in
   compiled C; the gap may be smaller than assumed, and whole-process time includes R's start-up.
