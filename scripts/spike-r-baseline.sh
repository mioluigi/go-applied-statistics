#!/usr/bin/env bash
# spike-r-baseline.sh: spike 001, Mac part (SFEDD phase 1).
#
#   1. Runs the course's R program UNCHANGED, three times, recording its output, whole-process
#      time and peak memory            -> docs/spikes/001-evidence/original-run-N.txt / -time-N.txt
#   2. Writes R reference values a Go program must reproduce
#                                       -> docs/spikes/001-evidence/sample-indices.tsv, bootstrap-se.tsv
# Installs nothing, changes nothing outside docs/spikes/001-evidence/.
set -euo pipefail
cd "$(dirname "$0")/.."
evidence=docs/spikes/001-evidence
fail() { echo "spike-r-baseline: $*" >&2; exit 1; }

[ "$(uname -s)" = "Darwin" ] || fail "this spike runs on the Mac (this is $(uname -s))."
command -v Rscript >/dev/null 2>&1 || fail "Rscript not found; R must be installed first."
[ -f r/original/run-bootstrap-median.R ] || fail "r/original/run-bootstrap-median.R is missing."

mkdir -p "$evidence"
for run in 1 2 3; do
  /usr/bin/time -l Rscript r/original/run-bootstrap-median.R \
    > "$evidence/original-run-$run.txt" 2> "$evidence/original-time-$run.txt" \
    || fail "the course's R program failed on run $run (see $evidence/original-time-$run.txt)."
  echo "run $run: $(awk '/ real /{print $1}' "$evidence/original-time-$run.txt") s, $(awk '/maximum resident set size/{print $1}' "$evidence/original-time-$run.txt") bytes peak"
done

Rscript r/spike/parity-reference.R "$evidence"
echo "spike-r-baseline: done; evidence in $evidence"
