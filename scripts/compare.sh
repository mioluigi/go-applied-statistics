#!/usr/bin/env bash
# compare.sh: the correctness gate (spec FR-7, §6). Needs R with the boot package.
#
# Runs five implementations on all 12 data sets with the same B and seed:
#   r-loop        R, sample() loop            (the reference)
#   r-boot        R, boot package
#   go-r          Go, R-compatible stream
#   go-block-1    Go, block stream, 1 worker
#   go-block-all  Go, block stream, all CPUs
# and checks them against each other by the rule in spec §6. Records go to results/comparison/,
# every check to results/comparison.tsv. Exits 1 if any check fails.
#
# Environment: REPLICATES (default 1000), SEED (default 431).
set -euo pipefail
cd "$(dirname "$0")/.."
replicates="${REPLICATES:-1000}"
seed="${SEED:-431}"
fail() { echo "compare: $*" >&2; exit 1; }
command -v Rscript >/dev/null 2>&1 || fail "Rscript not found; the comparison needs R."
scripts/check-inputs.sh >/dev/null || fail "the data sets fail the data gate (run scripts/check-inputs.sh)."

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/r" "$work/go" results/comparison
go build -o "$work/bootse" .
go build -o "$work/compare" ./cmd/compare
cpus="$(getconf _NPROCESSORS_ONLN)"

out=results/comparison
r_version="$(Rscript -e 'cat(R.version.string)')"
for name in r-loop r-boot; do
  printf '# %s; B = %s; seed = %s\n' "$r_version" "$replicates" "$seed" > "$out/$name.tsv"
done
for name in go-r go-block-1 go-block-all; do
  printf '# %s; B = %s; seed = %s\n' "$(go version | awk '{print $3}')" "$replicates" "$seed" > "$out/$name.tsv"
done

# Data set names come from the manifest, so the order is the same on every run.
datasets="$(awk -F'\t' '!/^#/ && $1 != "file" {print $1}' data/MANIFEST.tsv)"
for file in $datasets; do
  echo "compare: running the five implementations on $file"
  Rscript r/bootstrap-se.R "data/$file" "$replicates" "$seed" "$work/r/$file.medians" 2>/dev/null >> "$out/r-loop.tsv"
  Rscript r/bootstrap-se-boot.R "data/$file" "$replicates" "$seed" 2>/dev/null >> "$out/r-boot.tsv"
  "$work/bootse" -quiet -format tsv -data "data/$file" -replicates "$replicates" -seed "$seed" \
    -medians "$work/go/$file.medians" >> "$out/go-r.tsv"
  "$work/bootse" -quiet -format tsv -data "data/$file" -replicates "$replicates" -seed "$seed" \
    -stream block -workers 1 >> "$out/go-block-1.tsv"
  "$work/bootse" -quiet -format tsv -data "data/$file" -replicates "$replicates" -seed "$seed" \
    -stream block -workers "$cpus" >> "$out/go-block-all.tsv"
done

table=results/comparison.tsv
printf 'comparison\tdataset\tquantity\treference\tcandidate\tdifference\tlimit\tresult\n' > "$table"
failed=0
check() { "$work/compare" "$@" >> "$table" || failed=$((failed + 1)); }
check -label "Go (R stream) against R (sample loop): same draws" -rule same-draws \
  -reference "$out/r-loop.tsv" -candidate "$out/go-r.tsv" \
  -reference-medians "$work/r" -candidate-medians "$work/go"
check -label "Go (block stream) against R (sample loop): different draws" -rule different-draws \
  -reference "$out/r-loop.tsv" -candidate "$out/go-block-1.tsv"
check -label "Go (block stream), $cpus workers against 1 worker" -rule same-program \
  -reference "$out/go-block-1.tsv" -candidate "$out/go-block-all.tsv"
check -label "R (boot package) against R (sample loop): different draws" -rule different-draws \
  -reference "$out/r-loop.tsv" -candidate "$out/r-boot.tsv"
check -label "Go (R stream) against theory" -formula -candidate "$out/go-r.tsv" -data data

total=$(( $(wc -l < "$table" | tr -d ' ') - 1 ))
passed=$(grep -c 'PASS$' "$table" || true)
echo "compare: $passed of $total checks passed; details in $table"
[ "$failed" -eq 0 ] && [ "$passed" -eq "$total" ] || fail "the comparison FAILED; nothing may be timed until it passes."
