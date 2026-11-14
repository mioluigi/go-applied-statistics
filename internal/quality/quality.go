// Package quality holds checks on the source code itself: rules the rubric asks for that
// gofmt and go vet do not enforce.
package quality

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

// MaxFunctionLines is the longest a function may be, from its first to its last line.
const MaxFunctionLines = 60

// Check parses every non-test Go file under root and returns one message per violation:
// an exported identifier without a doc comment, or a function longer than MaxFunctionLines.
// Folders named testdata, and folders with their own go.mod below root, are skipped.
func Check(root string) ([]string, error) {
	return check(root, true)
}

// CheckIncludingTestdata is Check without the testdata exclusion below root; the check's own
// test uses it on a sample that must fail.
func CheckIncludingTestdata(root string) ([]string, error) {
	return check(root, false)
}

// check walks root and collects violations, optionally skipping folders outside the product.
func check(root string, skipFolders bool) ([]string, error) {
	var violations []string
	fileSet := token.NewFileSet()
	walk := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skipFolders && skippedFolder(root, path, entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		violations = append(violations, fileViolations(fileSet, file)...)
		return nil
	}
	if err := filepath.WalkDir(root, walk); err != nil {
		return nil, err
	}
	return violations, nil
}

// skippedFolder reports whether a folder is outside the code this check governs.
func skippedFolder(root, path, name string) bool {
	if path == root {
		return false
	}
	if name == "testdata" || strings.HasPrefix(name, ".") {
		return true
	}
	nestedModule, _ := filepath.Glob(filepath.Join(path, "go.mod"))
	return len(nestedModule) > 0
}

// fileViolations returns the violations in one parsed file.
func fileViolations(fileSet *token.FileSet, file *ast.File) []string {
	var violations []string
	report := func(position token.Pos, format string, arguments ...any) {
		violations = append(violations,
			fmt.Sprintf("%s: %s", fileSet.Position(position), fmt.Sprintf(format, arguments...)))
	}
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if declaration.Name.IsExported() && declaration.Doc == nil {
				report(declaration.Pos(), "exported function %s has no doc comment", declaration.Name.Name)
			}
			lines := fileSet.Position(declaration.End()).Line - fileSet.Position(declaration.Pos()).Line + 1
			if lines > MaxFunctionLines {
				report(declaration.Pos(), "function %s is %d lines long (limit %d)",
					declaration.Name.Name, lines, MaxFunctionLines)
			}
		case *ast.GenDecl:
			for _, name := range undocumentedNames(declaration) {
				report(declaration.Pos(), "exported %s has no doc comment", name)
			}
		}
	}
	return violations
}

// undocumentedNames lists the exported types, constants and variables of a declaration that
// have neither their own doc comment nor one on the enclosing declaration.
func undocumentedNames(declaration *ast.GenDecl) []string {
	if declaration.Doc != nil {
		return nil
	}
	var names []string
	for _, spec := range declaration.Specs {
		switch spec := spec.(type) {
		case *ast.TypeSpec:
			if spec.Name.IsExported() && spec.Doc == nil {
				names = append(names, spec.Name.Name)
			}
		case *ast.ValueSpec:
			for _, name := range spec.Names {
				if name.IsExported() && spec.Doc == nil {
					names = append(names, name.Name)
				}
			}
		}
	}
	return names
}
