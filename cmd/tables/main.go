// Command tables generates every number the README shows, from the saved measurements in
// results/ and the saved prices in prices/ (constitution 1.3).
//
// Usage, from the repository root:
//
//	go run ./cmd/tables            write results/tables.md and refresh the README's blocks
//	go run ./cmd/tables -check     exit 1 if the README's blocks differ from the generated ones
//
// The README holds one pair of markers per block:
//
//	<!-- generated:NAME:start -->
//	<!-- generated:NAME:end -->
//
// and everything between a pair is replaced by the block of that name.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mioluigi/go-applied-statistics/internal/report"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run generates the blocks and writes or checks the files.
func run(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("tables", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	check := flags.Bool("check", false, "do not write; fail if the README differs from the generated blocks")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	inputs, err := report.Load(*root)
	if err != nil {
		fmt.Fprintln(stderr, "tables:", err)
		return 1
	}
	blocks, err := inputs.Blocks()
	if err != nil {
		fmt.Fprintln(stderr, "tables:", err)
		return 1
	}
	readmePath := filepath.Join(*root, "README.md")
	current, err := os.ReadFile(readmePath)
	if err != nil {
		fmt.Fprintln(stderr, "tables:", err)
		return 1
	}
	updated, err := report.InsertBlocks(string(current), blocks)
	if err != nil {
		fmt.Fprintln(stderr, "tables:", err)
		return 1
	}
	if *check {
		if updated != string(current) {
			fmt.Fprintln(stderr, "tables: README.md differs from the generated blocks; run scripts/readme.sh")
			return 1
		}
		fmt.Fprintf(stdout, "tables: README.md matches the %d generated blocks\n", len(blocks))
		return 0
	}
	var tables strings.Builder
	tables.WriteString("# Generated tables\n\nWritten by `go run ./cmd/tables` from `results/` and `prices/`. Do not edit.\n")
	for _, block := range blocks {
		tables.WriteString("\n## " + block.Name + "\n\n" + block.Body)
	}
	if err := os.WriteFile(filepath.Join(*root, "results", "tables.md"), []byte(tables.String()), 0o644); err != nil {
		fmt.Fprintln(stderr, "tables:", err)
		return 1
	}
	if err := os.WriteFile(readmePath, []byte(updated), 0o644); err != nil {
		fmt.Fprintln(stderr, "tables:", err)
		return 1
	}
	fmt.Fprintf(stdout, "tables: wrote results/tables.md and refreshed %d blocks in README.md\n", len(blocks))
	return 0
}
