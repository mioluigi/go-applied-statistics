# CLAUDE.md: project context

Modern Applied Statistics with Go (Week 8): implement one computer-intensive statistical method in R and in
Go, show the results are comparable on the same data, measure time and memory, and recommend when Go should
replace R, with a cloud-cost saving. Process: SFEDD v2.1.

## Read these first (in order)

1. `docs/assignment-and-rubric.md`: steps S1–S12, rubric tags, who runs what
2. `docs/constitution.md`: non-negotiable rules
3. `docs/spikes/001-environment-and-baseline.md`: what each machine can do; riskiest assumptions
4. `specs/` (from phase 2): spec, plan, tasks; work the next unchecked task

## Working rules

- One task per commit; message starts with the task ID (`T05: …`); phase 0–4 documents use `P0: …` etc.
- Test first; mutation-check important tests (compile failures do not count).
- The workspace has **no R** and cannot reach `proxy.golang.org`: R runs on Luis's Mac, `go.sum` comes from
  the Mac, and workspace-only settings (`GOPROXY=direct`, `go.work`) are never committed.
- R's results enter Go tests as saved files with provenance; tests need neither R nor the internet.
- Never hand-type results or prices; generate them from `results/` and the saved price file.
- No personal data in files (no e-mail, user name, host name, home path).
- Changes after approval go in `docs/decisions.md` (newest first).
