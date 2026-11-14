#!/usr/bin/env bash
# env-probe.sh: SFEDD §5.7 environment probe, Mac side.
#
# Records what the official machine has (Go, R, R packages, timing tools, module proxy) in
# docs/spikes/env-mac.txt. It changes nothing outside that file and a temporary folder, installs
# nothing, and records no user name, host name or home path.
set -uo pipefail
cd "$(dirname "$0")/.."
out=docs/spikes/env-mac.txt
fail() { echo "env-probe: $*" >&2; exit 1; }
[ "$(uname -s)" = "Darwin" ] || fail "this probe is for the Mac (this is $(uname -s))."

# have NAME: prints the first line of "NAME --version", or MISSING.
have() {
  if command -v "$1" >/dev/null 2>&1; then "$@" 2>&1 | head -n 1; else echo "MISSING"; fi
}

# r_package NAME: prints the installed version of an R package, or MISSING.
r_package() {
  Rscript -e "cat(tryCatch(as.character(packageVersion('$1')), error = function(e) 'MISSING'))" 2>/dev/null \
    || echo "MISSING"
}

{
  echo "probe: $(date '+%Y-%m-%d %H:%M %Z')"
  echo "machine: $(sysctl -n hw.model) · $(sysctl -n machdep.cpu.brand_string) · $(( $(sysctl -n hw.memsize) / 1073741824 )) GB · macOS $(sw_vers -productVersion)"
  echo "cpus: $(sysctl -n hw.ncpu) (performance cores: $(sysctl -n hw.perflevel0.physicalcpu 2>/dev/null || echo n/a))"
  echo "go: $(have go version)"
  echo "git: $(have git --version)"
  echo "gh (GitHub CLI): $(have gh --version)"
  echo "R: $(have R --version)"
  echo "Rscript: $(command -v Rscript >/dev/null 2>&1 && echo present || echo MISSING)"
  if command -v Rscript >/dev/null 2>&1; then
    for package in boot bootstrap bench microbenchmark; do
      echo "R package $package: $(r_package "$package")"
    done
    echo "R RNGkind: $(Rscript -e 'cat(RNGkind(), sep = " / ")' 2>/dev/null)"
    echo "R set.seed(431); runif(3): $(Rscript -e 'set.seed(431); cat(format(runif(3), digits = 17))' 2>/dev/null)"
    echo "R set.seed(431); sample(10, 5, replace = TRUE): $(Rscript -e 'set.seed(431); cat(sample(10, 5, replace = TRUE))' 2>/dev/null)"
    echo "R median(c(1, 2, 3, 4)): $(Rscript -e 'cat(median(c(1, 2, 3, 4)))' 2>/dev/null)"
  fi
  echo "/usr/bin/time -l: $([ -x /usr/bin/time ] && echo present || echo MISSING)"

  scratch="$(mktemp -d)"
  if command -v go >/dev/null 2>&1 \
    && (cd "$scratch" && go mod init envprobe >/dev/null 2>&1 \
        && go get github.com/seehuhn/mt19937@v1.0.0 >/dev/null 2>&1); then
    echo "go module proxy (go get github.com/seehuhn/mt19937@v1.0.0): ok"
  else
    echo "go module proxy (go get github.com/seehuhn/mt19937@v1.0.0): FAILED"
  fi
  rm -rf "$scratch"
} > "$out"

cat "$out"
echo
echo "env-probe: wrote $out"
