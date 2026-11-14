#!/usr/bin/env bash
# build.sh: executables for macOS (Apple silicon and Intel) and Windows.
#
# The program is pure Go with no dependencies outside the standard library, so it
# cross-compiles from any machine. Output folder: dist/ (or the folder in DIST).
# The macOS .app is built by scripts/build-macos-app.sh.
set -euo pipefail
cd "$(dirname "$0")/.."
dist="${DIST:-dist}"
mkdir -p "$dist"
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -trimpath -o "$dist/bootse-macos-arm64" .
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -trimpath -o "$dist/bootse-macos-amd64" .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o "$dist/bootse.exe" .
ls -l "$dist"
