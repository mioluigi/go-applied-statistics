#!/usr/bin/env bash
# measure.sh: the official measurements (SFEDD §5.3; spec NFR-1 to NFR-6), on Luis's Mac only.
#
# Close other programs first. The script then:
#   0. refuses a stale or modified copy, a failing data gate, or a machine without R and boot;
#   1. runs the correctness comparison (scripts/compare.sh); nothing is timed unless it passes;
#   2. times whole processes with /usr/bin/time -l, conditions interleaved:
#        start-up     empty R script, R loading boot, the Go program printing its help
#        scaling      5 implementations x 4 sizes, lognormal data (the parallel one uses
#                     8 workers, the vCPU count of the larger cloud instance in prices/)
#        shapes       normal and mirrored-lognormal data at n = 10,000
#        course       the course's R program, unchanged
#      one line per run in results/runs.tsv;
#   3. runs the Go benchmarks (-benchmem)            -> results/benchmarks.txt
#   4. profiles one n = 100,000 run                  -> results/cpu.prof, mem.prof and their text
#   5. records the machine, versions, commit and load -> results/machine.txt
#
# Environment: RUNS (default 200; runs per condition up to n = 10,000 and for start-up),
#              RUNS_LARGE (default 30; n = 100,000 and the course program),
#              TIME_TOOL (default "/usr/bin/time -l").
set -euo pipefail
cd "$(dirname "$0")/.."
runs="${RUNS:-200}"
runs_large="${RUNS_LARGE:-30}"
time_tool="${TIME_TOOL:-/usr/bin/time -l}"
replicates=1000
seed=431
busy_load=2.0
fail() { echo "measure: $*" >&2; exit 1; }

[ "$(uname -s)" = "Darwin" ] || fail "official measurements run on the Mac (this is $(uname -s))."
command -v Rscript >/dev/null 2>&1 || fail "Rscript not found; the measurement needs R."
Rscript -e 'library(boot)' >/dev/null 2>&1 || fail "the R package boot is not installed."
[ -z "$(git status --porcelain)" ] || fail "uncommitted changes; commit or stash them, then run again."
git fetch -q origin
[ "$(git rev-parse HEAD)" = "$(git rev-parse '@{u}')" ] || fail "this copy differs from GitHub; run git pull (or push), then run again."
scripts/check-inputs.sh >/dev/null || fail "the data sets fail the data gate (run scripts/check-inputs.sh)."

load_average() { sysctl -n vm.loadavg | awk '{print $2}'; }
load_before="$(load_average)"
if awk -v load="$load_before" -v limit="$busy_load" 'BEGIN { exit !(load > limit) }'; then
  echo "measure: WARNING: the 1-minute load average is $load_before (above $busy_load). Other programs are busy; timings will be noisier." >&2
fi

mkdir -p results
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
go build -o "$work/bootse" .
cpus="$(getconf _NPROCESSORS_ONLN)"
# The parallel condition uses as many workers as the larger priced instance has vCPUs (8), so
# its timing can be costed on that instance; fewer if this machine has fewer CPUs.
workers=8
[ "$cpus" -lt "$workers" ] && workers="$cpus"

echo "measure: step 1 of 5: correctness comparison"
scripts/compare.sh

runs_file=results/runs.tsv
printf 'group\timplementation\tdataset\tn\trun\treal_s\tuser_s\tsys_s\tmax_rss_bytes\tpeak_footprint_bytes\tread_ms\tcompute_ms\n' > "$runs_file"

# time_run GROUP IMPLEMENTATION DATASET N RUN COMMAND...: runs the command once under the time
# tool and appends one line to runs.tsv. The program's own log gives read_ms and compute_ms.
time_run() {
  local group="$1" implementation="$2" dataset="$3" n="$4" run="$5"; shift 5
  local log="$work/run.log"
  if ! $time_tool "$@" > /dev/null 2> "$log"; then
    tail -n 15 "$log" >&2
    fail "$implementation failed on $dataset (run $run)."
  fi
  local field
  field() { sed -n "s/.*$1=\([0-9.eE+-]*\).*/\1/p" "$log" | tail -n 1; }
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$group" "$implementation" "$dataset" "$n" "$run" \
    "$(awk '/ real /{print $1}' "$log")" "$(awk '/ real /{print $3}' "$log")" "$(awk '/ real /{print $5}' "$log")" \
    "$(awk '/maximum resident set size/{print $1}' "$log")" "$(awk '/peak memory footprint/{print $1}' "$log")" \
    "$(field read_ms)" "$(field compute_ms)" >> "$runs_file"
}

# run_implementation GROUP IMPLEMENTATION FILE N RUN: one timed run of one implementation.
run_implementation() {
  local group="$1" implementation="$2" file="$3" n="$4" run="$5"
  case "$implementation" in
    r-loop)       time_run "$group" "$implementation" "$file" "$n" "$run" Rscript r/bootstrap-se.R "data/$file" "$replicates" "$seed" ;;
    r-boot)       time_run "$group" "$implementation" "$file" "$n" "$run" Rscript r/bootstrap-se-boot.R "data/$file" "$replicates" "$seed" ;;
    go-r)         time_run "$group" "$implementation" "$file" "$n" "$run" "$work/bootse" -format tsv -data "data/$file" -replicates "$replicates" -seed "$seed" ;;
    go-block-1)   time_run "$group" "$implementation" "$file" "$n" "$run" "$work/bootse" -format tsv -data "data/$file" -replicates "$replicates" -seed "$seed" -stream block -workers 1 ;;
    go-block-multi) time_run "$group" "$implementation" "$file" "$n" "$run" "$work/bootse" -format tsv -data "data/$file" -replicates "$replicates" -seed "$seed" -stream block -workers "$workers" ;;
    *) fail "unknown implementation $implementation" ;;
  esac
}

echo "measure: step 2 of 5: whole-process timing ($runs runs per condition; $runs_large for the long ones)"
for run in $(seq 1 "$runs"); do
  time_run startup r-empty - 0 "$run" Rscript r/startup-empty.R
  time_run startup r-boot-load - 0 "$run" Rscript r/startup-boot.R
  time_run startup go-help - 0 "$run" "$work/bootse" -h
done
for n in 100 1000 10000 100000; do
  count="$runs"; [ "$n" = "100000" ] && count="$runs_large"
  echo "measure:   scaling, n = $n ($count runs of each implementation)"
  for run in $(seq 1 "$count"); do
    for implementation in r-loop r-boot go-r go-block-1 go-block-multi; do
      run_implementation scaling "$implementation" "lognormal-$n.csv" "$n" "$run"
    done
  done
done
echo "measure:   shapes at n = 10000"
for run in $(seq 1 "$runs"); do
  for shape in normal mirrored-lognormal; do
    for implementation in r-loop go-r go-block-multi; do
      run_implementation shapes "$implementation" "$shape-10000.csv" 10000 "$run"
    done
  done
done
echo "measure:   the course's R program, unchanged ($runs_large runs)"
for run in $(seq 1 "$runs_large"); do
  time_run course r-course-program - 0 "$run" Rscript r/original/run-bootstrap-median.R
done

echo "measure: step 3 of 5: Go benchmarks"
{
  echo "# go test -run '^\$' -bench . -benchmem -count 5 ./rrand ./stats ./bootstrap"
  go test -run '^$' -bench . -benchmem -count 5 ./rrand ./stats ./bootstrap
} > results/benchmarks.txt

echo "measure: step 4 of 5: profiles of one n = 100,000 run"
"$work/bootse" -quiet -format tsv -data data/lognormal-100000.csv -cpuprofile results/cpu.prof -memprofile results/mem.prof > /dev/null
{ echo "# go tool pprof -top: bootse -data data/lognormal-100000.csv (B = 1,000, R stream)"; go tool pprof -top -nodecount=15 "$work/bootse" results/cpu.prof 2>/dev/null; } > results/cpu-profile.txt
{ echo "# go tool pprof -sample_index=alloc_space -top: same run"; go tool pprof -sample_index=alloc_space -top -nodecount=10 "$work/bootse" results/mem.prof 2>/dev/null; } > results/mem-profile.txt

echo "measure: step 5 of 5: machine record"
{
  echo "machine: $(sysctl -n hw.model) · $(sysctl -n machdep.cpu.brand_string) · $(( $(sysctl -n hw.memsize) / 1073741824 )) GB · macOS $(sw_vers -productVersion)"
  echo "cpus: $cpus"
  echo "parallel workers: $workers"
  echo "go: $(go version | awk '{print $3}')"
  echo "r: $(Rscript -e 'cat(R.version.string)')"
  echo "r package boot: $(Rscript -e 'cat(as.character(packageVersion("boot")))')"
  echo "commit: $(git rev-parse --short HEAD)"
  echo "measured: $(date '+%Y-%m-%d %H:%M %Z')"
  echo "runs per condition: $runs (large conditions: $runs_large); B = $replicates; seed = $seed"
  echo "time tool: $time_tool"
  echo "load average (1 minute) before: $load_before"
  echo "load average (1 minute) after: $(load_average)"
} > results/machine.txt
echo "measure: done; $(( $(wc -l < "$runs_file" | tr -d ' ') - 1 )) timed runs in $runs_file"
