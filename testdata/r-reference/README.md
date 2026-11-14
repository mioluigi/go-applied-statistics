# Reference values written by R

Third-party facts: what R 4.4.2 produced on the official Mac (Apple M1 Max) on 2026-11-14. Go tests read
these files, so they need neither R nor the internet. Each file's first line records the R version, the
generator settings and the call that produced it.

| File | Produced by | Commit |
|---|---|---|
| `sample-indices.tsv` | `r/spike/parity-reference.R` via `scripts/spike-r-baseline.sh` | `bf5b5ba` |
| `bootstrap-se.tsv` | same | `bf5b5ba` |
| `runif.tsv` | `scripts/env-probe.sh` (values copied from `docs/spikes/env-mac.txt`) | `80060ad` |

Do not edit by hand. To regenerate, run the producing script on the Mac.
