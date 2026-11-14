#!/usr/bin/env bash
# build-macos-app.sh: BootstrapSE.app, a double-clickable macOS application (adapted from
# Weeks 6 and 7). Output folder: dist/ (or the folder in DIST).
#
# The program is a command-line tool and a double-clicked app has no terminal, so the bundle
# is a thin wrapper: its launcher opens Terminal, runs the program on its built-in data set,
# shows how to analyse another file, and waits for Return so the result stays on screen.
# The program itself is unchanged.
set -euo pipefail
cd "$(dirname "$0")/.."
dist="${DIST:-dist}"
app="$dist/BootstrapSE.app"
rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -o "$app/Contents/Resources/bootse" .

cat >"$app/Contents/Resources/run-in-terminal.command" <<'SCRIPT'
#!/bin/bash
clear
program="$(dirname "$0")/bootse"
echo "Bootstrap standard errors of the median and the mean."
echo "Running on the built-in data set (1,000 lognormal observations), 1,000 resamples:"
echo
if "$program" -quiet; then
  echo
  echo "To analyse your own file (a header line, then one number per line), run in Terminal:"
  echo "  \"$program\" -data /path/to/your-file.csv"
  echo "and add -h to see every option."
else
  echo
  echo "The program reported a problem; see the message above."
fi
echo
read -r -p "Press Return to close this window. " _
SCRIPT

cat >"$app/Contents/MacOS/BootstrapSE" <<'SCRIPT'
#!/bin/bash
open -a Terminal "$(dirname "$0")/../Resources/run-in-terminal.command"
SCRIPT
chmod +x "$app/Contents/MacOS/BootstrapSE" "$app/Contents/Resources/run-in-terminal.command"

cat >"$app/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>               <string>BootstrapSE</string>
  <key>CFBundleDisplayName</key>        <string>Bootstrap SE</string>
  <key>CFBundleIdentifier</key>         <string>com.github.mioluigi.bootstrapse</string>
  <key>CFBundleExecutable</key>         <string>BootstrapSE</string>
  <key>CFBundlePackageType</key>        <string>APPL</string>
  <key>CFBundleShortVersionString</key> <string>1.0</string>
  <key>CFBundleVersion</key>            <string>1</string>
  <key>LSMinimumSystemVersion</key>     <string>13.0</string>
  <key>LSArchitecturePriority</key>     <array><string>arm64</string></array>
</dict>
</plist>
PLIST
echo "Built $app. Double-click it in Finder, or run: open $app"
