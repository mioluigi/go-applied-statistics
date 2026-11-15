#!/usr/bin/env bash
# acceptance.sh: checks every acceptance criterion of specs/001-bootstrap-median/spec.md (§8).
# Run it on a fresh clone before submitting. It changes nothing in the repository: builds and
# scratch files go to a temporary folder.
#
# Checks that need R run only where Rscript is installed; elsewhere they are reported as SKIP,
# never silently passed. Exit status: 0 if no check failed, 1 otherwise.
set -uo pipefail
cd "$(dirname "$0")/.."
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
passed=0; failed=0; skipped=0

# check NAME COMMAND...: runs the command quietly; PASS if it succeeds.
check() {
  local name="$1"; shift
  if "$@" > "$work/output.txt" 2>&1; then
    echo "PASS  $name"; passed=$((passed + 1))
  else
    echo "FAIL  $name"; sed 's/^/        /' "$work/output.txt" | tail -n 8; failed=$((failed + 1))
  fi
}
skip() { echo "SKIP  $1 ($2)"; skipped=$((skipped + 1)); }

# exits CODE COMMAND...: succeeds if the command exits with exactly that code.
exits() { local want="$1"; shift; "$@" > /dev/null 2>&1; [ "$?" -eq "$want" ]; }
# says TEXT COMMAND...: succeeds if the command's combined output contains the text.
says() { local text="$1"; shift; "$@" > "$work/said.txt" 2>&1; grep -q -F -- "$text" "$work/said.txt"; }
# has FILE TEXT: succeeds if the file contains the text.
has() { grep -q -F -- "$2" "$1"; }

have_r=no
command -v Rscript >/dev/null 2>&1 && have_r=yes

echo "== Build and quality (AC-16)"
check "gofmt reports no file" test -z "$(gofmt -l . 2>&1)"
check "go vet" go vet ./...
check "go test -race ./... (needs neither R nor the internet)" go test -race -count=1 ./...
check "program builds" go build -o "$work/bootse" .
check "comparison tool builds" go build -o "$work/compare" ./cmd/compare
DIST="$work/dist" scripts/build.sh > /dev/null 2>&1
DIST="$work/dist" scripts/build-macos-app.sh > /dev/null 2>&1
check "build scripts make the macOS binary, the Windows .exe and the .app (FR-14)" \
  test -s "$work/dist/bootse-macos-arm64" -a -s "$work/dist/bootse.exe" -a -x "$work/dist/BootstrapSE.app/Contents/MacOS/BootstrapSE"

echo "== Data (AC-1)"
check "12 data sets match the manifest; each has at least 100 observations" scripts/check-inputs.sh

echo "== Same results as R (AC-2, AC-3, AC-4, AC-6)"
check "Go reproduces R's saved reference values (AC-4)" go test -count=1 -run 'MatchesR|MatchesSampleInt|MatchesRunif|ReproducesR|RecordEqualsR' ./rrand ./bootstrap .
check "official comparison: every check passed" test "$(grep -c 'FAIL$' results/comparison.tsv)" = "0" -a "$(grep -c 'PASS$' results/comparison.tsv)" -ge 264
check "official comparison covers 12 data sets x 1,000 resample medians, all identical (AC-3)" \
  test "$(grep -c 'resample medians differing (of 1000)	0	0	0	0	PASS$' results/comparison.tsv)" = "12"
check "official comparison includes block stream, boot and theory (AC-6)" \
  test "$(cut -f1 results/comparison.tsv | sort -u | grep -c -E 'different draws|workers against 1 worker|against theory')" = "4"
if [ "$have_r" = yes ]; then
  Rscript r/original/run-bootstrap-median.R > "$work/course.txt" 2>/dev/null
  # The listing must appear as one contiguous block inside the program's output (AC-2).
  # The course's listing file has Windows line endings (CR LF) and R prints LF only, so the
  # carriage return is removed before comparing; nothing else is.
  check "course R program, unchanged, reproduces its listing (AC-2)" \
    awk 'NR==FNR { sub(/\r$/, ""); listing[++count] = $0; next }
         { sub(/\r$/, ""); output[++lines] = $0 }
         END { for (start = 1; start + count - 1 <= lines; start++) { same = 1
                 for (i = 1; i <= count; i++) if (output[start + i - 1] != listing[i]) { same = 0; break }
                 if (same) exit 0 }
               exit 1 }' r/original/listing-from-100-run-bootstrap-median.txt "$work/course.txt"
  mkdir -p "$work/r" "$work/go"
  for file in lognormal-100.csv normal-1000.csv; do
    Rscript r/bootstrap-se.R "data/$file" 1000 431 "$work/r/$file.medians" 2>/dev/null >> "$work/r-loop.tsv"
    "$work/bootse" -quiet -format tsv -data "data/$file" -medians "$work/go/$file.medians" >> "$work/go-r.tsv"
  done
  check "live: R and Go agree on two data sets, including every resample median (AC-3)" \
    "$work/compare" -label live -rule same-draws -reference "$work/r-loop.tsv" -candidate "$work/go-r.tsv" \
    -reference-medians "$work/r" -candidate-medians "$work/go"
else
  skip "course R program reproduces its listing (AC-2)" "R is not installed here"
  skip "live comparison of R and Go (AC-3)" "R is not installed here; the saved official comparison was checked above"
fi

echo "== Parallel results do not depend on the worker count (AC-5)"
check "block stream identical for 1, 2, 4, 7, 8 workers, 10 times, under the race detector" \
  go test -race -count=10 -run 'SameResultForAnyWorkerCount' ./bootstrap

echo "== Statistics and reuse (AC-7, AC-15)"
check "median, mean, standard deviation against independent answers (AC-7)" go test -count=1 ./stats
check "an outside statistic and an outside random source work with the engine (AC-15)" \
  go test -count=1 -run 'FromOutsideThePackage' ./bootstrap

echo "== The application (AC-8, AC-9, AC-10)"
printf 'value\n1.5\noops\n2.5\n' > "$work/text.csv"
printf 'value\n1.5\n' > "$work/one.csv"
check "no arguments: exit 0" exits 0 "$work/bootse"
for field in "Data set:" "n = 1000" "B = 1000" "seed 431" "stream r" "Sample median:" "bootstrap SE of the median:" "Sample mean:" "bootstrap SE of the mean:"; do
  check "no arguments: output names \"$field\" (AC-9)" says "$field" "$work/bootse" -quiet
done
check "missing file: exit 2" exits 2 "$work/bootse" -data "$work/absent.csv"
check "non-numeric value: exit 2" exits 2 "$work/bootse" -data "$work/text.csv"
check "non-numeric value: message names file and line" says 'text.csv, line 3: "oops" is not a number' "$work/bootse" -data "$work/text.csv"
check "one observation: exit 2" exits 2 "$work/bootse" -data "$work/one.csv"
check "one resample: exit 2" exits 2 "$work/bootse" -replicates 1
check "R stream with 4 workers: exit 2" exits 2 "$work/bootse" -workers 4
check "unknown stream: exit 2" exits 2 "$work/bootse" -stream fast
check "unknown flag: exit 2" exits 2 "$work/bootse" -speed 9
check "result file that cannot be written: exit 1" exits 1 "$work/bootse" -replicates 10 -medians "$work/no-such-folder/m.txt"
check "profiles are written on request (AC-10)" \
  sh -c "'$work/bootse' -quiet -replicates 50 -cpuprofile '$work/cpu.prof' -memprofile '$work/mem.prof' > /dev/null && test -s '$work/cpu.prof' -a -s '$work/mem.prof'"
check "log names read_ms and compute_ms (FR-11)" sh -c "'$work/bootse' 2>&1 >/dev/null | grep -q 'read_ms=' && '$work/bootse' 2>&1 >/dev/null | grep -q 'compute_ms='"

echo "== Saved measurements (AC-10, AC-11)"
check "committed profiles exist" test -s results/cpu.prof -a -s results/mem.prof -a -s results/cpu-profile.txt -a -s results/mem-profile.txt
for name in 'BenchmarkFloat64' 'BenchmarkStep1Median/before=sort' 'BenchmarkStep1Median/after=select' 'BenchmarkStep2Resample/before=allocate' \
            'BenchmarkStep2Resample/after=reuse' 'BenchmarkStep3Index/before=logarithm' 'BenchmarkStep4Generator/before=uniforms' \
            'BenchmarkStep4Generator/after=bits' 'BenchmarkEngine/before' 'BenchmarkEngine/after' 'BenchmarkWorkers'; do
  check "benchmarks: $name has -benchmem rows" sh -c "grep -F '$name' results/benchmarks.txt | grep -q 'allocs/op'"
done
# runs GROUP IMPLEMENTATION N MINIMUM: succeeds if runs.tsv has at least MINIMUM such runs.
runs() { [ "$(awk -F'\t' -v g="$1" -v i="$2" -v n="$3" '$1 == g && $2 == i && $4 == n' results/runs.tsv | wc -l | tr -d ' ')" -ge "$4" ]; }
for implementation in r-empty r-boot-load go-help; do
  check "runs: start-up $implementation, at least 200" runs startup "$implementation" 0 200
done
for implementation in r-loop r-boot go-r go-block-1 go-block-multi; do
  for n in 100 1000 10000; do
    check "runs: $implementation at n = $n, at least 200" runs scaling "$implementation" "$n" 200
  done
  check "runs: $implementation at n = 100000, at least 30" runs scaling "$implementation" 100000 30
done
check "runs: three shapes at n = 10,000 (the two other shapes, 200 each)" runs shapes r-loop 10000 400
check "runs: course program, at least 30" runs course r-course-program 0 30
check "runs: every run has a time and both memory figures" \
  sh -c "awk -F'\t' 'NR > 1 && (\$6 == \"\" || \$9 == \"\" || \$10 == \"\")' results/runs.tsv | wc -l | grep -q '^ *0$'"
check "machine record has machine, versions, commit and load averages" \
  sh -c "for key in 'machine:' 'go:' 'r:' 'commit:' 'load average (1 minute) before:' 'load average (1 minute) after:'; do grep -q \"^\$key\" results/machine.txt || exit 1; done"
measured="$(sed -n 's/^commit: //p' results/machine.txt)"
check "what was measured is what is submitted (no change to program, R code or data since commit $measured)" \
  git diff --quiet "$measured" HEAD -- main.go rrand/rrand.go stats/stats.go bootstrap/bootstrap.go bootstrap/stream.go dataset/dataset.go record/record.go r data
check "no personal data in results (user name, home path, e-mail)" \
  sh -c "! grep -r -I -l -E '/Users/|/home/[a-z]|@[a-z0-9-]+\.(com|co|edu|org)' results prices data testdata"

echo "== Generated numbers and prices (AC-12, AC-13)"
check "README numbers equal the generated ones" scripts/readme.sh --check
check "every price has a source and a capture date" go test -count=1 -run 'ReadPrices' ./internal/report

echo "== README content (AC-14)"
for text in 'https://cran.r-project.org/package=boot' 'https://cran.r-project.org/package=bootstrap' 'pkg.go.dev' 'Awesome Go' \
            'Gelman' 'https://github.com/seehuhn/mt19937' '## Recommendation' '## The R side' '## The Go side' '## Build, run, test' \
            'Creating the executables' '## Results' '## Cloud cost' 'AWS EC2' '## Effort' 'testing, benchmarking, profiling and logging' \
            '## GenAI Tools' 'docs/ai-dialogs'; do
  check "README has \"$text\"" has README.md "$text"
done
check "README blocks are filled (headline has numbers)" sh -c "sed -n '/generated:headline:start/,/generated:headline:end/p' README.md | grep -q 'comparison checks passed'"

echo "== AI dialogs (AC-17)"
check "docs/ai-dialogs has an index and at least one dialog" \
  sh -c "test -s docs/ai-dialogs/README.md && [ \"\$(ls docs/ai-dialogs | grep -v -c '^README.md\$')\" -ge 1 ]"
check "every file named in the dialog index exists" \
  sh -c "test -s docs/ai-dialogs/README.md && grep -o '\`[^\`]*\.\(pdf\|txt\|md\)\`' docs/ai-dialogs/README.md | tr -d '\`' | while read -r file; do test -e \"docs/ai-dialogs/\$file\" || test -e \"\$file\" || exit 1; done"

echo
echo "acceptance: $passed passed, $failed failed, $skipped skipped"
[ "$failed" -eq 0 ]
