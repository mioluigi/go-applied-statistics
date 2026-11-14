# Tasks 001

Status: **DRAFT, 2026-11-14**, for owner approval. One task = one commit (`T07: …`). 🔴 failing test first, 🟢 then code.
**Where**: C = Claude's workspace · M = Luis's Mac.

| ID | Where | Task | Verify | Covers |
|---|---|---|---|---|
| T01 | C | Module, layout, `.gitignore`, `CLAUDE.md`; `internal/quality` test (doc comments, function length), itself tested on a failing sample | `go vet ./... && go test ./...` | NFR-8 |
| T02 | C | `rrand`: R-compatible generator; tests against `testdata/r-reference/` (moved from spike evidence with provenance); mutation check | reference tests pass; mutants caught | FR-5, FR-R3, AC-4 |
| T03 | C | `stats`: two-pass `Mean`, `StandardDeviation`, sort-based `Median`; known answers incl. badly conditioned data (answers from exact arithmetic); mutation check | tests pass; mutants caught | FR-R4, AC-7 |
| T04 | C | R scripts: `generate-data.R`, `bootstrap-se.R`, `bootstrap-se-boot.R`, start-up scripts; `scripts/check-inputs.sh` (tested on pass and fail cases) | script tests; R parts verified in T05 | FR-1, FR-2 |
| T05 | **M** | Generate the 12 data sets and manifest; smoke-run both R scripts on one data set; commit, push | data gate passes; R record printed | FR-1, AC-1 |
| T06 | C | `dataset`: CSV reader, validation with file and line, record writer; synthetic fixtures shaped like the real files | tests pass | FR-4, FR-8 |
| T07 | C | `bootstrap` engine, R stream, one path with workers; test reproduces R's saved bootstrap standard errors and resample medians | tests pass; mutants caught | FR-3, FR-R1, AC-4, AC-15 |
| T08 | C | Block stream and parallel workers; identical results for 1, 2, 4, 8 workers; second random source; goroutine count restored | `go test -race -count=10` | FR-6, FR-R2, AC-5 |
| T09 | C | `bootse` command: flags, validation, text and tsv output, `-medians`, logging, profiles, embedded default; exit codes tested on the built binary | `main_test.go` builds and runs the binary | FR-8–FR-11, AC-8, AC-9 |
| T10 | C | `cmd/compare` (§6 rule; tested on passing and failing records) and `scripts/compare.sh` | tests pass | FR-7 |
| T11 | C | Performance steps, each a before/after benchmark: buffer reuse; `MedianInPlace` (selection); bit count without logarithm; equality with the plain variants tested | benchmarks run; equality tests pass | FR-12, S4 |
| T12 | C | `scripts/measure.sh` (guards tested on fail cases), benchmark and profile capture | guard tests | NFR-1–NFR-6 |
| T13 | C | `prices/aws-ec2.tsv` (captured from the provider's page, with source and date; Luis confirms) and the cost model | cost tests with hand-computed answers | FR-13, AC-13 |
| T14 | C | `cmd/tables` and `scripts/readme.sh`: tables, headline, cost, effort table, between README markers; `--check` mode | tests on fixture results | NFR-7, FR-15, AC-12 |
| T15 | C | Build scripts: binary, `.exe`, `.app` | cross-build in the workspace | FR-14 |
| T16 | **M** | Close other programs; run `scripts/measure.sh`; commit, push | `results/` complete; comparison all pass | AC-3, AC-6, AC-10, AC-11 |
| T17 | C | README: all sections of AC-14; generated blocks inserted; every sentence checked against the tables; spike prediction outcomes (addendum) | `scripts/readme.sh --check` | S1, S2, S8–S11, AC-14 |
| T18 | C | `scripts/acceptance.sh` (each check tested on a failing case) | passes in the workspace where R is not needed | AC-1–AC-17 |
| T19 | **M** + C | R2 Gemini repository review; triage; fixes | every finding decided | SFEDD phase 9 |
| T20 | **M** | Export AI dialogs to `docs/ai-dialogs/` with index; acceptance on a fresh clone; submit `.git` URL | acceptance passes | AC-17, S12 |

## Acceptance criteria (ticked at phase 10)

- [ ] AC-1 · [ ] AC-2 · [ ] AC-3 · [ ] AC-4 · [ ] AC-5 · [ ] AC-6 · [ ] AC-7 · [ ] AC-8 · [ ] AC-9
- [ ] AC-10 · [ ] AC-11 · [ ] AC-12 · [ ] AC-13 · [ ] AC-14 · [ ] AC-15 · [ ] AC-16 · [ ] AC-17
