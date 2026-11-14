#!/usr/bin/env bash
# make-data.sh: task T05, on a machine with R (Luis's Mac).
#
#   1. Generates the 12 data sets (r/generate-data.R) and completes data/MANIFEST.tsv with
#      each file's size and SHA-256.
#   2. Runs the data gate.
#   3. First real run of the R analysis scripts, on one data set; saves the base-R record and
#      its 1,000 resample medians as reference values for the Go tests.
#   4. Writes R reference draws for boundary sizes (r/reference/edge-indices.R).
# It installs nothing and writes only under data/ and testdata/r-reference/.
set -euo pipefail
cd "$(dirname "$0")/.."
fail() { echo "make-data: $*" >&2; exit 1; }
command -v Rscript >/dev/null 2>&1 || fail "Rscript not found; this step needs R."

sha256_of() {
  if command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
  else sha256sum "$1" | awk '{print $1}'; fi
}

mkdir -p data testdata/r-reference
Rscript r/generate-data.R data

generation=data/MANIFEST-generation.tsv
[ -f "$generation" ] || fail "r/generate-data.R did not write $generation."
{
  head -n 1 "$generation"
  printf 'file\tshape\tn\tgenerator\tseed\tbytes\tsha256\n'
  tail -n +3 "$generation" | while IFS=$'\t' read -r file shape n generator seed; do
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$file" "$shape" "$n" "$generator" "$seed" \
      "$(wc -c < "data/$file" | tr -d ' ')" "$(sha256_of "data/$file")"
  done
} > data/MANIFEST.tsv
rm "$generation"
scripts/check-inputs.sh

reference=testdata/r-reference
provenance="# $(Rscript -e 'cat(R.version.string)'); Rscript r/bootstrap-se.R data/lognormal-1000.csv 1000 431"
record="$(Rscript r/bootstrap-se.R data/lognormal-1000.csv 1000 431 "$reference/lognormal-1000-medians.tmp" 2> "$reference/timing.tmp")"
{ echo "$provenance"; printf 'dataset\tn\treplicates\tseed\tstream\tsample_median\tse_median\tsample_mean\tse_mean\n'; echo "$record"; } \
  > "$reference/lognormal-1000-record.tsv"
{ echo "$provenance; the 1000 resample medians, in order"; echo "median"; cat "$reference/lognormal-1000-medians.tmp"; } \
  > "$reference/lognormal-1000-medians.tsv"
echo "base R : $record"
echo "         $(cat "$reference/timing.tmp")"
rm "$reference/lognormal-1000-medians.tmp" "$reference/timing.tmp"
echo "boot   : $(Rscript r/bootstrap-se-boot.R data/lognormal-1000.csv 1000 431)"

Rscript r/reference/edge-indices.R "$reference/sample-indices-edge.tsv"
Rscript r/startup-empty.R
Rscript r/startup-boot.R
echo "make-data: done"
